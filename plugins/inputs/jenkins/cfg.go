package jenkins

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	"github.com/influxdata/telegraf/plugins/common/tls"

	"net/http"
	"sync"
	"time"
)

var (
	inputName   = `jenkins`
	l           = logger.DefaultSLogger(inputName)
	minInterval = time.Second
	maxInterval = time.Second * 30
	sample      = `
[[inputs.jenkins]]
	## The Jenkins URL in the format "schema://host:port",required
	url = "http://my-jenkins-instance:8080"
	## Metric Access Key ,generate in your-jenkins-host:/configure,required
	key = ""
	## Set response_timeout
	# response_timeout = "5s"
	
	## Optional TLS Config
	# tls_ca = "/xx/ca.pem"
	# tls_cert = "/xx/cert.pem"
	# tls_key = "/xx/key.pem"
	## Use SSL but skip chain & host verification
	# insecure_skip_verify = false
	[inputs.jenkins.log]
	#	files = []
	#	# grok pipeline script path
	#	pipeline = "jenkins.p"
	[inputs.jenkins.tags]
	# a = "b"`

	pipelineCfg = `
grok(_, "%{TIMESTAMP_ISO8601:time} \\[id=%{GREEDYDATA:id}\\]\t%{GREEDYDATA:status}\t")
default_time(time)
group_in(status, ["WARNING", "NOTICE"], "warning")
group_in(status, ["SEVERE", "ERROR"], "error")
group_in(status, ["INFO"], "info")

`
)

type Input struct {
	Url             string               `toml:"url"`
	Key             string               `toml:"key"`
	Interval        datakit.Duration     `toml:"interval"`
	ResponseTimeout datakit.Duration     `toml:"response_timeout"`
	Log             *inputs.TailerOption `toml:"log"`
	Tags            map[string]string    `toml:"tags"`

	tls.ClientConfig
	// HTTP client
	client *http.Client

	start time.Time
	tail  *inputs.Tailer

	lastErr      error
	wg           sync.WaitGroup
	collectCache []inputs.Measurement
}

func newCountFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}

func newRateFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     desc,
	}
}

func newByteFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeIByte,
		Desc:     desc,
	}
}
