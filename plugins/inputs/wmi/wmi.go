// +build windows

package wmi

import (
	"context"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	moduleLogger *logger.Logger
)

func (_ *Instance) SampleConfig() string {
	return sampleConfig
}

// func (_ *WmiAgent) Description() string {
// 	return `Collect metrics from Windows WMI.`
// }

func (_ *Instance) Catalog() string {
	return `wmi`
}

func (ag *Instance) Run() {

	moduleLogger = logger.SLogger(inputName)

	go func() {
		<-datakit.Exit.Wait()
		ag.cancelFun()
	}()

	if ag.MetricName == "" {
		ag.MetricName = "WMI"
	}

	if ag.Interval.Duration == 0 {
		ag.Interval.Duration = time.Minute * 5
	}

	ag.run(ag.ctx)
}

func (r *Instance) run(ctx context.Context) error {

	for {

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		for _, query := range r.Queries {

			select {
			case <-ctx.Done():
				return context.Canceled
			default:
			}

			if query.lastTime.IsZero() {
				query.lastTime = time.Now()
			} else {
				if time.Now().Sub(query.lastTime) < query.Interval.Duration {
					continue
				}
			}

			sql, err := query.ToSql()
			if err != nil {
				moduleLogger.Warnf("%s", err)
				continue
			}

			props := []string{}

			for _, ms := range query.Metrics {
				props = append(props, ms[0])
			}

			fieldsArr, err := DefaultClient.QueryEx(sql, props)
			if err != nil {
				moduleLogger.Errorf("query failed, %s", err)
				continue

			}

			for _, fields := range fieldsArr {
				io.FeedEx(io.Metric, r.MetricName, nil, fields)
			}

			query.lastTime = time.Now()
		}

		internal.SleepContext(ctx, time.Second)
	}
}

func NewAgent() *Instance {
	ac := &Instance{}
	ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
	return ac
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return NewAgent()
	})
}
