// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rabbitmq

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

var (
	inputName            = `rabbitmq`
	customObjectFeedName = inputName + "/CO"
	l                    = logger.DefaultSLogger(inputName)
	minInterval          = time.Second
	maxInterval          = time.Second * 30
	sample               = `
[[inputs.rabbitmq]]
  # rabbitmq url ,required
  url = "http://localhost:15672"

  # rabbitmq user, required
  username = "guest"

  # rabbitmq password, required
  password = "guest"

  # ##(optional) collection interval, default is 30s
  # interval = "30s"

  ## Optional TLS Config
  # tls_ca = "/xxx/ca.pem"
  # tls_cert = "/xxx/cert.cer"
  # tls_key = "/xxx/key.key"
  ## Use TLS but skip chain & host verification
  insecure_skip_verify = false

  ## Set true to enable election
  election = true

  # [inputs.rabbitmq.log]
  # files = []
  # #grok pipeline script path
  # pipeline = "rabbitmq.p"

  [inputs.rabbitmq.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...

`
	pipelineCfg = `
grok(_, "%{LOGLEVEL:status}%{DATA}====%{SPACE}%{DATA:time}%{SPACE}===%{SPACE}%{GREEDYDATA:msg}")

grok(_, "%{DATA:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}")

default_time(time)
`
)

const (
	OverviewMetric = "rabbitmq_overview"
	ExchangeMetric = "rabbitmq_exchange"
	NodeMetric     = "rabbitmq_node"
	QueueMetric    = "rabbitmq_queue"
)

type Input struct {
	URL      string            `toml:"url"`
	Username string            `toml:"username"`
	Password string            `toml:"password"`
	Interval datakit.Duration  `toml:"interval"`
	Log      *rabbitmqlog      `toml:"log"`
	Tags     map[string]string `toml:"tags"`

	Version            string
	Uptime             int
	CollectCoStatus    string
	CollectCoErrMsg    string
	LastCustomerObject *customerObjectMeasurement

	QueueNameIncludeDeprecated []string `toml:"queue_name_include,omitempty"`
	QueueNameExcludeDeprecated []string `toml:"queue_name_exclude,omitempty"`

	tls.ClientConfig

	// HTTP client
	client *http.Client
	host   string

	tail    *tailer.Tailer
	lastErr error
	start   time.Time
	alignTS int64

	Election     bool `toml:"election"`
	pause        bool
	pauseCh      chan bool
	lock         sync.Mutex
	collectCache []*point.Point

	semStop *cliutils.Sem // start stop signal
	feeder  dkio.Feeder
	Tagger  datakit.GlobalTagger

	UpState int
}

type rabbitmqlog struct {
	Files             []string `toml:"files"`
	Pipeline          string   `toml:"pipeline"`
	IgnoreStatus      []string `toml:"ignore"`
	CharacterEncoding string   `toml:"character_encoding"`
	MultilineMatch    string   `toml:"multiline_match"`
}

type OverviewResponse struct {
	Version      string        `json:"rabbitmq_version"`
	ClusterName  string        `json:"cluster_name"`
	MessageStats *MessageStats `json:"message_stats"`
	ObjectTotals *ObjectTotals `json:"object_totals"`
	QueueTotals  *QueueTotals  `json:"queue_totals"`
	Listeners    []Listeners   `json:"listeners"`
}

type Listeners struct {
	Protocol string `json:"protocol"`
}

// Details ...
type Details struct {
	Rate float64 `json:"rate"`
}

// MessageStats ...
type MessageStats struct {
	Ack                     int64
	AckDetails              Details `json:"ack_details"`
	Confirm                 int64   `json:"confirm"`
	ConfirmDetail           Details `json:"ack_details_details"`
	Deliver                 int64
	DeliverDetails          Details `json:"deliver_details"`
	DeliverGet              int64   `json:"deliver_get"`
	DeliverGetDetails       Details `json:"deliver_get_details"`
	Publish                 int64
	PublishDetails          Details `json:"publish_details"`
	Redeliver               int64
	RedeliverDetails        Details `json:"redeliver_details"`
	PublishIn               int64   `json:"publish_in"`
	PublishInDetails        Details `json:"publish_in_details"`
	PublishOut              int64   `json:"publish_out"`
	PublishOutDetails       Details `json:"publish_out_details"`
	ReturnUnroutable        int64   `json:"return_unroutable"`
	ReturnUnroutableDetails Details `json:"return_unroutable_details"`
}

// ObjectTotals ...
type ObjectTotals struct {
	Channels    int64
	Connections int64
	Consumers   int64
	Exchanges   int64
	Queues      int64
}

type QueueTotals struct {
	Messages       int64
	MessagesDetail Details `json:"messages_details"`

	MessagesReady       int64   `json:"messages_ready"`
	MessagesReadyDetail Details `json:"messages_ready_details"`

	MessagesUnacknowledged       int64   `json:"messages_unacknowledged"`
	MessagesUnacknowledgedDetail Details `json:"messages_unacknowledged_details"`
}

type Exchange struct {
	Name         string
	MessageStats `json:"message_stats"`
	Type         string
	Internal     bool
	Vhost        string
	Durable      bool
	AutoDelete   bool `json:"auto_delete"`
}

type Node struct {
	Name              string
	DiskFreeAlarm     bool  `json:"disk_free_alarm"`
	MemAlarm          bool  `json:"mem_alarm"`
	Running           bool  `json:"running"`
	DiskFree          int64 `json:"disk_free"`
	DiskFreeLimit     int64 `json:"disk_free_limit"`
	FdTotal           int64 `json:"fd_total"`
	FdUsed            int64 `json:"fd_used"`
	MemLimit          int64 `json:"mem_limit"`
	MemUsed           int64 `json:"mem_used"`
	ProcTotal         int64 `json:"proc_total"`
	ProcUsed          int64 `json:"proc_used"`
	RunQueue          int64 `json:"run_queue"`
	SocketsTotal      int64 `json:"sockets_total"`
	SocketsUsed       int64 `json:"sockets_used"`
	Uptime            int64 `json:"uptime"`
	MnesiaDiskTxCount int64 `json:"mnesia_disk_tx_count"`
	MnesiaRAMTxCount  int64 `json:"mnesia_ram_tx_count"`
	GcNum             int64 `json:"gc_num"`
	IoWriteBytes      int64 `json:"io_write_bytes"`
	IoReadBytes       int64 `json:"io_read_bytes"`
	GcBytesReclaimed  int64 `json:"gc_bytes_reclaimed"`

	IoWriteAvgTime float64 `json:"io_write_avg_time"`
	IoReadAvgTime  float64 `json:"io_read_avg_time"`
	IoSeekAvgTime  float64 `json:"io_seek_avg_time"`
	IoSyncAvgTime  float64 `json:"io_sync_avg_time"`

	GcNumDetails             Details `json:"gc_num_details"`
	MnesiaRAMTxCountDetails  Details `json:"mnesia_ram_tx_count_details"`
	MnesiaDiskTxCountDetails Details `json:"mnesia_disk_tx_count_details"`
	GcBytesReclaimedDetails  Details `json:"gc_bytes_reclaimed_details"`
	IoReadAvgTimeDetails     Details `json:"io_read_avg_time_details"`
	IoReadBytesDetails       Details `json:"io_read_bytes_details"`
	IoWriteAvgTimeDetails    Details `json:"io_write_avg_time_details"`
	IoWriteBytesDetails      Details `json:"io_write_bytes_details"`
}

type Queue struct {
	QueueTotals          // just to not repeat the same code
	MessageStats         `json:"message_stats"`
	Memory               int64   `json:"memory"`
	Consumers            int64   `json:"consumers"`
	ConsumerUtilisation  float64 `json:"consumer_utilisation"` //nolint:misspell
	HeadMessageTimestamp int64   `json:"head_message_timestamp"`
	Name                 string
	Node                 string
	Vhost                string
	Durable              bool
	AutoDelete           bool   `json:"auto_delete"`
	IdleSince            string `json:"idle_since"`
}

func (ipt *Input) createHTTPClient() (*http.Client, error) {
	tlsCfg, err := ipt.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: time.Second * 10,
	}

	return client, nil
}

func (ipt *Input) requestJSON(u string, target interface{}) error {
	u = fmt.Sprintf("%s%s", ipt.URL, u)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(ipt.Username, ipt.Password)
	resp, err := ipt.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck
	return json.NewDecoder(resp.Body).Decode(target)
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

func newOtherFieldInfo(datatype, ftype, unit, desc string) *inputs.FieldInfo { //nolint:unparam
	return &inputs.FieldInfo{
		DataType: datatype,
		Type:     ftype,
		Unit:     unit,
		Desc:     desc,
	}
}

func newByteFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     desc,
	}
}

func (ipt *Input) metricAppend(metric *point.Point) {
	ipt.lock.Lock()
	ipt.collectCache = append(ipt.collectCache, metric)
	ipt.lock.Unlock()
}
