package container

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"regexp"
	"sync"
	"time"

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
	Endpoint       string            `toml:"endpoint"`
	EnableMetric   bool              `toml:"enable_metric"`
	EnableObject   bool              `toml:"enable_object"`
	EnableLogging  bool              `toml:"enable_logging"`
	MetricInterval string            `toml:"metric_interval"`
	Kubernetes     *Kubernetes       `toml:"kubelet"`
	ClientConfig                     // tls config
	LogFilters     LogFilters        `toml:"logfilter"`
	Tags           map[string]string `toml:"tags"`

	DropTags            []string `toml:"drop_tags"`
	IgnoreImageName     []string `toml:"ignore_image_name"`
	IgnoreContainerName []string `toml:"ignore_container_name"`

	newClient func(string, *tls.Config) (Client, error)

	metricDuration   time.Duration
	containerLogList map[string]context.CancelFunc

	ignoreImageNameRegexps     []*regexp.Regexp
	ignoreContainerNameRegexps []*regexp.Regexp

	client Client

	wg sync.WaitGroup
	mu sync.Mutex
}

func newInput() *Input {
	return &Input{
		Endpoint:         dockerEndpoint,
		Tags:             make(map[string]string),
		containerLogList: make(map[string]context.CancelFunc),
		newClient:        NewClient,
		metricDuration:   minMetricDuration,
	}
}

func (*Input) SampleConfig() string {
	return sampleCfg
}

func (*Input) Catalog() string {
	return "container"
}

func (*Input) PipelineConfig() map[string]string {
	return nil
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&containersMeasurement{},
		&containersLogMeasurement{},
		&kubeletPodMeasurement{},
		// &kubeletNodeMeasurement{},
	}
}

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLinux}
}

// TODO
func (*Input) RunPipeline() {
}

func (this *Input) Run() {
	l = logger.SLogger(inputName)

	if this.initCfg() {
		return
	}
	l.Info("container input start")

	if this.EnableObject {
		if this.Kubernetes != nil {
			do(func() ([]*io.Point, error) {
				return this.Kubernetes.GatherPodMetrics(this.DropTags, "object")
			}, datakit.Object, "k8s pod object gather failed")
		}

		do(func() ([]*io.Point, error) { return this.gather(objectCategory) }, datakit.Object, "container object gather failed")
	}

	if this.EnableLogging {
		this.gatherLog()
	}

	metricsTick := time.NewTicker(this.metricDuration)
	defer metricsTick.Stop()

	objectTick := time.NewTicker(objectDuration)
	defer objectTick.Stop()

	loggingTick := time.NewTicker(loggingHitDuration)
	defer loggingTick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			// clean logging
			this.cancelTails()
			this.wg.Wait()

			l.Info("container exit success")
			return

		case <-metricsTick.C:
			if this.Kubernetes != nil {
				do(func() ([]*io.Point, error) {
					return this.Kubernetes.GatherPodMetrics(this.DropTags, "metric")
				}, datakit.Metric, "k8s pod metrics gather failed")
			}
			if this.EnableMetric {
				do(func() ([]*io.Point, error) { return this.gather(metricCategory) }, datakit.Metric, "container metrics gather failed")
			}

		case <-objectTick.C:
			if this.EnableObject {
				if this.Kubernetes != nil {
					do(func() ([]*io.Point, error) {
						return this.Kubernetes.GatherPodMetrics(this.DropTags, "object")
					}, datakit.Object, "k8s pod object gather failed")
				}

				do(func() ([]*io.Point, error) { return this.gather(objectCategory) }, datakit.Object, "container object gather failed")
			}

		case <-loggingTick.C:
			if this.EnableLogging {
				this.gatherLog()
			}
		}
	}
}

func (this *Input) initCfg() bool {
	// 如果配置文件中使用默认 endpoint 且该文件不存在，说明其没有安装 docker（经测试，docker service 停止后，sock 文件依然存在）
	// 此行为是为了应对 default_enabled_inputs 行为，避免在没有安装 docker 的主机上开启 docker，然后无限 error
	if this.Endpoint == dockerEndpoint {
		if _, err := os.Stat(dockerEndpointPath); os.IsNotExist(err) {
			l.Infof("check defaultEndpoint: %s is not exist, maybe docker.service is not installed, exit", this.Endpoint)
			// 预料之中的退出
			return true
		}
	}

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
			io.FeedLastError(inputName, fmt.Sprintf("load config: %s", err.Error()))
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	return false
}

func do(gatherFn func() ([]*io.Point, error), category, prefixlog string) {
	startTime := time.Now()
	pts, err := gatherFn()
	if err != nil {
		l.Error(err)
		io.FeedLastError(inputName, fmt.Sprintf("%s: %s", prefixlog, err))
		return
	}
	cost := time.Since(startTime)
	if err := io.Feed(inputName, category, pts, &io.Option{CollectCost: cost}); err != nil {
		l.Error(err)
		io.FeedLastError(inputName, fmt.Sprintf("%s: %s", prefixlog, err))
	}

}
