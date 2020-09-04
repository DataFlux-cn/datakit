package envoy

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "envoy"

	sampleCfg = `
[[inputs.prom]]
    # envoy metrics from http(https)://HOST:PORT/stats/prometheus
    # usually modify host and port
    # required
    url = "http://127.0.0.1:8090/stats/prometheus"
    
    # valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    # required
    interval = "10s"

    ## Optional TLS Config
    tls_open = false
    # tls_ca = "/tmp/ca.crt"
    # tls_cert = "/tmp/peer.crt"
    # tls_key = "/tmp/peer.key"
    
    ## Internal configuration. Don't modify.
    name = "envoy"
    ignore_measurement = ["envoy_http", "envoy_listener"]
    ignore_fields_key_prefix= ["envoy_server_worker"]

    # [inputs.prom.tags]
    # from = "127.0.0.1:9901"
    # tags1 = "value1"
`
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Envoy{}
	})
}

type Envoy struct {
}

func (*Envoy) SampleConfig() string {
	return sampleCfg
}

func (*Envoy) Catalog() string {
	return inputName
}

func (*Envoy) Run() {
}
