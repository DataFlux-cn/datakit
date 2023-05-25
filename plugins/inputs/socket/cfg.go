// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package socket

import (
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	KillGrace = 5 * time.Second
	TCP       = "tcp"
	UDP       = "udp"
)

var (
	inputName   = "socket"
	metricName  = inputName
	l           = logger.DefaultSLogger(inputName)
	minInterval = time.Second * 10
	maxInterval = time.Second * 60
	sample      = `
[[inputs.socket]]
  ## Support TCP/UDP.
  ## If the quantity to be detected is too large, it is recommended to open more collectors
  dest_url = ["tcp://host:port", "udp://host:port"]

  ## @param interval - number - optional - default: 30
  interval = "30s"
  ## @param interval - number - optional - default: 10
  udp_timeout = "10s"
  ## @param interval - number - optional - default: 10
  tcp_timeout = "10s"

[inputs.socket.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"`
)

type Input struct {
	DestURL    []string         `toml:"dest_url"`
	Interval   datakit.Duration `toml:"interval"` // 单位为秒
	UDPTimeOut datakit.Duration `toml:"udp_timeout"`
	TCPTimeOut datakit.Duration `toml:"tcp_timeout"`

	collectCache []inputs.Measurement
	Tags         map[string]string `toml:"tags"`
	semStop      *cliutils.Sem     // start stop signal
	platform     string
}

type TCPMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

type UDPMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *TCPMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOpt())
}

func (m *UDPMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOpt())
}

func (m *TCPMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "tcp",
		Tags: map[string]interface{}{
			"dest_host": &inputs.TagInfo{Desc: "示例 `wwww.baidu.com`"},
			"dest_port": &inputs.TagInfo{Desc: "示例 80"},
			"proto":     &inputs.TagInfo{Desc: "示例 `tcp`"},
		},
		Fields: map[string]interface{}{
			"response_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "TCP 连接时间",
			},
			"response_time_with_dns": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "连接时间（含 DNS 解析）",
			},
			"success": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "只有 1/-1 两种状态。1 表示成功/-1 表示失败",
			},
		},
	}
}

func (m *UDPMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "udp",
		Fields: map[string]interface{}{
			"success": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "只有 1/-1 两种状态。1 表示成功/-1 表示失败",
			},
		},
		Tags: map[string]interface{}{
			"dest_host": &inputs.TagInfo{Desc: "目的主机的 host"},
			"dest_port": &inputs.TagInfo{Desc: "目的主机的端口号"},
			"proto":     &inputs.TagInfo{Desc: "示例 `udp`"},
		},
	}
}
