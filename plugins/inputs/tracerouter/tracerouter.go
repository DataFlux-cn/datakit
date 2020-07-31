//+build !windows

package tracerouter

import (
	"bytes"
	"fmt"
	"time"

	"github.com/aeden/traceroute"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	l    *logger.Logger
	name = "tracerouter"
)

func (t *TraceRouter) Description() string {
	return "trace router"
}

func (t *TraceRouter) SampleConfig() string {
	return configSample
}

func (t *TraceRouter) Catalog() string {
	return "network"
}

func (t *TraceRouter) Init() error {
	return nil
}

func (t *TraceRouter) Gather() error {
	return nil
}

func (t *TraceRouter) Run() {
	l = logger.SLogger("tracerouter")

	l.Info("tracerouter input started...")

	t.checkCfg()

	tick := time.NewTicker(t.IntervalDuration)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			t.handle()
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		}
	}
}

func (t *TraceRouter) checkCfg() {
	// 采集频度
	t.IntervalDuration = 10 * time.Minute

	if t.Interval != "" {
		du, err := time.ParseDuration(t.Interval)
		if err != nil {
			l.Errorf("bad interval %s: %s, use default: 10m", t.Interval, err.Error())
		} else {
			t.IntervalDuration = du
		}
	}

	// 指标集名称
	if t.Metric != "" {
		t.Metric = "tracerouter"
	}
}

func (t *TraceRouter) handle() {
	host := t.Addr
	options := traceroute.TracerouteOptions{}
	options.SetMaxHops(traceroute.DEFAULT_MAX_HOPS + 1)
	options.SetFirstHop(traceroute.DEFAULT_FIRST_HOP)

	resHop, err := traceroute.Traceroute(host, &options)
	if err != nil {
		l.Errorf("tracerouter error %v", err)
	}

	t.parseHopData(resHop)
}

func (t *TraceRouter) parseHopData(resultHop traceroute.TracerouteResult) {
	tags := make(map[string]string)
	fields := make(map[string]interface{})
	lines := [][]byte{}
	tm := time.Now()

	for _, hop := range resultHop.Hops {
		if hop.Success {
			addr := fmt.Sprintf("%v.%v.%v.%v", hop.Address[0], hop.Address[1], hop.Address[2], hop.Address[3])

			tags["dist_addr"] = t.Addr
			fields["hop_num"] = fmt.Sprintf("%d", hop.TTL)
			fields["hop_addr"] = addr
			fields["resp_time"] = hop.ElapsedTime.Microseconds()

			pt, err := io.MakeMetric(t.Metric, tags, fields, tm)
			if err != nil {
				l.Errorf("make metric point error %v", err)
			}

			fmt.Println("======>", pt)

			lines = append(lines, pt)
		}
	}

	pushLines := bytes.Join(lines, []byte("\n"))

	err := io.NamedFeed(pushLines, io.Metric, name)
	if err != nil {
		l.Errorf("push metric point error %v", err)
	}
}

func init() {
	inputs.Add(name, func() inputs.Input {
		return &TraceRouter{}
	})
}
