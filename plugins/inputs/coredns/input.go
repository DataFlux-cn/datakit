// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package coredns collect coreDNS metrics by using input prom
package coredns

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName    = "coredns"
	configSample = `
[[inputs.prom]]
## Exporter 地址
# 此处修改成CoreDNS的prom监听地址
url = "http://127.0.0.1:9153/metrics"

## 采集器别名
source = "coredns"

## 指标类型过滤, 可选值为 counter, gauge, histogram, summary
# 默认只采集 counter 和 gauge 类型的指标
# 如果为空，则不进行过滤
metric_types = ["counter", "gauge"]

## 指标名称过滤
# 支持正则，可以配置多个，即满足其中之一即可
# 如果为空，则不进行过滤
# CoreDNS的prom默认提供大量Go运行时的指标，这里忽略
metric_name_filter = ["^coredns_(acl|cache|dnssec|forward|grpc|hosts|template|dns)_([a-z_]+)$"]

## 指标集名称前缀
# 配置此项，可以给指标集名称添加前缀
# measurement_prefix = ""

## 指标集名称
# 默认会将指标名称以下划线"_"进行切割，切割后的第一个字段作为指标集名称，剩下字段作为当前指标名称
# 如果配置measurement_name, 则不进行指标名称的切割
# 最终的指标集名称会添加上measurement_prefix前缀
# measurement_name = "prom"

## 采集间隔 "ns", "us" (or "µs"), "ms", "s", "m", "h"
interval = "10s"

## 过滤tags, 可配置多个tag
# 匹配的tag将被忽略
# tags_ignore = [""]

## TLS 配置
tls_open = false
# tls_ca = "/tmp/ca.crt"
# tls_cert = "/tmp/peer.crt"
# tls_key = "/tmp/peer.key"

## 自定义指标集名称
# 可以将包含前缀prefix的指标归为一类指标集
# 自定义指标集名称配置优先measurement_name配置项

[[inputs.prom.measurements]]
	prefix = "coredns_acl_"
	name = "coredns_acl"
[[inputs.prom.measurements]]
	prefix = "coredns_cache_"
	name = "coredns_cache"
[[inputs.prom.measurements]]
	prefix = "coredns_dnssec_"
	name = "coredns_dnssec"
[[inputs.prom.measurements]]
	prefix = "coredns_forward_"
	name = "coredns_forward"
[[inputs.prom.measurements]]
	prefix = "coredns_grpc_"
	name = "coredns_grpc"
[[inputs.prom.measurements]]
	prefix = "coredns_hosts_"
	name = "coredns_hosts"
[[inputs.prom.measurements]]
	prefix = "coredns_template_"
	name = "coredns_template"
[[inputs.prom.measurements]]
	prefix = "coredns_dns_"
	name = "coredns"
`
)

type Input struct{}

var _ inputs.InputV2 = (*Input)(nil)

func (i *Input) Terminate() {
	// do nothing
}

func (i *Input) Catalog() string {
	return inputName
}

func (i *Input) SampleConfig() string {
	return configSample
}

func (i *Input) Run() {
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&ACLMeasurement{},
		&CacheMeasurement{},
		&DNSSecMeasurement{},
		&ForwardMeasurement{},
		&GrpcMeasurement{},
		&HostsMeasurement{},
		&TemplateMeasurement{},
		&PromMeasurement{},
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
