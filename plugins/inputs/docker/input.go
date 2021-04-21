package docker

import (
	"context"
	"crypto/tls"
	"sync"
	"time"

	"github.com/docker/docker/api/types"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newInput()
	})
}

type Input struct {
	Endpoint              string            `toml:"endpoint"`
	CollectMetric         bool              `toml:"collect_metric"`
	CollectObject         bool              `toml:"collect_object"`
	CollectLogging        bool              `toml:"collect_logging"`
	CollectMetricInterval string            `toml:"collect_metric_interval"`
	CollectObjectInterval string            `toml:"collect_object_interval"`
	IncludeExited         bool              `toml:"include_exited"`
	ClientConfig                            // tls config
	LogOption             []*LogOption      `toml:"log_option"`
	Tags                  map[string]string `toml:"tags"`

	collectMetricDuration time.Duration
	collectObjectDuration time.Duration
	timeoutDuration       time.Duration

	newEnvClient         func() (Client, error)
	newClient            func(string, *tls.Config) (Client, error)
	containerLogsOptions types.ContainerLogsOptions
	containerLogList     map[string]context.CancelFunc

	client     Client
	kubernetes *Kubernetes

	opts types.ContainerListOptions
	wg   sync.WaitGroup
	mu   sync.Mutex
}

func newInput() *Input {
	return &Input{
		Endpoint:              defaultEndpoint,
		Tags:                  make(map[string]string),
		newEnvClient:          NewEnvClient,
		newClient:             NewClient,
		collectMetricDuration: minimumCollectMetricDuration,
		collectObjectDuration: minimumCollectObjectDuration,
		timeoutDuration:       defaultAPITimeout,
		containerLogList:      make(map[string]context.CancelFunc),
	}
}

func (*Input) SampleConfig() string {
	return sampleCfg
}

func (*Input) Catalog() string {
	return "docker"
}

func (*Input) PipelineConfig() map[string]string {
	return nil
}

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLinux}
}

func (this *Input) Run() {
	l = logger.SLogger(inputName)

	if this.initCfg() {
		return
	}
	l.Info("docker input start")

	if this.CollectMetric {
		go this.gatherMetric(this.collectMetricDuration)
	}

	if this.CollectObject {
		go this.gatherObject(this.collectObjectDuration)
	}

	if this.CollectLogging {
		// 共用同一个interval
		go this.gatherLoggoing(this.collectMetricDuration)
	}

	l.Info("docker exit success")
}

func (this *Input) initCfg() bool {
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		if err := this.loadCfg(); err != nil {
			l.Error(err)
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	return false
}

func (n *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&dockerContainersMeasurement{},
		&dockerContainersLogMeasurement{},
	}
}

// type dockerMeasurement struct {
// 	name   string
// 	tags   map[string]string
// 	fields map[string]interface{}
// 	ts     time.Time
// }
//
// type dockerLogMeasurement struct {
// 	name   string
// 	tags   map[string]string
// 	fields map[string]interface{}
// 	ts     time.Time
// }

const (
	dockerContainersName = "docker_containers"
	ker
)

type dockerContainersMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (this *dockerContainersMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(this.name, this.tags, this.fields, this.ts)
}

func (this *dockerContainersMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: dockerContainersName,
		Desc: "Docker 容器相关",
		Tags: map[string]interface{}{
			"container_id":   inputs.NewTagInfo(`容器 ID`),
			"container_name": inputs.NewTagInfo(`容器名称`),
			"images_name":    inputs.NewTagInfo(`容器镜像名称`),
			"docker_image":   inputs.NewTagInfo(`镜像名称+版本号`),
			"name":           inputs.NewTagInfo(`对象数据的指定 ID，（仅在对象数据中存在）`),
			"container_host": inputs.NewTagInfo(`容器内部的主机名（仅在对象数据中存在）`),
			"host":           inputs.NewTagInfo(`容器宿主机的主机名`),
			"stats":          inputs.NewTagInfo(`运行状态，running/exited/removed`),
			"pod_name":       inputs.NewTagInfo(`pod名称`),
			"pod_namesapce":  inputs.NewTagInfo(`pod命名空间`),
			// "kube_container_name": inputs.NewTagInfo(`TODO`),
			// "kube_daemon_set":     inputs.NewTagInfo(`TODO`),
			// "kube_deployment":     inputs.NewTagInfo(`TODO`),
			// "kube_namespace":      inputs.NewTagInfo(`TODO`),
			// "kube_ownerref_kind":  inputs.NewTagInfo(`TODO`),
		},
		Fields: map[string]interface{}{
			"from_kubernetes":    &inputs.FieldInfo{DataType: inputs.Bool, Unit: inputs.UnknownUnit, Desc: "该容器是否由 Kubernetes 创建"},
			"cpu_usage_percent":  &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "CPU使用率"},
			"mem_usage_percent":  &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "内存使用率"},
			"cpu_delta":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeIByte, Desc: "容器 CPU 增量"},
			"cpu_system_delta":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeIByte, Desc: "系统 CPU 增量"},
			"cpu_numbers":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "CPU 核心数"},
			"mem_available":      &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeIByte, Desc: "内存可用总量"},
			"mem_usage":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeIByte, Desc: "内存使用量"},
			"mem_failed_count":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeIByte, Desc: "内存分配失败的次数"},
			"network_bytes_rcvd": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeIByte, Desc: "从网络接收到的总字节数"},
			"network_bytes_sent": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeIByte, Desc: "向网络发送出的总字节数"},
			"block_read_byte":    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeIByte, Desc: "从容器文件系统读取的总字节数"},
			"block_write_byte":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeIByte, Desc: "向容器文件系统写入的总字节数"},
		},
	}
}

type dockerContainersLogMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (this *dockerContainersLogMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(this.name, this.tags, this.fields, this.ts)
}

func (this *dockerContainersLogMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "Docker 日志指标",
		Desc: "默认使用容器名，如果容器名能匹配 `log_option.container_name_match` 正则，则使用对应的 `source` 字段值",
		Tags: map[string]interface{}{
			"container_name": inputs.NewTagInfo(`容器名称`),
			"image_name":     inputs.NewTagInfo(`容器镜像名称`),
			"stream":         inputs.NewTagInfo(`数据流方式，stdout/stderr/tty`),
		},
		Fields: map[string]interface{}{
			"from_kubernetes": &inputs.FieldInfo{DataType: inputs.Bool, Unit: inputs.UnknownUnit, Desc: "该容器是否由 Kubernetes 创建"},
			"service":         &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "服务名称"},
			"status":          &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "日志状态，info/emerg/alert/critical/error/warning/debug/OK"},
			"message":         &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "日志源数据"},
		},
	}
}
