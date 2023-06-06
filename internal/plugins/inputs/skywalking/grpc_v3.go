// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package skywalking

import (
	"context"
	"errors"
	"io"
	"net"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	commonv3old "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/skywalking/compiled/v8.3.0/common/v3"
	agentv3old "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/skywalking/compiled/v8.3.0/language/agent/v3"
	profilev3old "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/skywalking/compiled/v8.3.0/language/profile/v3"
	mgmtv3old "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/skywalking/compiled/v8.3.0/management/v3"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	configv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/skywalking/compiled/v9.3.0/agent/configuration/v3"
	commonv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/skywalking/compiled/v9.3.0/common/v3"
	eventv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/skywalking/compiled/v9.3.0/event/v3"
	agentv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/skywalking/compiled/v9.3.0/language/agent/v3"
	profilev3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/skywalking/compiled/v9.3.0/language/profile/v3"
	loggingv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/skywalking/compiled/v9.3.0/logging/v3"
	mgmtv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/skywalking/compiled/v9.3.0/management/v3"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

func runGRPCV3(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Errorf("### skywalking grpc server v3 listening on %s failed: %v", addr, err)

		return
	}
	log.Debugf("### skywalking grpc v3 listening on: %s", addr)

	skySvr = grpc.NewServer()
	// register API version 8.3.0
	agentv3old.RegisterTraceSegmentReportServiceServer(skySvr, &TraceReportServerV3Old{})
	agentv3old.RegisterJVMMetricReportServiceServer(skySvr, &JVMMetricReportServerV3Old{})
	profilev3old.RegisterProfileTaskServer(skySvr, &ProfileTaskServerV3Old{})
	mgmtv3old.RegisterManagementServiceServer(skySvr, &ManagementServerV3Old{})
	// register API version 9.3.0
	agentv3.RegisterTraceSegmentReportServiceServer(skySvr, &TraceReportServerV3{})
	eventv3.RegisterEventServiceServer(skySvr, &EventServerV3{})
	agentv3.RegisterJVMMetricReportServiceServer(skySvr, &JVMMetricReportServerV3{})
	loggingv3.RegisterLogReportServiceServer(skySvr, &LoggingServerV3{})
	profilev3.RegisterProfileTaskServer(skySvr, &ProfileTaskServerV3{})
	mgmtv3.RegisterManagementServiceServer(skySvr, &ManagementServerV3{})
	configv3.RegisterConfigurationDiscoveryServiceServer(skySvr, &DiscoveryServerV3{})

	if err = skySvr.Serve(listener); err != nil {
		log.Error(err.Error())
	}

	log.Debug("### skywalking v3 exits")
}

type TraceReportServerV3Old struct {
	agentv3old.UnimplementedTraceSegmentReportServiceServer
}

func (*TraceReportServerV3Old) Collect(tsr agentv3old.TraceSegmentReportService_CollectServer) error {
	for {
		segobj, err := tsr.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return tsr.SendAndClose(&commonv3old.Commands{})
			}
			log.Error(err.Error())

			return err
		}
		log.Debugf("### TraceReportServerV3Old:Collect SegmentObject:%#v", segobj)

		bts, err := proto.Marshal(segobj)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		if localCache == nil || !localCache.Enabled() {
			newSegObj := &agentv3.SegmentObject{}
			if err = proto.Unmarshal(bts, newSegObj); err != nil {
				log.Error(err.Error())
				continue
			}
			dktrace := parseSegmentObjectV3(newSegObj)
			if len(dktrace) != 0 && afterGatherRun != nil {
				afterGatherRun.Run(inputName, itrace.DatakitTraces{dktrace}, false)
			}
		} else {
			if err = localCache.Put(storage.SKY_WALKING_GRPC_KEY, bts); err != nil {
				log.Error(err.Error())
			}
		}
	}
}

func (*TraceReportServerV3Old) CollectInSync(ctx context.Context, col *agentv3old.SegmentCollection) (*commonv3old.Commands, error) {
	log.Debugf("### TraceReportServerV3Old:CollectInSync SegmentCollection: %#v", col)

	for _, segobj := range col.Segments {
		bts, err := proto.Marshal(segobj)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		if localCache == nil || !localCache.Enabled() {
			newSegObj := &agentv3.SegmentObject{}
			if err = proto.Unmarshal(bts, newSegObj); err != nil {
				log.Error(err.Error())
				continue
			}
			dktrace := parseSegmentObjectV3(newSegObj)
			if len(dktrace) != 0 && afterGatherRun != nil {
				afterGatherRun.Run(inputName, itrace.DatakitTraces{dktrace}, false)
			}
		} else {
			if err = localCache.Put(storage.SKY_WALKING_GRPC_KEY, bts); err != nil {
				log.Error(err.Error())
			}
		}
	}

	return &commonv3old.Commands{}, nil
}

type JVMMetricReportServerV3Old struct {
	agentv3old.UnimplementedJVMMetricReportServiceServer
}

func (*JVMMetricReportServerV3Old) Collect(ctx context.Context, jvm *agentv3old.JVMMetricCollection) (*commonv3old.Commands, error) {
	log.Debugf("### JVMMetricReportServerV3Old:Collect %#v", jvm)

	start := time.Now()
	bts, err := proto.Marshal(jvm)
	if err != nil {
		log.Error(err.Error())

		return &commonv3old.Commands{}, err
	}
	newjvm := &agentv3.JVMMetricCollection{}
	if err = proto.Unmarshal(bts, newjvm); err != nil {
		log.Error(err.Error())

		return &commonv3old.Commands{}, err
	}

	metrics := processMetricsV3(newjvm, start)
	if len(metrics) != 0 {
		if err := inputs.FeedMeasurement(jvmMetricName, datakit.Metric, metrics, &dkio.Option{CollectCost: time.Since(start)}); err != nil {
			dkio.FeedLastError(jvmMetricName, err.Error())
		}
	}

	return &commonv3old.Commands{}, nil
}

type ProfileTaskServerV3Old struct {
	profilev3old.UnimplementedProfileTaskServer
}

func (*ProfileTaskServerV3Old) GetProfileTaskCommands(ctx context.Context,
	task *profilev3old.ProfileTaskCommandQuery,
) (*commonv3old.Commands, error) {
	log.Debugf("### ProfileTaskServerV3Old:GetProfileTaskCommands ProfileTaskCommandQuery: %#v", task)

	return &commonv3old.Commands{}, nil
}

func (*ProfileTaskServerV3Old) CollectSnapshot(psrv profilev3old.ProfileTask_CollectSnapshotServer) error {
	for {
		profile, err := psrv.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return psrv.SendAndClose(&commonv3old.Commands{})
			}
			log.Debug(err.Error())

			return err
		}
		log.Debugf("### ProfileTaskServerV3Old:CollectSnapshot ThreadSnapshot: %#v", profile)

		newProfile := profilev3.ThreadSnapshot{
			TaskId:         profile.TaskId,
			TraceSegmentId: profile.TraceSegmentId,
			Time:           profile.Time,
			Sequence:       profile.Sequence,
		}
		if profile.Stack != nil {
			newProfile.Stack = &profilev3.ThreadStack{CodeSignatures: profile.Stack.CodeSignatures}
		}
		processProfileV3(&newProfile)
	}
}

func (*ProfileTaskServerV3Old) ReportTaskFinish(ctx context.Context, reporter *profilev3old.ProfileTaskFinishReport) (*commonv3old.Commands, error) {
	log.Debugf("### ProfileTaskServerV3Old:ReportTaskFinish ProfileTaskFinishReport: %#v", reporter)

	return &commonv3old.Commands{}, nil
}

type ManagementServerV3Old struct {
	mgmtv3old.UnimplementedManagementServiceServer
}

func (*ManagementServerV3Old) ReportInstanceProperties(ctx context.Context, mgmt *mgmtv3old.InstanceProperties) (*commonv3old.Commands, error) {
	log.Debugf("### ManagementServerV3Old:ReportInstanceProperties InstanceProperties: %#v", mgmt)

	return &commonv3old.Commands{}, nil
}

func (*ManagementServerV3Old) KeepAlive(ctx context.Context, ping *mgmtv3old.InstancePingPkg) (*commonv3old.Commands, error) {
	log.Debugf("### ManagementServerV3Old:KeepAlive InstancePingPkg: %#v", ping)

	return &commonv3old.Commands{}, nil
}

type TraceReportServerV3 struct {
	agentv3.UnimplementedTraceSegmentReportServiceServer
}

func (*TraceReportServerV3) Collect(tsr agentv3.TraceSegmentReportService_CollectServer) error {
	for {
		segobj, err := tsr.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return tsr.SendAndClose(&commonv3.Commands{})
			}
			log.Error(err.Error())

			return err
		}
		log.Debugf("### TraceReportServerV3:Collect SegmentObject: %#v", segobj)

		if localCache == nil || !localCache.Enabled() {
			dktrace := parseSegmentObjectV3(segobj)
			if len(dktrace) != 0 && afterGatherRun != nil {
				afterGatherRun.Run(inputName, itrace.DatakitTraces{dktrace}, false)
			}
		} else {
			if bts, err := proto.Marshal(segobj); err != nil {
				log.Error(err.Error())
			} else {
				if err = localCache.Put(storage.SKY_WALKING_GRPC_KEY, bts); err != nil {
					log.Error(err.Error())
				}
			}
		}
	}
}

func (*TraceReportServerV3) CollectInSync(ctx context.Context, col *agentv3.SegmentCollection) (*commonv3.Commands, error) {
	log.Debugf("### TraceReportServerV3:CollectInSync SegmentCollection: %#v", col)

	for _, segobj := range col.Segments {
		if localCache == nil || !localCache.Enabled() {
			dktrace := parseSegmentObjectV3(segobj)
			if len(dktrace) != 0 && afterGatherRun != nil {
				afterGatherRun.Run(inputName, itrace.DatakitTraces{dktrace}, false)
			}
		} else {
			if bts, err := proto.Marshal(segobj); err != nil {
				log.Error(err.Error())
				continue
			} else {
				if err = localCache.Put(storage.SKY_WALKING_GRPC_KEY, bts); err != nil {
					log.Error(err.Error())
				}
			}
		}
	}

	return &commonv3.Commands{}, nil
}

type EventServerV3 struct {
	eventv3.UnimplementedEventServiceServer
}

func (*EventServerV3) Collect(esrv eventv3.EventService_CollectServer) error {
	for {
		event, err := esrv.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return esrv.SendAndClose(&commonv3.Commands{})
			}
			log.Debug(err.Error())

			return err
		}

		log.Debugf("### EventServerV3:Collect Event: %#v", event)
	}
}

type JVMMetricReportServerV3 struct {
	agentv3.UnimplementedJVMMetricReportServiceServer
}

func (*JVMMetricReportServerV3) Collect(ctx context.Context, jvm *agentv3.JVMMetricCollection) (*commonv3.Commands, error) {
	log.Debugf("### JVMMetricReportServerV3:Collect JVMMetricCollection: %#v", jvm)

	start := time.Now()
	metrics := processMetricsV3(jvm, start)
	if len(metrics) != 0 {
		if err := inputs.FeedMeasurement(jvmMetricName, datakit.Metric, metrics, &dkio.Option{CollectCost: time.Since(start)}); err != nil {
			dkio.FeedLastError(jvmMetricName, err.Error())
		}
	}

	return &commonv3.Commands{}, nil
}

type LoggingServerV3 struct {
	loggingv3.UnsafeLogReportServiceServer
}

func (*LoggingServerV3) Collect(server loggingv3.LogReportService_CollectServer) error {
	for {
		logData, err := server.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return server.SendAndClose(&commonv3.Commands{})
			}
			log.Debug(err.Error())

			return err
		}
		log.Debugf("### LoggingServerV3:Collect LogData: %#v", logData)

		if pt, err := processLogV3(logData); err != nil {
			log.Error(err.Error())
		} else {
			if err = dkio.Feed(logData.Service, datakit.Logging, []*point.Point{pt}, nil); err != nil {
				log.Error(err.Error())
			}
		}
	}
}

type ProfileTaskServerV3 struct {
	profilev3.UnimplementedProfileTaskServer
}

func (*ProfileTaskServerV3) GetProfileTaskCommands(ctx context.Context, task *profilev3.ProfileTaskCommandQuery) (*commonv3.Commands, error) {
	log.Debugf("### ProfileTaskServerV3:GetProfileTaskCommands ProfileTaskCommandQuery: %#v", task)

	return &commonv3.Commands{}, nil
}

func (*ProfileTaskServerV3) CollectSnapshot(psrv profilev3.ProfileTask_CollectSnapshotServer) error {
	for {
		profile, err := psrv.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return psrv.SendAndClose(&commonv3.Commands{})
			}
			log.Debug(err.Error())

			return err
		}
		log.Debugf("### ProfileTaskServerV3:CollectSnapshot ThreadSnapshot: %#v", profile)

		processProfileV3(profile)
	}
}

func (*ProfileTaskServerV3) ReportTaskFinish(ctx context.Context, reporter *profilev3.ProfileTaskFinishReport) (*commonv3.Commands, error) {
	log.Debugf("### ProfileTaskServerV3:ReportTaskFinish ProfileTaskFinishReport: %#v", reporter)

	return &commonv3.Commands{}, nil
}

type ManagementServerV3 struct {
	mgmtv3.UnimplementedManagementServiceServer
}

func (*ManagementServerV3) ReportInstanceProperties(ctx context.Context, mgmt *mgmtv3.InstanceProperties) (*commonv3.Commands, error) {
	log.Debugf("### ManagementServerV3:ReportInstanceProperties InstanceProperties: %#v", mgmt)

	return &commonv3.Commands{}, nil
}

func (*ManagementServerV3) KeepAlive(ctx context.Context, ping *mgmtv3.InstancePingPkg) (*commonv3.Commands, error) {
	log.Debugf("### ManagementServerV3:KeepAlive InstancePingPkg: %#v", ping)

	return &commonv3.Commands{}, nil
}

type DiscoveryServerV3 struct {
	configv3.UnimplementedConfigurationDiscoveryServiceServer
}

func (*DiscoveryServerV3) FetchConfigurations(ctx context.Context, cfgReq *configv3.ConfigurationSyncRequest) (*commonv3.Commands, error) {
	log.Debugf("### DiscoveryServerV3:FetchConfigurations ConfigurationSyncRequest: %#v", cfgReq)

	return &commonv3.Commands{}, nil
}