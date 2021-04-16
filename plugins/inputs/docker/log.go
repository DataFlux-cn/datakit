package docker

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"

	iod "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

const (
	// Maximum bytes of a log line before it will be split, size is mirroring
	// docker code:
	// https://github.com/moby/moby/blob/master/daemon/logger/copier.go#L21
	maxLineBytes = 16 * 1024

	// ES value can be at most 32766 bytes long
	maxFieldsLength = 32766

	pipelineTimeField = "time"

	useIOHighFreq = true
)

type containerLog struct {
}

func (this *Input) addToContainerList(containerID string, cancel context.CancelFunc) error {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.containerLogList[containerID] = cancel
	return nil
}

func (this *Input) removeFromContainerList(containerID string) error {
	this.mu.Lock()
	defer this.mu.Unlock()
	delete(this.containerLogList, containerID)
	return nil
}

func (this *Input) containerInContainerList(containerID string) bool {
	this.mu.Lock()
	defer this.mu.Unlock()
	_, ok := this.containerLogList[containerID]
	return ok
}

func (this *Input) cancelTails() error {
	this.mu.Lock()
	defer this.mu.Unlock()
	for _, cancel := range this.containerLogList {
		cancel()
	}
	return nil
}

func (this *Input) gatherLog() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, this.timeoutDuration)
	defer cancel()

	cList, err := this.client.ContainerList(ctx, this.opts)
	if err != nil {
		l.Error(err)
		return
	}

	for _, container := range cList {
		if this.containerInContainerList(container.ID) {
			continue
		}

		ctx, cancel := context.WithCancel(context.Background())
		this.addToContainerList(container.ID, cancel)

		// Start a new goroutine for every new container that has logs to collect
		this.wg.Add(1)
		go func(container types.Container) {
			defer this.wg.Done()
			defer this.removeFromContainerList(container.ID)

			err = this.tailContainerLogs(ctx, container)
			if err != nil && err != context.Canceled {
				l.Error(err)
			}
		}(container)
	}
}

func (this *Input) tailContainerLogs(ctx context.Context, container types.Container) error {
	// ignore imageVersion
	imageName, _ := ParseImage(container.Image)
	containerName := getContainerName(container.Names)

	tags := map[string]string{
		"container_name": containerName,
		"image_name":     imageName,
	}
	for k, v := range this.Tags {
		tags[k] = v
	}

	hasTTY, err := this.hasTTY(ctx, container)
	if err != nil {
		return err
	}

	logReader, err := this.client.ContainerLogs(ctx, container.ID, this.containerLogsOptions)
	if err != nil {
		return err
	}

	// If the container is using a TTY, there is only a single stream
	// (stdout), and data is copied directly from the container output stream,
	// no extra multiplexing or headers.
	//
	// If the container is *not* using a TTY, streams for stdout and stderr are
	// multiplexed.

	for _, opt := range this.LogOption {
		if opt.nameCompile != nil && opt.nameCompile.MatchString(containerName) {
			if hasTTY {
				return tailStream(logReader, "tty", container, opt, tags)
			} else {
				return tailMultiplexed(logReader, container, opt, tags)
			}
		}
	}

	if hasTTY {
		return tailStream(logReader, "tty", container, nil, tags)
	} else {
		return tailMultiplexed(logReader, container, nil, tags)
	}

}

func (this *Input) hasTTY(ctx context.Context, container types.Container) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, this.timeoutDuration)
	defer cancel()
	c, err := this.client.ContainerInspect(ctx, container.ID)
	if err != nil {
		return false, err
	}
	return c.Config.Tty, nil
}

func tailStream(reader io.ReadCloser, stream string, container types.Container, opt *LogOption, baseTags map[string]string) error {
	defer reader.Close()

	tags := make(map[string]string, len(baseTags)+1)
	for k, v := range baseTags {
		tags[k] = v
	}
	tags["stream"] = stream

	r := bufio.NewReaderSize(reader, maxLineBytes)

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

		containerName := getContainerName(container.Names)
		// default measurement is containerName, if source not empty use $source.
		var measurement = containerName
		var fields = map[string]interface{}{
			"service":         containerName,
			"from_kubernetes": contianerIsFromKubernetes(containerName),
		}

		// l.Debugf("get %d bytes from source: %s", len(message), measurement)
		if opt != nil {
			if pipe := opt.pipelinePool.Get(); pipe != nil {
				fields, err = pipe.(*pipeline.Pipeline).Run(message).Result()
				if err != nil {
					l.Errorf("run pipeline error, %s", err)
				}
			} else {
				// 当 opt 存在但 pipeline 不存在时，执行默认操作
				// 与经过 pipeline 处理但是失败后，返回的 fields 相同，只有一个 message
				fields["message"] = message
			}
			if opt.Source != "" {
				measurement = opt.Source
			}
			if opt.Service != "" {
				fields["service"] = opt.Service
			}
		} else {
			fields["message"] = message
		}

		if err := checkFieldsLength(fields, maxFieldsLength); err != nil {
			// 只有在碰到非 message 字段，且长度超过最大限制时才会返回 error
			// 防止通过 pipeline 添加巨长字段的恶意行为
			l.Error(err)
			continue
		}

		fmt.Println(fields["message"])

		addStatus(fields)

		ts, err = takeTime(fields)
		if err != nil {
			ts = time.Now()
			l.Error(err)
		}

		pt, err := iod.MakePoint(measurement, tags, fields, ts)
		if err != nil {
			l.Error(err)
		} else {
			if err := iod.Feed(inputName, iod.Logging, []*iod.Point{pt}, &iod.Option{HighFreq: useIOHighFreq}); err != nil {
				l.Error(err)
			}
		}
	}
}

func tailMultiplexed(src io.ReadCloser, container types.Container, opt *LogOption, baseTags map[string]string) error {
	outReader, outWriter := io.Pipe()
	errReader, errWriter := io.Pipe()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := tailStream(outReader, "stdout", container, opt, baseTags)
		if err != nil {
			l.Error(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := tailStream(errReader, "stderr", container, opt, baseTags)
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

func takeTime(fields map[string]interface{}) (ts time.Time, err error) {
	// time should be nano-second
	if v, ok := fields[pipelineTimeField]; ok {
		nanots, ok := v.(int64)
		if !ok {
			err = fmt.Errorf("invalid filed `%s: %v', should be nano-second, but got `%s'",
				pipelineTimeField, v, reflect.TypeOf(v).String())
			return
		}

		ts = time.Unix(nanots/int64(time.Second), nanots%int64(time.Second))
		delete(fields, pipelineTimeField)
	} else {
		ts = time.Now()
	}

	return
}

// checkFieldsLength 指定字段长度 "小于等于" maxlength
func checkFieldsLength(fields map[string]interface{}, maxlength int) error {
	for k, v := range fields {
		switch vv := v.(type) {
		// FIXME:
		// need  "case []byte" ?
		case string:
			if len(vv) <= maxlength {
				continue
			}
			if k == "message" {
				fields[k] = vv[:maxlength]
			} else {
				return fmt.Errorf("fields: %s, length=%d, out of maximum length", k, len(vv))
			}
		default:
			// nil
		}
	}
	return nil
}

var statusMap = map[string]string{
	"f":        "emerg",
	"emerg":    "emerg",
	"a":        "alert",
	"alert":    "alert",
	"c":        "critical",
	"critical": "critical",
	"e":        "error",
	"error":    "error",
	"w":        "warning",
	"warning":  "warning",
	"i":        "info",
	"info":     "info",
	"d":        "debug",
	"trace":    "debug",
	"verbose":  "debug",
	"debug":    "debug",
	"o":        "OK",
	"s":        "OK",
	"ok":       "OK",
}

func addStatus(fields map[string]interface{}) {
	// map 有 "status" 字段
	statusField, ok := fields["status"]
	if !ok {
		fields["status"] = "info"
		return
	}
	// "status" 类型必须是 string
	statusStr, ok := statusField.(string)
	if !ok {
		fields["status"] = "info"
		return
	}

	// 查询 statusMap 枚举表并替换
	if v, ok := statusMap[strings.ToLower(statusStr)]; !ok {
		fields["status"] = "info"
	} else {
		fields["status"] = v
	}
}
