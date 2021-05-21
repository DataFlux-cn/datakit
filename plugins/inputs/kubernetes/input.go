package kubernetes

import (
	"context"
	"k8s.io/client-go/rest"
	"sync"
	"time"

	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	maxInterval = 30 * time.Minute
	minInterval = 15 * time.Second
)

var (
	inputName   = "kubernetes"
	catalogName = "kubernetes"
	l           = logger.DefaultSLogger("kubernetes")
)

var availableCollectors = map[string]func(ctx context.Context, i *Input){
	// "daemonsets":             collectDaemonSets,
	// "deployments":            collectDeployments,
	// "endpoints":              collectEndpoints,
	// "ingress":                collectIngress,
	// "nodes":                  collectNodes,
	// "pods":                   collectPods,
	// "services":               collectServices,
	// "statefulsets":           collectStatefulSets,
	// "persistentvolumes":      collectPersistentVolumes,
	// "persistentvolumeclaims": collectPersistentVolumeClaims,
}

type Input struct {
	Service                string `toml:"service"`
	Interval               datakit.Duration
	ObjectInterval         string               `toml:"object_Interval"`
	ObjectIntervalDuration time.Duration        `toml:"_"`
	Tags                   map[string]string    `toml:"tags"`
	collectCache           []inputs.Measurement `toml:"-"`
	collectObjectCache     []inputs.Measurement `toml:"-"`
	lastErr                error

	KubeConfigPath    string `toml:"kube_config_path"`
	URL               string `toml:"url"`
	BearerToken       string `toml:"bearer_token"`
	BearerTokenString string `toml:"bearer_token_string"`
	Namespace         string `toml:"namespace"`
	Timeout           string `toml:"timeout"`
	TimeoutDuration   time.Duration
	ResourceExclude   []string `toml:"resource_exclude"`
	ResourceInclude   []string `toml:"resource_include"`

	SelectorInclude []string `toml:"selector_include"`
	SelectorExclude []string `toml:"selector_exclude"`

	tls.ClientConfig
	client         *client
	mu             sync.Mutex
	selectorFilter filter.Filter
}

func (i *Input) initCfg() error {
	TimeoutDuration, err := time.ParseDuration(i.Timeout)
	if err != nil {
		TimeoutDuration = 5 * time.Second
	}

	var (
		config *rest.Config
	)

	i.TimeoutDuration = TimeoutDuration

	if i.KubeConfigPath != "" {
		config, err = createConfigByKubePath(i.KubeConfigPath)
		if err != nil {
			return err
		}
	} else if i.URL != "" {
		config, err = createConfigByToken(i.URL, i.BearerTokenString, i.TLSCA, i.InsecureSkipVerify)
	}

	i.client, err = newClient(config, 5*time.Second)
	if err != nil {
		return err
	}

	i.globalTag()

	return nil
}

func (i *Input) globalTag() {
	i.Tags["url"] = i.URL
	i.Tags["service_name"] = i.Service
}

func (i *Input) Collect() error {
	i.collectCache = []inputs.Measurement{}

	ctx := context.Background()
	wg := sync.WaitGroup{}

	resourceFilter, err := filter.NewIncludeExcludeFilter(i.ResourceInclude, i.ResourceExclude)
	if err != nil {
		return err
	}

	i.selectorFilter, err = filter.NewIncludeExcludeFilter(i.SelectorInclude, i.SelectorExclude)
	if err != nil {
		return err
	}

	for collector, f := range availableCollectors {
		if resourceFilter.Match(collector) {
			wg.Add(1)
			go func(f func(ctx context.Context, i *Input)) {
				defer wg.Done()
				f(ctx, i)
			}(f)
		}
	}

	wg.Wait()

	if i.lastErr != nil {
		io.FeedLastError(inputName, i.lastErr.Error())
		i.lastErr = nil
	}

	return nil
}

func (i *Input) Run() {
	l = logger.SLogger("kubernetes")
	i.Interval.Duration = datakit.ProtectedInterval(minInterval, maxInterval, i.Interval.Duration)

	i.ObjectIntervalDuration = 5 * time.Minute
	i.ObjectIntervalDuration, _ = time.ParseDuration(i.ObjectInterval)

	i.lastErr = i.initCfg()

	if i.lastErr != nil {
		io.FeedLastError(inputName, i.lastErr.Error())
		i.lastErr = nil
		return
	}

	tick := time.NewTicker(i.Interval.Duration)
	defer tick.Stop()

	objectTick := time.NewTicker(i.ObjectIntervalDuration)
	defer objectTick.Stop()

	go func() {
		start := time.Now()
		if err := i.CollectObject(); err != nil {
			io.FeedLastError(inputName, err.Error())
		} else {
			if err := inputs.FeedMeasurement(inputName, datakit.Object, i.collectObjectCache,
				&io.Option{CollectCost: time.Since(start)}); err != nil {
				io.FeedLastError(inputName, err.Error())
			}
			i.collectObjectCache = i.collectObjectCache[:0]
		}
	}()

	n := 0
	for {
		n++
		select {
		case <-tick.C:
			l.Debugf("kubernetes metric input gathering...")
			start := time.Now()
			if err := i.Collect(); err != nil {
				io.FeedLastError(inputName, err.Error())
			} else {
				if err := inputs.FeedMeasurement(inputName, datakit.Metric, i.collectCache,
					&io.Option{CollectCost: time.Since(start), HighFreq: (n%2 == 0)}); err != nil {
					io.FeedLastError(inputName, err.Error())
				}

				i.collectCache = i.collectCache[:0] // NOTE: do not forget to clean cache
			}
		case <-objectTick.C:
			l.Debugf("kubernetes object input gathering...")
			start := time.Now()
			if err := i.CollectObject(); err != nil {
				io.FeedLastError(inputName, err.Error())
			} else {
				if err := inputs.FeedMeasurement(inputName, datakit.Object, i.collectObjectCache,
					&io.Option{CollectCost: time.Since(start)}); err != nil {
					io.FeedLastError(inputName, err.Error())
				}
				i.collectObjectCache = i.collectObjectCache[:0]
			}
		case <-datakit.Exit.Wait():
			l.Info("kubernetes exit")
			return
		}
	}
}

func (i *Input) Catalog() string { return catalogName }

func (i *Input) SampleConfig() string { return configSample }

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
