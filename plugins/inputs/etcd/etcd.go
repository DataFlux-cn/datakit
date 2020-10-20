package etcd

import (
	ifxcli "github.com/influxdata/influxdb1-client/v2"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/prom"
)

const (
	inputName = "etcd"

	sampleCfg = `
[[inputs.prom]]
    # etcd metrics from http(https)://HOST:PORT/metrics
    # usually modify host and port
    # required
    url = "http://127.0.0.1:2379/metrics"
    
    # valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    # required
    interval = "10s"
    
    ## Optional TLS Config
    tls_open = false
    # tls_ca = "/tmp/ca.crt"
    # tls_cert = "/tmp/peer.crt"
    # tls_key = "/tmp/peer.key"

    # [inputs.prom.tags]
    # from = "127.0.0.1:2379"
    # tags1 = "value1"
`
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &prom.Prom{
			Interval:       datakit.Cfg.MainCfg.Interval,
			InputName:      inputName,
			CatalogStr:     inputName,
			SampleCfg:      sampleCfg,
			Tags:           make(map[string]string),
			IgnoreFunc:     ignore,
			PromToNameFunc: nil,
		}
	})
}

var ignoreMeasurements = []string{
	"etcd_grpc_server",
	"etcd_server",
	"etcd_go_info",
	"etcd_cluster",
}

func ignore(pt *ifxcli.Point) bool {
	for _, m := range ignoreMeasurements {
		if pt.Name() == m {
			return true
		}
	}
	return false
}
