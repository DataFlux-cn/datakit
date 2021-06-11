package jvm

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	defaultInterval   = "60s"
	MaxGatherInterval = 30 * time.Minute
	MinGatherInterval = 1 * time.Second
	inputName         = "jvm"
)

const (
	JvmConfigSample = `[[inputs.jvm]]
  # default_tag_prefix      = ""
  # default_field_prefix    = ""
  # default_field_separator = "."

  # username = ""
  # password = ""
  # response_timeout = "5s"

  ## Optional TLS config
  # tls_ca   = "/var/private/ca.pem"
  # tls_cert = "/var/private/client.pem"
  # tls_key  = "/var/private/client-key.pem"
  # insecure_skip_verify = false

  ## Monitor Intreval
  # interval   = "60s"

  # Add agents URLs to query
  urls = ["http://localhost:8080/jolokia"]

  ## Add metrics to read
  [[inputs.jvm.metric]]
    name  = "java_runtime"
    mbean = "java.lang:type=Runtime"
    paths = ["Uptime"]

  [[inputs.jvm.metric]]
    name  = "java_memory"
    mbean = "java.lang:type=Memory"
    paths = ["HeapMemoryUsage", "NonHeapMemoryUsage", "ObjectPendingFinalizationCount"]

  [[inputs.jvm.metric]]
    name     = "java_garbage_collector"
    mbean    = "java.lang:name=*,type=GarbageCollector"
    paths    = ["CollectionTime", "CollectionCount"]
    tag_keys = ["name"]

  [[inputs.jvm.metric]]
    name  = "java_threading"
    mbean = "java.lang:type=Threading"
    paths = ["TotalStartedThreadCount", "ThreadCount", "DaemonThreadCount", "PeakThreadCount"]

  [[inputs.jvm.metric]]
    name  = "java_class_loading"
    mbean = "java.lang:type=ClassLoading"
    paths = ["LoadedClassCount", "UnloadedClassCount", "TotalLoadedClassCount"]

  [[inputs.jvm.metric]]
    name     = "java_memory_pool"
    mbean    = "java.lang:name=*,type=MemoryPool"
    paths    = ["Usage", "PeakUsage", "CollectionUsage"]
    tag_keys = ["name"]

  [inputs.jvm.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...`
)

var JvmTypeMap = map[string]string{
	"Uptime":                         "int",
	"HeapMemoryUsageinit":            "int",
	"HeapMemoryUsageused":            "int",
	"HeapMemoryUsagemax":             "int",
	"HeapMemoryUsagecommitted":       "int",
	"NonHeapMemoryUsageinit":         "int",
	"NonHeapMemoryUsageused":         "int",
	"NonHeapMemoryUsagemax":          "int",
	"NonHeapMemoryUsagecommitted":    "int",
	"ObjectPendingFinalizationCount": "int",
	"CollectionTime":                 "int",
	"CollectionCount":                "int",
	"DaemonThreadCount":              "int",
	"PeakThreadCount":                "int",
	"ThreadCount":                    "int",
	"TotalStartedThreadCount":        "int",
	"LoadedClassCount":               "int",
	"TotalLoadedClassCount":          "int",
	"UnloadedClassCount":             "int",
	"Usageinit":                      "int",
	"Usagemax":                       "int",
	"Usagecommitted":                 "int",
	"Usageused":                      "int",
	"PeakUsageinit":                  "int",
	"PeakUsagemax":                   "int",
	"PeakUsagecommitted":             "int",
	"PeakUsageused":                  "int",
}

type Input struct {
	inputs.JolokiaAgent
	Tags map[string]string `toml:"tags"`
}

var (
	log = logger.DefaultSLogger(inputName)
)

func (i *Input) Run() {
	log = logger.SLogger(inputName)
	if i.Interval == "" {
		i.Interval = defaultInterval
	}

	i.PluginName = inputName

	i.JolokiaAgent.Tags = i.Tags
	i.JolokiaAgent.Types = JvmTypeMap
	i.JolokiaAgent.L = log
	i.JolokiaAgent.Collect()
}

func (i *Input) Catalog() string      { return inputName }
func (i *Input) SampleConfig() string { return JvmConfigSample }
func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&JavaRuntimeMemt{},
		&JavaMemoryMemt{},
		&JavaGcMemt{},
		//&JavaLastGcMemt{},
		&JavaThreadMemt{},
		&JavaClassLoadMemt{},
		&JavaMemoryPoolMemt{},
	}
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
