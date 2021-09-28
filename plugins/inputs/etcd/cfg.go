package etcd

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	metricName  = inputName
	l           = logger.DefaultSLogger(inputName)
	minInterval = time.Second
	maxInterval = time.Second * 30
	sample      = ``
)

type Measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *Measurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Fields: map[string]interface{}{
			"server_has_leader":                        newCountFieldInfo("领导者是否存在。1是存在，0不是"),
			"server_leader_changes_seen_total":         newCountFieldInfo("解释到的领导者变更次数"),
			"server_proposals_committed_total":         newCountFieldInfo("提交的共识提案总数"),
			"server_proposals_applied_total":           newCountFieldInfo("已应用的共识提案总数"),
			"server_proposals_pending":                 newCountFieldInfo("当前待处理提案的数量"),
			"server_proposals_failed_total":            newCountFieldInfo("看到的失败提案总数"),
			"network_client_grpc_sent_bytes_total":     newCountFieldInfo("发送到 grpc 客户端的总字节数"),
			"network_client_grpc_received_bytes_total": newCountFieldInfo("接收到 grpc 客户端的总字节数"),
		},
		Tags: map[string]interface{}{},
	}
}

func newCountFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}