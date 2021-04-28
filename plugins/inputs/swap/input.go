package swap

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	minInterval = time.Second
	maxInterval = time.Minute
)

var (
	inputName  = "swap"
	metricName = inputName
	swapLogger = logger.SLogger(inputName)
	sampleCfg  = `
[[inputs.swap]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'
  ## 
  [inputs.swap.tags]
    # tag1 = "a"
`
)

type swapMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *swapMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricName,
		Fields: map[string]interface{}{
			"total": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeIByte,
				Desc: "Host swap memory free"},
			"used": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeIByte,
				Desc: "Host swap memory used"},
			"free": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeIByte,
				Desc: "Host swap memory total"},
			"used_percent": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "Host swap memory percentage used"},
			"in": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeIByte,
				Desc: "Moving data from swap space to main memory of the machine"},
			"out": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeIByte,
				Desc: "Moving main memory contents to swap disk when main memory space fills up"},
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "主机名"},
		},
	}
}

func (m *swapMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

type Input struct {
	Interval             datakit.Duration
	Tags                 map[string]string
	collectCache         []inputs.Measurement
	collectCacheLast1Ptr inputs.Measurement
	swapStat             SwapStat
}

func (i *Input) appendMeasurement(name string, tags map[string]string, fields map[string]interface{}, ts time.Time) {
	tmp := &swapMeasurement{name: name, tags: tags, fields: fields, ts: ts}
	i.collectCache = append(i.collectCache, tmp)
	i.collectCacheLast1Ptr = tmp
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) Catalog() string {
	return "host"
}

func (i *Input) SampleConfig() string {
	return sampleCfg
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&swapMeasurement{},
	}
}

func (i *Input) Collect() error {
	i.collectCache = make([]inputs.Measurement, 0)
	swap, err := i.swapStat()
	ts := time.Now()
	if err != nil {
		return fmt.Errorf("error getting swap memory info: %s", err)
	}

	fields := map[string]interface{}{
		"total":        swap.Total,
		"used":         swap.Used,
		"free":         swap.Free,
		"used_percent": swap.UsedPercent,

		"in":  swap.Sin,
		"out": swap.Sout,
	}
	tags := map[string]string{}
	for k, v := range i.Tags {
		tags[k] = v
	}
	i.appendMeasurement(metricName, tags, fields, ts)

	return nil
}

func (i *Input) Run() {
	swapLogger.Infof("system input started")
	i.Interval.Duration = datakit.ProtectedInterval(minInterval, maxInterval, i.Interval.Duration)
	tick := time.NewTicker(i.Interval.Duration)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			start := time.Now()
			if err := i.Collect(); err == nil {
				if errFeed := inputs.FeedMeasurement(metricName, io.Metric, i.collectCache,
					&io.Option{CollectCost: time.Since(start)}); errFeed != nil {
					io.FeedLastError(inputName, errFeed.Error())
					swapLogger.Error(errFeed)
				}
			} else {
				io.FeedLastError(inputName, err.Error())
				swapLogger.Error(err)
			}
		case <-datakit.Exit.Wait():
			swapLogger.Infof("system input exit")
			return
		}
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			swapStat: PSSwapStat,
			Interval: datakit.Duration{Duration: time.Second * 10},
		}
	})
}
