package etcd

import (
	"context"
	"sync"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type Etcd struct {
	Config Config `toml:"etcd"`

	ctx    context.Context
	cancel context.CancelFunc
	acc    telegraf.Accumulator
	wg     *sync.WaitGroup
}

func init() {
	inputs.Add(pluginName, func() inputs.Input {
		e := &Etcd{}
		e.ctx, e.cancel = context.WithCancel(context.Background())
		return e
	})
}

func (e *Etcd) Start(acc telegraf.Accumulator) error {

	e.acc = acc
	e.wg = new(sync.WaitGroup)

	for _, sub := range e.Config.Subscribes {
		e.wg.Add(1)
		stream := newStream(&sub, e)
		go stream.start(e.wg)
	}

	return nil
}

func (e *Etcd) Stop() {
	e.cancel()
	e.wg.Wait()
}

func (_ *Etcd) SampleConfig() string {
	return etcdConfigSample
}

func (_ *Etcd) Catalog() string {
	return "etcd"
}

func (_ *Etcd) Description() string {
	return "Convert Etcd collection data to Dataway"
}

func (_ *Etcd) Gather(telegraf.Accumulator) error {
	return nil
}

func (e *Etcd) ProcessPts(pts []*influxdb.Point) error {
	for _, pt := range pts {
		fields, err := pt.Fields()
		if err != nil {
			return err
		}
		pt_metric, err := metric.New(pt.Name(), pt.Tags(), fields, pt.Time())
		if err != nil {
			return err
		}
		e.acc.AddMetric(pt_metric)
	}
	return nil
}
