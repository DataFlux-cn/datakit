// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rabbitmq

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func getNode(n *Input) {
	var Nodes []Node
	err := n.requestJSON("/api/nodes", &Nodes)
	if err != nil {
		l.Error(err.Error())
		n.lastErr = err
		return
	}
	ts := time.Now()
	for _, node := range Nodes {
		tags := map[string]string{
			"url":       n.URL,
			"node_name": node.Name,
		}
		for k, v := range n.Tags {
			tags[k] = v
		}
		fields := map[string]interface{}{
			"disk_free_alarm":   node.DiskFreeAlarm,
			"disk_free":         node.DiskFree,
			"fd_used":           node.FdUsed,
			"mem_alarm":         node.MemAlarm,
			"mem_limit":         node.MemLimit,
			"mem_used":          node.MemUsed,
			"run_queue":         node.RunQueue,
			"running":           node.Running,
			"sockets_used":      node.SocketsUsed,
			"io_write_avg_time": node.IoWriteAvgTime,
			"io_read_avg_time":  node.IoReadAvgTime,
		}
		metric := &NodeMeasurement{
			name:     NodeMetric,
			tags:     tags,
			fields:   fields,
			ts:       ts,
			election: n.Election,
		}
		metricAppend(metric)
	}
}

type NodeMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	ts       time.Time
	election bool
}

func (m *NodeMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOptElectionV2(m.election))
}

//nolint:lll
func (m *NodeMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: NodeMetric,
		Fields: map[string]interface{}{
			"disk_free_alarm": newOtherFieldInfo(inputs.Bool, inputs.Gauge, inputs.UnknownUnit, "Does the node have disk alarm"),
			"disk_free":       newByteFieldInfo("Current free disk space"),
			"fd_used":         newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.UnknownUnit, "Used file descriptors"),
			"mem_alarm":       newOtherFieldInfo(inputs.Bool, inputs.Gauge, inputs.UnknownUnit, "Does the node have mem alarm"),
			"mem_limit":       newByteFieldInfo("Memory usage high watermark in bytes"),
			"mem_used":        newByteFieldInfo("Memory used in bytes"),
			"run_queue":       newCountFieldInfo("Average number of Erlang processes waiting to run"),
			"running":         newOtherFieldInfo(inputs.Bool, inputs.Gauge, inputs.UnknownUnit, "Is the node running or not"),
			"sockets_used":    newCountFieldInfo("Number of file descriptors used as sockets"),

			// See: https://documentation.solarwinds.com/en/success_center/appoptics/content/kb/host_infrastructure/integrations/rabbitmq.htm
			"io_read_avg_time":  newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.DurationMS, "avg wall time (milliseconds) for each disk read operation in the last statistics interval"),
			"io_write_avg_time": newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.DurationMS, "avg wall time (milliseconds) for each disk write operation in the last statistics interval"),
			"io_seek_avg_time":  newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.DurationMS, "average wall time (milliseconds) for each seek operation in the last statistics interval"),
			"io_sync_avg_time":  newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.DurationMS, "average wall time (milliseconds) for each fsync() operation in the last statistics interval"),
		},

		Tags: map[string]interface{}{
			"url":       inputs.NewTagInfo("rabbitmq url"),
			"node_name": inputs.NewTagInfo("rabbitmq node name"),
		},
	}
}
