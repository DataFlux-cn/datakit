// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package system collect system level metrics
package system

import (
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	conntrackutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/hostutil/conntrack"
	filefdutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/hostutil/filefd"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	_ inputs.ReadEnv   = (*Input)(nil)
	_ inputs.Singleton = (*Input)(nil)
)

const (
	minInterval = time.Second
	maxInterval = time.Minute
)

const (
	inputName           = "system"
	metricNameSystem    = "system"
	metricNameConntrack = "conntrack"
	metricNameFilefd    = "filefd"
	sampleCfg           = `
[[inputs.system]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'

  [inputs.system.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	Interval  datakit.Duration
	Fielddrop []string // Deprecated
	Tags      map[string]string

	collectCache []inputs.Measurement

	semStop *cliutils.Sem // start stop signal
}

func (ipt *Input) Singleton() {
}

func (*Input) Catalog() string {
	return "host"
}

func (*Input) SampleConfig() string {
	return sampleCfg
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&systemMeasurement{},
		&conntrackMeasurement{},
		&filefdMeasurement{},
	}
}

func (ipt *Input) Collect() error {
	// clear collectCache
	ipt.collectCache = make([]inputs.Measurement, 0)

	ts := time.Now()

	loadAvg, err := load.Avg()

	if err != nil && !strings.Contains(err.Error(), "not implemented") {
		return err
	}
	numCPUs, err := cpu.Counts(true)
	if err != nil {
		return err
	}

	tags := map[string]string{}
	for k, v := range ipt.Tags {
		tags[k] = v
	}

	if runtime.GOOS == "linux" {
		conntrackStat := conntrackutil.GetConntrackInfo()

		conntrackM := conntrackMeasurement{
			name: metricNameConntrack,
			fields: map[string]interface{}{
				"entries":             conntrackStat.Current,
				"entries_limit":       conntrackStat.Limit,
				"stat_found":          conntrackStat.Found,
				"stat_invalid":        conntrackStat.Invalid,
				"stat_ignore":         conntrackStat.Ignore,
				"stat_insert":         conntrackStat.Insert,
				"stat_insert_failed":  conntrackStat.InsertFailed,
				"stat_drop":           conntrackStat.Drop,
				"stat_early_drop":     conntrackStat.EarlyDrop,
				"stat_search_restart": conntrackStat.SearchRestart,
			},
			tags: tags,
			ts:   ts,
		}

		ipt.collectCache = append(ipt.collectCache, &conntrackM)

		filefdStat, err := filefdutil.GetFileFdInfo()
		if err != nil {
			l.Warnf("filefdutil.GetFileFdInfo(): %s, ignored", err.Error())
		} else {
			filefdM := filefdMeasurement{
				name: metricNameFilefd,
				fields: map[string]interface{}{
					"allocated":    filefdStat.Allocated,
					"maximum_mega": filefdStat.MaximumMega,
				},
				tags: tags,
				ts:   ts,
			}

			ipt.collectCache = append(ipt.collectCache, &filefdM)
		}
	}

	sysM := systemMeasurement{
		name: metricNameSystem,
		fields: map[string]interface{}{
			"load1":           loadAvg.Load1,
			"load5":           loadAvg.Load5,
			"load15":          loadAvg.Load15,
			"load1_per_core":  loadAvg.Load1 / float64(numCPUs),
			"load5_per_core":  loadAvg.Load5 / float64(numCPUs),
			"load15_per_core": loadAvg.Load15 / float64(numCPUs),
			"n_cpus":          numCPUs,
		},
		tags: tags,
		ts:   ts,
	}

	users, err := host.Users()
	if err != nil {
		l.Warnf("Users: %s, ignored", err.Error())
	}
	sysM.fields["n_users"] = len(users)

	uptime, err := host.Uptime()
	if err != nil {
		l.Warnf("Uptime: %s, ignored", err.Error())
	}
	sysM.fields["uptime"] = uptime

	ipt.collectCache = append(ipt.collectCache, &sysM)

	return err
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("system input started")
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()
	for {
		start := time.Now()
		if err := ipt.Collect(); err != nil {
			l.Errorf("Collect: %s", err)
			io.FeedLastError(inputName, err.Error())
		}

		if len(ipt.collectCache) > 0 {
			if err := inputs.FeedMeasurement(inputName, datakit.Metric, ipt.collectCache,
				&io.Option{CollectCost: time.Since(start)}); err != nil {
				l.Errorf("FeedMeasurement: %s", err)
			}
		}

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Info("system input exit")
			return

		case <-ipt.semStop.Wait():
			l.Info("system input return")
			return
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

// ReadEnv support envs：
//   ENV_INPUT_SYSTEM_TAGS : "a=b,c=d"
//   ENV_INPUT_SYSTEM_INTERVAL : datakit.Duration
func (ipt *Input) ReadEnv(envs map[string]string) {
	if tagsStr, ok := envs["ENV_INPUT_SYSTEM_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	//   ENV_INPUT_SYSTEM_INTERVAL : datakit.Duration
	if str, ok := envs["ENV_INPUT_SYSTEM_INTERVAL"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_SYSTEM_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.Interval.Duration = config.ProtectedInterval(minInterval,
				maxInterval,
				da)
		}
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Interval: datakit.Duration{Duration: time.Second * 10},

			semStop: cliutils.NewSem(),
			Tags:    make(map[string]string),
		}
	})
}
