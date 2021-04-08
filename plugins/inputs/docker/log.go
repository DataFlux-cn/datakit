package docker

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/stdcopy"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	iod "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (

	// Maximum bytes of a log line before it will be split, size is mirroring
	// docker code:
	// https://github.com/moby/moby/blob/master/daemon/logger/copier.go#L21
	maxLineBytes = 16 * 1024

	updateInterval = 5 * time.Second
)

var (
	containerStates = []string{"created", "restarting", "running", "removing", "paused", "exited", "dead"}
)

type DockerLogs struct {
	Endpoint              string            `toml:"endpoint"`
	FromBeginning         bool              `toml:"from_beginning"`
	Timeout               string            `toml:"timeout"`
	LabelInclude          []string          `toml:"docker_label_include"`
	LabelExclude          []string          `toml:"docker_label_exclude"`
	ContainerInclude      []string          `toml:"container_name_include"`
	ContainerExclude      []string          `toml:"container_name_exclude"`
	ContainerStateInclude []string          `toml:"container_state_include"`
	ContainerStateExclude []string          `toml:"container_state_exclude"`
	Source                string            `toml:"source"`
	PipelinePath          string            `toml:"pipeline_path"`
	IncludeSourceTag      bool              `toml:"source_tag"`
	Tags                  map[string]string `toml:"tags"`

	timeoutDuration time.Duration
	ClientConfig

	newEnvClient func() (Client, error)
	newClient    func(string, *tls.Config) (Client, error)

	client          Client
	labelFilter     Filter
	containerFilter Filter
	stateFilter     Filter
	opts            types.ContainerListOptions
	wg              sync.WaitGroup
	mu              sync.Mutex
	containerList   map[string]context.CancelFunc

	pipe *pipeline.Pipeline
}

func (*DockerLogs) SampleConfig() string {
	return sampleCfg
}

func (*DockerLogs) Catalog() string {
	return "docker"
}

func (d *DockerLogs) Run() {
	l = logger.SLogger(inputName)

	if d.initCfg() {
		return
	}

	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	l.Info("dockerlog input start")
	for {
		select {
		case <-datakit.Exit.Wait():
			d.Stop()
			l.Info("exit")
			return

		case <-ticker.C:
			d.gather()
		}
	}
}

func (d *DockerLogs) gather() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, d.timeoutDuration)
	defer cancel()
	containers, err := d.client.ContainerList(ctx, d.opts)
	if err != nil {
		l.Error(err)
		return
	}

	for _, container := range containers {
		if d.containerInContainerList(container.ID) {
			continue
		}

		containerName := d.matchedContainerName(container.Names)
		if containerName == "" {
			continue
		}

		ctx, cancel := context.WithCancel(context.Background())
		d.addToContainerList(container.ID, cancel)

		// Start a new goroutine for every new container that has logs to collect
		d.wg.Add(1)
		go func(container types.Container) {
			defer d.wg.Done()
			defer d.removeFromContainerList(container.ID)

			err = d.tailContainerLogs(ctx, container, containerName)
			if err != nil && err != context.Canceled {
				l.Error(err)
			}
		}(container)
	}
}

func (d *DockerUtil) tailContainerLogs(ctx context.Context, container types.Container, containerName string) error {
	imageName, imageVersion := ParseImage(container.Image)
	tags := map[string]string{
		"container_name":    containerName,
		"container_image":   imageName,
		"container_version": imageVersion,
		"endpoint":          d.Endpoint,
	}

	for k, v := range d.Tags {
		tags[k] = v
	}

	measurement := containerName
	if d.Source != "" {
		measurement = d.Source
	}

	// Add matching container labels as tags
	for k, label := range container.Labels {
		if d.labelFilter.Match(k) {
			tags[k] = label
		}
	}

	hasTTY, err := d.hasTTY(ctx, container)
	if err != nil {
		return err
	}

	tail := "0"
	if d.FromBeginning {
		tail = "all"
	}

	logOptions := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Details:    false,
		Follow:     true,
		Tail:       tail,
	}

	logReader, err := d.client.ContainerLogs(ctx, container.ID, logOptions)
	if err != nil {
		return err
	}

	// If the container is using a TTY, there is only a single stream
	// (stdout), and data is copied directly from the container output stream,
	// no extra multiplexing or headers.
	//
	// If the container is *not* using a TTY, streams for stdout and stderr are
	// multiplexed.
	if hasTTY {
		return tailStream(measurement, tags, container.ID, logReader, "tty", d.pipe)
	} else {
		return tailMultiplexed(measurement, tags, container.ID, logReader, d.pipe)
	}
}

func parseLine(line []byte) (time.Time, string, error) {
	parts := bytes.SplitN(line, []byte(" "), 2)

	switch len(parts) {
	case 1:
		parts = append(parts, []byte(""))
	}

	tsString := string(parts[0])

	// Keep any leading space, but remove whitespace from end of line.
	// This preserves space in, for example, stacktraces, while removing
	// annoying end of line characters and is similar to how other logging
	// plugins such as syslog behave.
	message := bytes.TrimRightFunc(parts[1], unicode.IsSpace)

	ts, err := time.Parse(time.RFC3339Nano, tsString)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("error parsing timestamp %q: %v", tsString, err)
	}

	return ts, string(message), nil
}

func tailStream(measurement string, baseTags map[string]string, containerID string, reader io.ReadCloser, stream string, pipe *pipeline.Pipeline) error {
	defer reader.Close()

	tags := make(map[string]string, len(baseTags)+1)
	for k, v := range baseTags {
		tags[k] = v
	}
	tags["stream"] = stream

	r := bufio.NewReaderSize(reader, maxLineBytes)

	var pts []*iod.Point

	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if len(line) == 0 {
			continue
		}

		ts, message, err := parseLine(line)
		if err != nil {
			l.Error(err)
			continue
		}

		var fields = make(map[string]interface{})

		if pipe != nil {
			fields, err = pipe.Run(message).Result()
			if err != nil {
				l.Errorf("run pipeline error, %s", err)
				continue
			}
		} else {
			fields["message"] = message
		}

		if _, ok := fields["container_id"]; !ok {
			fields["container_id"] = containerID
		}

		if v, ok := fields["time"]; ok { // time should be nano-second
			nanots, ok := v.(int64)
			if !ok {
				l.Warn("filed `time' should be nano-second, but got `%s'", reflect.TypeOf(v).String())
				continue
			}

			ts = time.Unix(nanots/int64(time.Second), nanots%int64(time.Second))
			delete(fields, "time")
		}

		pt, err := iod.MakePoint(measurement, tags, fields, ts)
		if err != nil {
			l.Error(err)
		} else {
			pts = append(pts, pt)
		}
	}

	if err := iod.Feed(inputName, iod.Logging, pts, &iod.Option{HighFreq: true}); err != nil {
		l.Error(err)
		return err
	}
	return nil
}

func tailMultiplexed(measurement string, tags map[string]string, containerID string, src io.ReadCloser, pipe *pipeline.Pipeline) error {
	outReader, outWriter := io.Pipe()
	errReader, errWriter := io.Pipe()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := tailStream(measurement, tags, containerID, outReader, "stdout", pipe)
		if err != nil {
			l.Error(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := tailStream(measurement, tags, containerID, errReader, "stderr", pipe)
		if err != nil {
			l.Error(err)
		}
	}()

	_, err := stdcopy.StdCopy(outWriter, errWriter, src)
	outWriter.Close()
	errWriter.Close()
	src.Close()
	wg.Wait()
	return err
}

func (d *DockerLogs) matchedContainerName(names []string) string {
	// Check if all container names are filtered; in practice I believe
	// this array is always of length 1.
	for _, name := range names {
		trimmedName := strings.TrimPrefix(name, "/")
		match := d.containerFilter.Match(trimmedName)
		if match {
			return trimmedName
		}
	}
	return ""
}

func (d *DockerLogs) hasTTY(ctx context.Context, container types.Container) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, d.timeoutDuration)
	defer cancel()
	c, err := d.client.ContainerInspect(ctx, container.ID)
	if err != nil {
		return false, err
	}
	return c.Config.Tty, nil
}

// Adapts some of the logic from the actual Docker library's image parsing
// routines:
// https://github.com/docker/distribution/blob/release/2.7/reference/normalize.go
func ParseImage(image string) (string, string) {
	domain := ""
	remainder := ""

	i := strings.IndexRune(image, '/')

	if i == -1 || (!strings.ContainsAny(image[:i], ".:") && image[:i] != "localhost") {
		remainder = image
	} else {
		domain, remainder = image[:i], image[i+1:]
	}

	imageName := ""
	imageVersion := "unknown"

	i = strings.LastIndex(remainder, ":")
	if i > -1 {
		imageVersion = remainder[i+1:]
		imageName = remainder[:i]
	} else {
		imageName = remainder
	}

	if domain != "" {
		imageName = domain + "/" + imageName
	}

	return imageName, imageVersion
}
