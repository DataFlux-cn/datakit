// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package nsq

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	nsqTopics = "nsq_topics"
	nsqNodes  = "nsq_nodes"
)

type nsqTopicMeasurement struct{}

//nolint:lll
func (*nsqTopicMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: nsqTopics,
		Desc: "NSQ 集群所有 topic 的指标",
		Tags: map[string]interface{}{
			"topic":   inputs.NewTagInfo("topic 名称"),
			"channel": inputs.NewTagInfo("channel 名称"),
		},
		Fields: map[string]interface{}{
			"depth":           &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "在当前 channel 中未被消费的消息总数"},
			"backend_depth":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "超出 men-queue-size 的未被消费的消息总数"},
			"in_flight_count": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "发送过程中或者客户端处理过程中的数量，客户端没有发送 FIN、REQ(重新入队列) 和超时的消息数量"},
			"deferred_count":  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "重新入队并且还没有准备好重新发送的消息数量"},
			"message_count":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "当前 channel 处理的消息总数量"},
			"requeue_count":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "超时或者客户端发送 REQ 的消息数量"},
			"timeout_count":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "超时未处理的消息数量"},
		},
	}
}

type nsqNodesMeasurement struct{}

//nolint:lll
func (*nsqNodesMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: nsqNodes,
		Desc: "NSQ 集群所有 node 的指标",
		Tags: map[string]interface{}{
			"server_host": inputs.NewTagInfo("服务地址，即 `host:ip`"),
			"host":        inputs.NewTagInfo("Hostname"),
		},
		Fields: map[string]interface{}{
			"depth":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "在当前 node 中未被消费的消息总数"},
			"backend_depth": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "超出 men-queue-size 的未被消费的消息总数"},
			"message_count": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "当前 node 处理的消息总数量"},
		},
	}
}
