package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"time"

	"github.com/influxdata/toml"
	"google.golang.org/grpc"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	swV3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/skywalkingGrpcV3/v3"
)

type SkywalkingServerV3 struct {}

type TraceAdapter struct {
	Source        string

	Duration      int64
	TimestampUs   int64
	Content       string

	Class         string
	ServiceName   string
	OperationName string
	ParentID      string
	TraceID       string
	SpanID        string
	IsError       string
	SpanType      string
	EndPoint      string

	Tags          map[string]string
}

type Skywalking struct {
	GrpcPort int32
	Tags     map[string]string
}

const (
	US_PER_SECOND   int64 = 1000000
	SPAN_TYPE_ENTRY  = "entry"
	SPAN_TYPE_LOCAL  = "local"
	SPAN_TYPE_EXIT   = "exit"

)

var (
	flagCfgStr    = flag.String("cfg", "", "toml config string")
	flagRPCServer = flag.String("rpc-server", "unix://"+datakit.GRPCDomainSock, "gRPC server")
	flagLog       = flag.String("log", filepath.Join(datakit.InstallDir, "externals", "skywalkingGrpcV3.log"), "log file")
	flagLogLevel  = flag.String("log-level", "info", "log file")

	l      *logger.Logger
	rpcCli dkio.DataKitClient
	skywalkingV3 Skywalking
)


func main() {
	flag.Parse()

	cfgdata, err := base64.StdEncoding.DecodeString(*flagCfgStr)
	if err != nil {
		panic(err)
	}

	logger.SetGlobalRootLogger(*flagLog, *flagLogLevel, logger.OPT_ENC_CONSOLE|logger.OPT_SHORT_CALLER)
	l = logger.SLogger("skywalkingGrpcV3")
	l.Infof("log level: %s", *flagLogLevel)

	if err := toml.Unmarshal(cfgdata, &skywalkingV3); err != nil {
		l.Errorf("failed to parse toml `%s': %s", *flagCfgStr, err)
		return
	}

	l.Infof("gRPC dial %s...", *flagRPCServer)
	conn, err := grpc.Dial(*flagRPCServer, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Second*5))
	if err != nil {
		l.Fatalf("connect RCP failed: %s", err)
	}

	l.Infof("gRPC connect %s ok", *flagRPCServer)
	defer conn.Close()

	rpcCli = dkio.NewDataKitClient(conn)

	skywalkingGrpcServRun(fmt.Sprintf(":%d", skywalkingV3.GrpcPort))
}


func skywalkingGrpcServRun(addr string) {
	l.Infof("skywalking V3 gRPC starting...")

	rpcListener, err := net.Listen("tcp", addr)
	if err != nil {
		l.Errorf("start skywalking V3 gRPC server %s failed: %v", addr, err)
		return
	}

	l.Infof("start skywalking V3 gRPC server on %s ok", addr)

	rpcServer := grpc.NewServer()
	swV3.RegisterTraceSegmentReportServiceServer(rpcServer, &SkywalkingServerV3{})
	if err := rpcServer.Serve(rpcListener); err != nil {
		l.Error(err)
	}
}

func (s *SkywalkingServerV3) Collect(tsc swV3.TraceSegmentReportService_CollectServer) error {
	for {
		sgo, err := tsc.Recv()
		if err == io.EOF {
			return tsc.SendAndClose(&swV3.Commands{})
		}
		if err != nil {
			return err
		}
		err = skywalkGrpcToLineProto(sgo)
		if err != nil {
			return err
		}
	}
	return nil
}

func skywalkGrpcToLineProto(sg *swV3.SegmentObject) error {
	var lines [][]byte
	for _, span := range sg.Spans {
		t := TraceAdapter{}

		t.Source = "skywalking"

		t.Duration = (span.EndTime -span.StartTime)*1000
		t.TimestampUs = span.StartTime * 1000
		js ,err := json.Marshal(span)
		if err != nil {
			return err
		}
		t.Content = string(js)
		t.Class         = "tracing"
		t.ServiceName   = sg.Service
		t.OperationName = span.OperationName
		if span.SpanType == swV3.SpanType_Entry {
			if len(span.Refs) > 0 {
				t.ParentID      = fmt.Sprintf("%s%d", span.Refs[0].ParentTraceSegmentId,
					span.Refs[0].ParentSpanId)
			}
		} else {
			t.ParentID      = fmt.Sprintf("%s%d", sg.TraceSegmentId, span.ParentSpanId)
		}

		t.TraceID       = sg.TraceId
		t.SpanID        = fmt.Sprintf("%s%d", sg.TraceSegmentId, span.SpanId)
		if span.IsError {
			t.IsError   = "true"
		}
		if span.SpanType == swV3.SpanType_Entry {
			t.SpanType  = SPAN_TYPE_ENTRY
		} else if span.SpanType == swV3.SpanType_Local {
			t.SpanType  = SPAN_TYPE_LOCAL
		} else {
			t.SpanType  = SPAN_TYPE_EXIT
		}
		t.EndPoint      = span.Peer

		t.Tags = skywalkingV3.Tags
		pt, err := t.MkLineProto()
		if err != nil {
			l.Error(err)
			continue
		}
		lines = append(lines, pt)
		l.Debug(string(pt))
	}

	if len(lines) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := rpcCli.Send(ctx, &dkio.Request{
		Lines:     bytes.Join(lines, []byte("\n")),
		Precision: "ns",
		Name:      "skywalkingGrpcV3",
	})
	if err != nil {
		l.Errorf("feed error: %s", err.Error())
		return err
	}
	l.Debugf("feed %d points, error: `%s'", r.GetPoints(), r.GetErr())
	return nil
}


func (tAdpt *TraceAdapter) MkLineProto() ([]byte, error) {
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["__class"]         = tAdpt.Class
	tags["__operationName"] = tAdpt.OperationName
	tags["__serviceName"]   = tAdpt.ServiceName
	tags["__parentID"]      = tAdpt.ParentID
	tags["__traceID"]       = tAdpt.TraceID
	tags["__spanID"]        = tAdpt.SpanID

	for tag, tagV := range tAdpt.Tags {
		tags[tag] = tagV
	}
	if tAdpt.IsError == "true" {
		tags["__isError"] = "true"
	} else {
		tags["__isError"] = "false"
	}

	if tAdpt.EndPoint != "" {
		tags["__endpoint"] = tAdpt.EndPoint
	} else {
		tags["__endpoint"] = "null"
	}

	if tAdpt.SpanType != "" {
		tags["__spanType"] = tAdpt.SpanType
	} else {
		tags["__spanType"] = SPAN_TYPE_ENTRY
	}

	fields["__duration"] = tAdpt.Duration
	fields["__content"]  = tAdpt.Content

	ts := time.Unix(tAdpt.TimestampUs/US_PER_SECOND, (tAdpt.TimestampUs%US_PER_SECOND)*1000)

	pt, err := dkio.MakeMetric(tAdpt.Source, tags, fields, ts)
	if err != nil {
		return nil, fmt.Errorf("build metric err: %s", err)
	}

	lineProtoStr := string(pt)
	l.Debugf(lineProtoStr)

	return pt, nil
}
