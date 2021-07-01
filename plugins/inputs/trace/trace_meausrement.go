package trace

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type TraceMeasurement struct {
	Name   string
	Tags   map[string]string
	Fields map[string]interface{}
	Ts     time.Time
}

func (t *TraceMeasurement) LineProto() (*io.Point, error) {
	data, err := io.MakePoint(t.Name, t.Tags, t.Fields, t.Ts)
	return data, err
}

func (t *TraceMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "",
		Fields: map[string]interface{}{
			FIELD_PARENTID: &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The parent id of current span."},
			FIELD_TRACEID:  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Trace id."},
			FIELD_SPANID:   &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Span id."},
			FIELD_DURATION: &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationUS, Desc: "Duration of span."},
			FIELD_START:    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationUS, Desc: "start time of span."},
			FIELD_MSG:      &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The origin content of span."},
			FIELD_RESOURCE: &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The resource name."},
			FIELD_PID:      &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The application process id."},
		},
		Tags: map[string]interface{}{
			TAG_PROJECT:        &inputs.TagInfo{"project name"},
			TAG_OPERATION:      &inputs.TagInfo{"span name"},
			TAG_SERVICE:        &inputs.TagInfo{"service name"},
			TAG_VERSION:        &inputs.TagInfo{"application version info"},
			TAG_ENV:            &inputs.TagInfo{"application environment info"},
			TAG_HTTP_METHOD:    &inputs.TagInfo{"http request method name"},
			TAG_HTTP_CODE:      &inputs.TagInfo{"http response code"},
			TAG_TYPE:           &inputs.TagInfo{"span  service type"},
			TAG_ENDPOINT:       &inputs.TagInfo{"endpoint info"},
			TAG_SPAN_STATUS:    &inputs.TagInfo{"span status"},
			TAG_SPAN_TYPE:      &inputs.TagInfo{"span type"},
			TAG_CONTAINER_HOST: &inputs.TagInfo{"container hostname"},
		},
	}
}
