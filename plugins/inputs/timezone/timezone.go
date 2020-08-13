package timezone

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type TimeIoFeed func(data []byte, category, name string) error

type Timezone struct {
	Active      bool
	Interval    interface{}
	MetricsName string
	Tags        map[string]string
}

type TzInput struct {
	Timezone
}

type TzOutput struct {
	ioFeed TimeIoFeed
}

type TzParams struct {
	input  TzInput
	output TzOutput
	log    *logger.Logger
}

const (
	defaultMetricName = "timezone"
	defaultInterval   = "60s"
)

var (
	inputName = "timezone"
	Sample    = `### active     : whether to monitor timezone changes.
### interval   : monitor interval, the default value is "60s".
### metricsName: the name of metric, default is "timezone"

[inputs.timezone]
  active      = true
  interval    = "60s"
  metricsName = "timezone"
  [inputs.timezone.tags]
		host = '{{.Hostname}}'
#    tag1 = "tag1"
#    tag2 = "tag2"
#    tagn = "tagn"`
)

func (t *Timezone) SampleConfig() string {
	return Sample
}

func (t *Timezone) Catalog() string {
	return "timezone"
}

func (t *Timezone) Run() {
	if t.Active == false {
		return
	}

	if t.Interval == nil {
		t.Interval = defaultInterval
	}

	if t.MetricsName == "" {
		t.MetricsName = defaultMetricName
	}

	input := TzInput{*t}
	output := TzOutput{io.NamedFeed}
	p := TzParams{input, output, logger.SLogger("timezone")}

	p.log.Info("timezone input started...")
	p.gather()
}

func (p *TzParams) gather() {
	var d time.Duration
	var err error

	switch p.input.Interval.(type) {
	case int64:
		d = time.Duration(p.input.Interval.(int64)) * time.Second
	case string:
		d, err = time.ParseDuration(p.input.Interval.(string))
		if err != nil {
			p.log.Errorf("parse interval err: %s", err.Error())
			return
		}
	default:
		p.log.Errorf("interval type unsupported")
		return
	}

	tick := time.NewTicker(d)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			err = p.getMetrics()
			if err != nil {
				p.log.Errorf("getMetrics err: %s", err.Error())
			}

		case <-datakit.Exit.Wait():
			p.log.Info("input timezone exit")
			return
		}
	}
}

func (p *TzParams) getMetrics() error {
	fields := make(map[string]interface{})

	timezone, err := getOsTimezone()
	if err != nil {
		return err
	}

	fields["tz"] = timezone

	pt, err := io.MakeMetric(p.input.MetricsName, p.input.Tags, fields, time.Now())
	if err != nil {
		return err
	}

	if err := p.output.ioFeed(pt, io.Metric, inputName); err != nil {
		return err
	}
	return nil
}

func getOsTimezone() (string, error) {
	var outInfo bytes.Buffer
	os := runtime.GOOS

	if os == "linux" || os == "darwin" {
		cmd := exec.Command("date", `+%Z`)
		cmd.Stdout = &outInfo
		cmd.Run()
		return strings.Trim(outInfo.String(), "\n"), nil
	} else if os == "windows" {
		cmd := exec.Command("tzutil.exe", "/g")
		cmd.Stdout = &outInfo
		cmd.Run()
		return outInfo.String(), nil
	} else {
		return "", fmt.Errorf("Os: %s unsuport get timezone", os)
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		tz := &Timezone{}
		return tz
	})
}
