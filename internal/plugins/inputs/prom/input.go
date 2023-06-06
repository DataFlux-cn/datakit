// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package prom scrape prometheus exporter metrics.
package prom

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	iprom "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
)

var _ inputs.ElectionInput = (*Input)(nil)

const (
	inputName               = "prom"
	catalog                 = "prom"
	defaultIntervalDuration = time.Second * 30
)

// defaultMaxFileSize is the default max response body size, in bytes.
// This field is used only when metrics are written to file, i.e. Output is configured.
// If the size of response body is over defaultMaxFileSize, metrics will be discarded.
// 32 MB.
const defaultMaxFileSize int64 = 32 * 1024 * 1024

var l = logger.DefaultSLogger(inputName)

type Input struct {
	Source   string           `toml:"source" json:"source"`
	Interval string           `toml:"interval" json:"interval"`
	Timeout  datakit.Duration `toml:"timeout" json:"timeout"`

	URL                    string       `toml:"url,omitempty"` // Deprecated
	URLs                   []string     `toml:"urls" json:"urls"`
	IgnoreReqErr           bool         `toml:"ignore_req_err" json:"ignore_req_err"`
	MetricTypes            []string     `toml:"metric_types" json:"metric_types"`
	MetricNameFilter       []string     `toml:"metric_name_filter" json:"metric_name_filter"`
	MetricNameFilterIgnore []string     `toml:"metric_name_filter_ignore" json:"metric_name_filter_ignore"`
	MeasurementPrefix      string       `toml:"measurement_prefix" json:"measurement_prefix"`
	MeasurementName        string       `toml:"measurement_name" json:"measurement_name"`
	Measurements           []iprom.Rule `toml:"measurements" json:"measurements"`
	Output                 string       `toml:"output" json:"output"`
	MaxFileSize            int64        `toml:"max_file_size" json:"max_file_size"`

	TLSOpen    bool   `toml:"tls_open" json:"tls_open"`
	UDSPath    string `toml:"uds_path" json:"uds_path"`
	CacertFile string `toml:"tls_ca" json:"tls_ca"`
	CertFile   string `toml:"tls_cert" json:"tls_cert"`
	KeyFile    string `toml:"tls_key" json:"tls_key"`

	TagsIgnore  []string            `toml:"tags_ignore" json:"tags_ignore"`
	TagsRename  *iprom.RenameTags   `toml:"tags_rename" json:"tags_rename"`
	AsLogging   *iprom.AsLogging    `toml:"as_logging" json:"as_logging"`
	IgnoreTagKV map[string][]string `toml:"ignore_tag_kv_match" json:"ignore_tag_kv_match"`
	HTTPHeaders map[string]string   `toml:"http_headers" json:"http_headers"`

	Tags               map[string]string `toml:"tags" json:"tags"`
	DisableHostTag     bool              `toml:"disable_host_tag" json:"disable_host_tag"`
	DisableInstanceTag bool              `toml:"disable_instance_tag" json:"disable_instance_tag"`
	DisableInfoTag     bool              `toml:"disable_info_tag" json:"disable_info_tag"`

	Auth map[string]string `toml:"auth" json:"auth"`

	pm     *iprom.Prom
	Feeder io.Feeder

	Election bool `toml:"election" json:"election"`
	chPause  chan bool
	pause    bool

	urls []*url.URL

	semStop *cliutils.Sem // start stop signal

	isInitialized bool

	// Input holds logger because prom have different types of instances.
	l *logger.Logger
}

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) SampleMeasurement() []inputs.Measurement { return nil }

func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (*Input) Catalog() string { return catalog }

func (i *Input) SetTags(m map[string]string) {
	if i.Tags == nil {
		i.Tags = make(map[string]string)
	}
	for k, v := range m {
		if _, ok := i.Tags[k]; !ok {
			i.Tags[k] = v
		}
	}
}

func (i *Input) ElectionEnabled() bool {
	return i.Election
}

func (i *Input) Run() {
	if i.setup() {
		return
	}

	tick := time.NewTicker(i.pm.Option().GetIntervalDuration())
	defer tick.Stop()

	i.l.Info("prom start")

	for {
		if i.pause {
			i.l.Debug("prom paused")
		} else {
			if err := i.RunningCollect(); err != nil {
				i.l.Warn(err)
			}
		}

		select {
		case <-datakit.Exit.Wait():
			i.l.Info("prom exit")
			return

		case <-i.semStop.Wait():
			i.l.Info("prom return")
			return

		case <-tick.C:

		case i.pause = <-i.chPause:
			// nil
		}
	}
}

func (i *Input) GetIntervalDuration() time.Duration {
	if !i.isInitialized {
		if err := i.Init(); err != nil {
			i.l.Infof("prom init error: %s", err)
			return defaultIntervalDuration
		}
	}
	return i.pm.Option().GetIntervalDuration()
}

func (i *Input) RunningCollect() error {
	if !i.isInitialized {
		if err := i.Init(); err != nil {
			return err
		}
	}

	ioname := inputName + "/" + i.Source

	start := time.Now()
	pts, err := i.doCollect()
	if err != nil {
		return err
	}
	if pts == nil {
		return fmt.Errorf("points got nil from doCollect")
	}

	if i.AsLogging != nil && i.AsLogging.Enable {
		// Feed measurement as logging.
		for _, pt := range pts {
			// We need to feed each point separately because
			// each point might have different measurement name.
			if err := i.Feeder.Feed(string(pt.Name()), point.Logging, []*point.Point{pt},
				&io.Option{CollectCost: time.Since(start)}); err != nil {
				i.Feeder.FeedLastError(ioname, err.Error())
			}
		}
	} else {
		err := i.Feeder.Feed(ioname, point.Metric, pts,
			&io.Option{CollectCost: time.Since(start)})
		if err != nil {
			i.l.Errorf("Feed: %s", err)
			i.Feeder.FeedLastError(ioname, err.Error())
		}
	}
	return nil
}

func (i *Input) doCollect() ([]*point.Point, error) {
	i.l.Debugf("collect URLs %v", i.URLs)

	// If Output is configured, data is written to local file specified by Output.
	// Data will no more be written to datakit io.
	if i.Output != "" {
		err := i.WriteMetricText2File()
		if err != nil {
			i.l.Errorf("WriteMetricText2File: %s", err.Error())
		}
		return nil, nil
	}

	pts, err := i.Collect()
	if err != nil {
		i.l.Errorf("Collect: %s", err)
		i.Feeder.FeedLastError(i.Source, err.Error())

		// Try testing the connect
		for _, u := range i.urls {
			if err := net.RawConnect(u.Hostname(), u.Port(), time.Second*3); err != nil {
				i.l.Errorf("failed to connect to %s:%s, %s", u.Hostname(), u.Port(), err)
			}
		}

		return nil, err
	}

	if pts == nil {
		return nil, fmt.Errorf("points got nil from Collect")
	}

	// Processing election information.
	var opts map[string]string
	if i.Election {
		opts = dkpt.GlobalElectionTags()
	} else {
		opts = dkpt.GlobalHostTags()
	}

	for j := 0; j < len(pts); j++ {
		for k, v := range opts {
			pts[j].AddTag([]byte(k), []byte(v))
		}
	}

	return pts, nil
}

// Collect collects metrics from all URLs.
func (i *Input) Collect() ([]*point.Point, error) {
	if i.pm == nil {
		return nil, fmt.Errorf("i.pm is nil")
	}
	var points []*point.Point
	for _, u := range i.URLs {
		uu, err := url.Parse(u)
		if err != nil {
			return nil, err
		}
		var pts []*point.Point
		if uu.Scheme != "http" && uu.Scheme != "https" {
			pts, err = i.CollectFromFile(u)
		} else {
			pts, err = i.CollectFromHTTP(u)
		}
		if err != nil {
			return nil, err
		}
		points = append(points, pts...)
	}

	return points, nil
}

func (i *Input) CollectFromHTTP(u string) ([]*point.Point, error) {
	if i.pm == nil {
		return nil, nil
	}
	return i.pm.CollectFromHTTPV2(u)
}

func (i *Input) CollectFromFile(filepath string) ([]*point.Point, error) {
	if i.pm == nil {
		return nil, nil
	}
	return i.pm.CollectFromFileV2(filepath)
}

// WriteMetricText2File collects from all URLs and then
// directly writes them to file specified by field Output.
func (i *Input) WriteMetricText2File() error {
	// Remove if file already exists.
	if _, err := os.Stat(i.Output); err == nil {
		if err := os.Remove(i.Output); err != nil {
			return err
		}
	}
	for _, u := range i.URLs {
		if err := i.pm.WriteMetricText2File(u); err != nil {
			return err
		}
		stat, err := os.Stat(i.Output)
		if err != nil {
			return err
		}
		if stat.Size() > i.MaxFileSize {
			return fmt.Errorf("file size is too large, max: %d, got: %d", i.MaxFileSize, stat.Size())
		}
	}
	return nil
}

func (i *Input) Terminate() {
	if i.semStop != nil {
		i.semStop.Close()
	}
}

func (i *Input) setup() bool {
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}
		time.Sleep(1 * time.Second) // sleep a while
		if err := i.Init(); err != nil {
			continue
		} else {
			break
		}
	}

	return false
}

func (i *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	select {
	case i.chPause <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (i *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	select {
	case i.chPause <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func (i *Input) Init() error {
	i.l = logger.SLogger(inputName + "/" + i.Source)

	if i.URL != "" {
		i.URLs = append(i.URLs, i.URL)
	}
	for _, u := range i.URLs {
		uu, err := url.Parse(u)
		if err != nil {
			return err
		}
		i.urls = append(i.urls, uu)
	}

	opts := []iprom.PromOption{
		iprom.WithLogger(i.l), // WithLogger must in the first
		iprom.WithSource(i.Source),
		iprom.WithInterval(i.Interval),
		iprom.WithTimeout(i.Timeout),
		iprom.WithIgnoreReqErr(i.IgnoreReqErr),
		iprom.WithMetricTypes(i.MetricTypes),
		iprom.WithMetricNameFilter(i.MetricNameFilter),
		iprom.WithMetricNameFilterIgnore(i.MetricNameFilterIgnore),
		iprom.WithMeasurementPrefix(i.MeasurementPrefix),
		iprom.WithMeasurementName(i.MeasurementName),
		iprom.WithMeasurements(i.Measurements),
		iprom.WithOutput(i.Output),
		iprom.WithMaxFileSize(i.MaxFileSize),
		iprom.WithTLSOpen(i.TLSOpen),
		iprom.WithUDSPath(i.UDSPath),
		iprom.WithCacertFile(i.CacertFile),
		iprom.WithCertFile(i.CertFile),
		iprom.WithKeyFile(i.KeyFile),
		iprom.WithTagsIgnore(i.TagsIgnore),
		iprom.WithTagsRename(i.TagsRename),
		iprom.WithAsLogging(i.AsLogging),
		iprom.WithIgnoreTagKV(i.IgnoreTagKV),
		iprom.WithHTTPHeaders(i.HTTPHeaders),
		iprom.WithTags(i.Tags),
		iprom.WithDisableHostTag(i.DisableHostTag),
		iprom.WithDisableInstanceTag(i.DisableInstanceTag),
		iprom.WithDisableInfoTag(i.DisableInfoTag),
		iprom.WithAuth(i.Auth),
	}

	pm, err := iprom.NewProm(opts...)
	if err != nil {
		i.l.Warnf("prom.NewProm: %s, ignored", err)
		return err
	}
	i.pm = pm
	i.isInitialized = true

	return nil
}

var maxPauseCh = inputs.ElectionPauseChannelLength

func NewProm() *Input {
	return &Input{
		chPause:     make(chan bool, maxPauseCh),
		MaxFileSize: defaultMaxFileSize,
		Source:      "prom",
		Interval:    "30s",
		Election:    true,
		Tags:        make(map[string]string),

		semStop: cliutils.NewSem(),
		Feeder:  io.DefaultFeeder(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return NewProm()
	})
}