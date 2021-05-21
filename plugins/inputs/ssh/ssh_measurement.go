package ssh

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"time"
)

type SshMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (s *SshMeasurement) LineProto() (*io.Point, error) {
	data, err := io.MakePoint(s.name, s.tags, s.fields, s.ts)
	return data, err
}

func (s *SshMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Fields: map[string]interface{}{
			"ssh_check":          &inputs.FieldInfo{DataType: inputs.Bool, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "ssh service status"},
			"ssh_err":            &inputs.FieldInfo{DataType: inputs.String, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "fail reason of connet ssh service"},
			"sftp_check":         &inputs.FieldInfo{DataType: inputs.Bool, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "sftp service status"},
			"sftp_err":           &inputs.FieldInfo{DataType: inputs.String, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "fail reason of connet sftp service"},
			"sftp_response_time": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "response time of sftp service"},
		},
		Tags: map[string]interface{}{
			"host": inputs.TagInfo{"the host of ssh"},
		},
	}
}
