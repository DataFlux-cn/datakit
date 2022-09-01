// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/opentelemetry/collector"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/opentelemetry/mock"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/metadata"
)

func TestExportTrace_Export(t *testing.T) {
	trace := &ExportTrace{storage: collector.NewSpansStorage(nil)}
	endpoint := "localhost:20010"
	m := mock.MockOtlpGrpcCollector{Trace: trace}
	go m.StartServer(t, endpoint)
	<-time.After(5 * time.Millisecond)
	t.Log("start server")
	ctx := context.Background()
	exp := mock.NewGRPCExporter(t, ctx, endpoint)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(
			exp,
			// add following two options to ensure flush
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(10),
		),
	)
	t.Cleanup(func() { require.NoError(t, tp.Shutdown(ctx)) })

	tr := tp.Tracer("test-tracer")
	testKvs := []attribute.KeyValue{
		attribute.Int("Int", 1),
		attribute.Int64("Int64", int64(3)),
		attribute.Float64("Float64", 2.22),
		attribute.Bool("Bool", true),
		attribute.String("String", "test"),
	}
	_, span := tr.Start(ctx, "AlwaysSample")
	span.SetAttributes(testKvs...)
	time.Sleep(5 * time.Millisecond) // span.Duration
	span.End()
	t.Log("span end")
	// Flush and close.
	func() {
		ctx, cancel := mock.ContextWithTimeout(ctx, t, 10*time.Second)
		defer cancel()
		require.NoError(t, tp.Shutdown(ctx))
	}()

	// Wait >2 cycles.
	<-time.After(40 * time.Millisecond)

	// Now shutdown the exporter
	require.NoError(t, exp.Shutdown(ctx))

	// Shutdown the collector too so that we can begin
	// verification checks of expected data back.
	m.StopFunc()
	t.Log("stop server")
	expected := map[string]string{
		"Int":     "1",
		"Int64":   "3",
		"Float64": "2.22",
		"Bool":    "true",
		"String":  "test",
	}
	dktraces := trace.storage.GetDKTrace()

	if len(dktraces) != 1 {
		t.Errorf("dktraces.len != 1")
		return
	}

	for _, dktrace := range dktraces {
		if len(dktrace) != 1 {
			t.Errorf("dktrace.len != 1")
			return
		}
		for _, datakitSpan := range dktrace {
			if len(datakitSpan.Tags) < 5 {
				t.Errorf("tags count less 5")
				return
			}
			for key, val := range datakitSpan.Tags {
				for rkey := range expected {
					if key == rkey {
						if rval, ok := expected[rkey]; !ok || rval != val {
							t.Errorf("key=%s dk_span_val=%s  expetrd_val=%s", key, val, rval)
						}
					}
				}
			}
			if datakitSpan.Resource != "AlwaysSample" {
				t.Errorf("span.resource is %s  and real name is AlwaysSample", datakitSpan.Resource)
			}

			if datakitSpan.Operation != "AlwaysSample" {
				t.Errorf("span.Operation is %s  and real name is AlwaysSample", datakitSpan.Operation)
			}
			bts, _ := json.MarshalIndent(datakitSpan, "    ", "  ")
			t.Logf("json span = \n %s", string(bts))
		}
	}
}

func TestExportMetric_Export(t *testing.T) {
	metric := &ExportMetric{storage: collector.NewSpansStorage(nil)}
	endpoint := "localhost:20010"
	m := mock.MockOtlpGrpcCollector{Metric: metric}
	go m.StartServer(t, endpoint)
	<-time.After(5 * time.Millisecond)
	t.Log("start server")

	ctx := context.Background()
	exp := mock.NewMetricGRPCExporter(t, ctx, endpoint)

	err := exp.Export(ctx, mock.TestResource, mock.OneRecord)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := exp.Shutdown(ctx); err != nil {
			panic(err)
		}
	}()
	m.StopFunc()
	ms := metric.storage.GetDKMetric()
	if len(ms) != 1 {
		t.Errorf("metric len != 1")
	}
	want := &collector.OtelResourceMetric{
		Operation: "foo",
		Attributes: map[string]string{
			"abc": "def",
			"one": "1",
		},
		Resource:  "onelib",
		ValueType: "int",
		Value:     42,
		StartTime: uint64(time.Date(2020, time.December, 8, 19, 15, 0, 0, time.UTC).UnixNano()),
		UnitTime:  uint64(time.Date(2020, time.December, 8, 19, 16, 0, 0, time.UTC).UnixNano()),
	}

	got := ms[0]
	if !reflect.DeepEqual(got.Attributes, want.Attributes) {
		t.Errorf("tags got.tag=%+v  want.tag=%+v", got.Attributes, want.Attributes)
	}
	if got.Operation != want.Operation {
		t.Errorf("operation got=%s want=%s", got.Operation, want.Operation)
	}
}

func Test_checkHandler(t *testing.T) {
	type args struct {
		headers map[string]string
		md      metadata.MD
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "case_no_header",
			args: args{
				headers: map[string]string{},
				md:      make(metadata.MD),
			},
			want: true,
		},
		{
			name: "case_check_header_len_1",
			args: args{
				headers: map[string]string{"must_have_header1": "1"},
				md:      map[string][]string{"must_have_header1": {"1"}},
			},
			want: true,
		},
		{
			name: "case_check_header_len_2",
			args: args{
				headers: map[string]string{"must_have_header1": "1,2"},
				md:      map[string][]string{"must_have_header1": {"1", "2"}},
			},
			want: true,
		},
		{
			name: "case_check_invalid_header",
			args: args{
				headers: map[string]string{"must_have_header1": "1,2", "header2": "2"},
				md:      map[string][]string{"must_have_header1": {"1", "2"}},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkHandler(tt.args.headers, tt.args.md); got != tt.want {
				t.Errorf("checkHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}
