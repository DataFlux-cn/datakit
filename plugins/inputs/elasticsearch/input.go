// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package elasticsearch Collect ElasticSearch metrics.
package elasticsearch

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	internalIo "io"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var _ inputs.ElectionInput = (*Input)(nil)

var mask = regexp.MustCompile(`https?:\/\/\S+:\S+@`)

const (
	statsPath      = "/_nodes/stats"
	statsPathLocal = "/_nodes/_local/stats"
)

type nodeStat struct {
	Host       string            `json:"host"`
	Name       string            `json:"name"`
	Roles      []string          `json:"roles"`
	Attributes map[string]string `json:"attributes"`
	Indices    interface{}       `json:"indices"`
	OS         interface{}       `json:"os"`
	Process    interface{}       `json:"process"`
	JVM        interface{}       `json:"jvm"`
	ThreadPool interface{}       `json:"thread_pool"`
	FS         interface{}       `json:"fs"`
	Transport  interface{}       `json:"transport"`
	HTTP       interface{}       `json:"http"`
	Breakers   interface{}       `json:"breakers"`
}

type clusterHealth struct {
	ActivePrimaryShards         int                    `json:"active_primary_shards"`
	ActiveShards                int                    `json:"active_shards"`
	ActiveShardsPercentAsNumber float64                `json:"active_shards_percent_as_number"`
	ClusterName                 string                 `json:"cluster_name"`
	DelayedUnassignedShards     int                    `json:"delayed_unassigned_shards"`
	InitializingShards          int                    `json:"initializing_shards"`
	NumberOfDataNodes           int                    `json:"number_of_data_nodes"`
	NumberOfInFlightFetch       int                    `json:"number_of_in_flight_fetch"`
	NumberOfNodes               int                    `json:"number_of_nodes"`
	NumberOfPendingTasks        int                    `json:"number_of_pending_tasks"`
	RelocatingShards            int                    `json:"relocating_shards"`
	Status                      string                 `json:"status"`
	TaskMaxWaitingInQueueMillis int                    `json:"task_max_waiting_in_queue_millis"`
	TimedOut                    bool                   `json:"timed_out"`
	UnassignedShards            int                    `json:"unassigned_shards"`
	Indices                     map[string]indexHealth `json:"indices"`
}

type indexState struct {
	Indices map[string]struct {
		Managed bool   `json:"managed"`
		Step    string `json:"step"`
	} `json:"indices"`
}

type indexHealth struct {
	ActivePrimaryShards int    `json:"active_primary_shards"`
	ActiveShards        int    `json:"active_shards"`
	InitializingShards  int    `json:"initializing_shards"`
	NumberOfReplicas    int    `json:"number_of_replicas"`
	NumberOfShards      int    `json:"number_of_shards"`
	RelocatingShards    int    `json:"relocating_shards"`
	Status              string `json:"status"`
	UnassignedShards    int    `json:"unassigned_shards"`
}

type clusterStats struct {
	NodeName    string      `json:"node_name"`
	ClusterName string      `json:"cluster_name"`
	Status      string      `json:"status"`
	Indices     interface{} `json:"indices"`
	Nodes       interface{} `json:"nodes"`
}

type indexStat struct {
	Primaries interface{}              `json:"primaries"`
	Total     interface{}              `json:"total"`
	Shards    map[string][]interface{} `json:"shards"`
}

//nolint:lll
const sampleConfig = `
[[inputs.elasticsearch]]
  ## Elasticsearch 服务器配置
  # 支持 Basic 认证
  # servers = ["http://user:pass@localhost:9200"]
  servers = ["http://localhost:9200"]

  ## 采集间隔
  # 单位 "ns", "us" (or "µs"), "ms", "s", "m", "h"
  interval = "10s"

  ## HTTP 超时设置
  http_timeout = "5s"

  ## 发行版本：elasticsearch/opendistro/opensearch
  distribution = "elasticsearch"

  ## 默认 local 是开启的，只采集当前 Node 自身指标，如果需要采集集群所有 Node，需要将 local 设置为 false
  local = true

  ## 设置为 true 可以采集 cluster health
  cluster_health = false

  ## cluster health level 设置，indices (默认) 和 cluster
  # cluster_health_level = "indices"

  ## 设置为 true 时可以采集 cluster stats.
  cluster_stats = false

  ## 只从 master Node 获取 cluster_stats，这个前提是需要设置 local = true
  cluster_stats_only_from_master = true

  ## 需要采集的 Indices, 默认为 _all
  indices_include = ["_all"]

  ## indices 级别，可取值：shards/cluster/indices
  indices_level = "shards"

  ## node_stats 可支持配置选项有 indices/os/process/jvm/thread_pool/fs/transport/http/breaker
  # 默认是所有
  # node_stats = ["jvm", "http"]

  ## HTTP Basic Authentication 用户名和密码
  # username = ""
  # password = ""

  ## TLS Config
  tls_open = false
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Set true to enable election
  election = true

  # [inputs.elasticsearch.log]
  # files = []
  # #grok pipeline script path
  # pipeline = "elasticsearch.p"

  [inputs.elasticsearch.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`

//nolint:lll
const pipelineCfg = `
# Elasticsearch_search_query
grok(_, "^\\[%{TIMESTAMP_ISO8601:time}\\]\\[%{LOGLEVEL:status}%{SPACE}\\]\\[i.s.s.(query|fetch)%{SPACE}\\] (\\[%{HOSTNAME:nodeId}\\] )?\\[%{NOTSPACE:index}\\]\\[%{INT}\\] took\\[.*\\], took_millis\\[%{INT:duration}\\].*")

# Elasticsearch_slow_indexing
grok(_, "^\\[%{TIMESTAMP_ISO8601:time}\\]\\[%{LOGLEVEL:status}%{SPACE}\\]\\[i.i.s.index%{SPACE}\\] (\\[%{HOSTNAME:nodeId}\\] )?\\[%{NOTSPACE:index}/%{NOTSPACE}\\] took\\[.*\\], took_millis\\[%{INT:duration}\\].*")

# Elasticsearch_default
grok(_, "^\\[%{TIMESTAMP_ISO8601:time}\\]\\[%{LOGLEVEL:status}%{SPACE}\\]\\[%{NOTSPACE:name}%{SPACE}\\]%{SPACE}(\\[%{HOSTNAME:nodeId}\\])?.*")

cast(shard, "int")
cast(duration, "int")

duration_precision(duration, "ms", "ns")

nullif(nodeId, "")
default_time(time)
`

type Input struct {
	Interval                   string   `toml:"interval"`
	Local                      bool     `toml:"local"`
	Distribution               string   `toml:"distribution"`
	Servers                    []string `toml:"servers"`
	HTTPTimeout                string   `toml:"http_timeout"`
	ClusterHealth              bool     `toml:"cluster_health"`
	ClusterHealthLevel         string   `toml:"cluster_health_level"`
	ClusterStats               bool     `toml:"cluster_stats"`
	ClusterStatsOnlyFromMaster bool     `toml:"cluster_stats_only_from_master"`
	IndicesInclude             []string `toml:"indices_include"`
	IndicesLevel               string   `toml:"indices_level"`
	NodeStats                  []string `toml:"node_stats"`
	Username                   string   `toml:"username"`
	Password                   string   `toml:"password"`
	Log                        *struct {
		Files             []string `toml:"files"`
		Pipeline          string   `toml:"pipeline"`
		IgnoreStatus      []string `toml:"ignore"`
		CharacterEncoding string   `toml:"character_encoding"`
		MultilineMatch    string   `toml:"multiline_match"`
	} `toml:"log"`

	Tags map[string]string `toml:"tags"`

	TLSOpen            bool   `toml:"tls_open"`
	CacertFile         string `toml:"tls_ca"`
	CertFile           string `toml:"tls_cert"`
	KeyFile            string `toml:"tls_key"`
	InsecureSkipVerify bool   `toml:"insecure_skip_verify"`

	httpTimeout     Duration
	client          *http.Client
	serverInfo      map[string]serverInfo
	serverInfoMutex sync.Mutex
	duration        time.Duration
	tail            *tailer.Tailer

	collectCache []inputs.Measurement

	Election bool `toml:"election"`
	pause    bool
	pauseCh  chan bool

	semStop *cliutils.Sem // start stop signal
}

func (i *Input) ElectionEnabled() bool {
	return i.Election
}

//nolint:lll
func (i *Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		inputName: {
			"ElasticSearch log":             `[2021-06-01T11:45:15,927][WARN ][o.e.c.r.a.DiskThresholdMonitor] [master] high disk watermark [90%] exceeded on [A2kEFgMLQ1-vhMdZMJV3Iw][master][/tmp/elasticsearch-cluster/nodes/0] free: 17.1gb[7.3%], shards will be relocated away from this node; currently relocating away shards totalling [0] bytes; the node is expected to continue to exceed the high disk watermark when these relocations are complete`,
			"ElasticSearch search slow log": `[2021-06-01T11:56:06,712][WARN ][i.s.s.query              ] [master] [shopping][0] took[36.3ms], took_millis[36], total_hits[5 hits], types[], stats[], search_type[QUERY_THEN_FETCH], total_shards[1], source[{"query":{"match":{"name":{"query":"Nariko","operator":"OR","prefix_length":0,"max_expansions":50,"fuzzy_transpositions":true,"lenient":false,"zero_terms_query":"NONE","auto_generate_synonyms_phrase_query":true,"boost":1.0}}},"sort":[{"price":{"order":"desc"}}]}], id[],`,
			"ElasticSearch index slow log":  `[2021-06-01T11:56:19,084][WARN ][i.i.s.index              ] [master] [shopping/X17jbNZ4SoS65zKTU9ZAJg] took[34.1ms], took_millis[34], type[_doc], id[LgC3xXkBLT9WrDT1Dovp], routing[], source[{"price":222,"name":"hello"}]`,
		},
	}
}

type userPrivilege struct {
	Cluster struct {
		Monitor bool `json:"monitor"`
	} `json:"cluster"`
	Index map[string]struct {
		Monitor bool `json:"monitor"`
		Ilm     bool `json:"manage_ilm"`
	} `json:"index"`
}

type serverInfo struct {
	nodeID        string
	masterID      string
	version       string
	userPrivilege *userPrivilege
}

func (i serverInfo) isMaster() bool {
	return i.nodeID == i.masterID
}

var maxPauseCh = inputs.ElectionPauseChannelLength

func NewElasticsearch() *Input {
	return &Input{
		httpTimeout:                Duration{Duration: time.Second * 5},
		ClusterStatsOnlyFromMaster: true,
		ClusterHealthLevel:         "indices",
		pauseCh:                    make(chan bool, maxPauseCh),
		Election:                   true,
		semStop:                    cliutils.NewSem(),
	}
}

// perform status mapping.
func mapHealthStatusToCode(s string) int {
	switch strings.ToLower(s) {
	case "green":
		return 1
	case "yellow":
		return 2
	case "red":
		return 3
	}
	return 0
}

// perform shard status mapping.
func mapShardStatusToCode(s string) int {
	switch strings.ToUpper(s) {
	case "UNASSIGNED":
		return 1
	case "INITIALIZING":
		return 2
	case "STARTED":
		return 3
	case "RELOCATING":
		return 4
	}
	return 0
}

var (
	inputName   = "elasticsearch"
	catalogName = "db"
	l           = logger.DefaultSLogger("elasticsearch")
)

func (*Input) Catalog() string {
	return catalogName
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (*Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		"elasticsearch": pipelineCfg,
	}
	return pipelineMap
}

func (i *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:  inputName,
			Service: inputName,
			Pipeline: func() string {
				if i.Log != nil {
					return i.Log.Pipeline
				}
				return ""
			}(),
		},
	}
}

func (i *Input) extendSelfTag(tags map[string]string) {
	if i.Tags != nil {
		for k, v := range i.Tags {
			tags[k] = v
		}
	}
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&nodeStatsMeasurement{},
		&indicesStatsMeasurement{},
		&clusterStatsMeasurement{},
		&clusterHealthMeasurement{},
	}
}

func (i *Input) setServerInfo() error {
	if len(i.Distribution) == 0 {
		i.Distribution = "elasticsearch"
	}
	i.serverInfo = make(map[string]serverInfo)

	g := goroutine.NewGroup(goroutine.Option{Name: goroutine.GetInputName(inputName)})
	for _, serv := range i.Servers {
		func(s string) {
			g.Go(func(ctx context.Context) error {
				var err error
				info := serverInfo{}

				// 获取nodeID和masterID
				if i.ClusterStats || len(i.IndicesInclude) > 0 || len(i.IndicesLevel) > 0 {
					// Gather node ID
					if info.nodeID, err = i.gatherNodeID(s + "/_nodes/_local/name"); err != nil {
						return fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@"))
					}

					// get cat/master information here so NodeStats can determine
					// whether this node is the Master
					if info.masterID, err = i.getCatMaster(s + "/_cat/master"); err != nil {
						return fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@"))
					}
				}

				if info.version, err = i.getVersion(s); err != nil {
					return fmt.Errorf(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@"))
				}

				if mask.MatchString(s) {
					info.userPrivilege = i.getUserPrivilege(s)
				}

				i.serverInfoMutex.Lock()
				i.serverInfo[s] = info
				i.serverInfoMutex.Unlock()

				return nil
			})
		}(serv)
	}
	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func (i *Input) Collect() error {
	if err := i.setServerInfo(); err != nil {
		return err
	}

	g := goroutine.NewGroup(goroutine.Option{Name: goroutine.GetInputName(inputName)})
	for _, serv := range i.Servers {
		func(s string) {
			g.Go(func(ctx context.Context) error {
				var clusterName string
				var err error
				url := i.nodeStatsURL(s)

				// Always gather node stats
				if clusterName, err = i.gatherNodeStats(url); err != nil {
					l.Warn(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@"))
				}

				if i.ClusterHealth {
					url = s + "/_cluster/health"
					if i.ClusterHealthLevel != "" {
						url = url + "?level=" + i.ClusterHealthLevel
					}
					if err := i.gatherClusterHealth(url, s); err != nil {
						l.Warn(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@"))
					}
				}

				if i.ClusterStats && (i.serverInfo[s].isMaster() || !i.ClusterStatsOnlyFromMaster || !i.Local) {
					if err := i.gatherClusterStats(s + "/_cluster/stats"); err != nil {
						l.Warn(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@"))
					}
				}

				if len(i.IndicesInclude) > 0 &&
					(i.serverInfo[s].isMaster() ||
						!i.ClusterStatsOnlyFromMaster ||
						!i.Local) {
					if i.IndicesLevel != "shards" {
						if err := i.gatherIndicesStats(s+
							"/"+
							strings.Join(i.IndicesInclude, ",")+
							"/_stats?ignore_unavailable=true", clusterName); err != nil {
							l.Warn(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@"))
						}
					} else {
						if err := i.gatherIndicesStats(s+
							"/"+
							strings.Join(i.IndicesInclude, ",")+
							"/_stats?level=shards&ignore_unavailable=true", clusterName); err != nil {
							l.Warn(mask.ReplaceAllString(err.Error(), "http(s)://XXX:XXX@"))
						}
					}
				}
				return nil
			})
		}(serv)
	}

	return g.Wait()
}

const (
	maxInterval = 1 * time.Minute
	minInterval = 1 * time.Second
)

func (i *Input) RunPipeline() {
	if i.Log == nil || len(i.Log.Files) == 0 {
		return
	}

	opt := &tailer.Option{
		Source:            inputName,
		Service:           inputName,
		Pipeline:          i.Log.Pipeline,
		GlobalTags:        i.Tags,
		IgnoreStatus:      i.Log.IgnoreStatus,
		CharacterEncoding: i.Log.CharacterEncoding,
		MultilinePatterns: []string{i.Log.MultilineMatch},
		Done:              i.semStop.Wait(),
	}

	var err error
	i.tail, err = tailer.NewTailer(i.Log.Files, opt)
	if err != nil {
		l.Error(err)
		io.FeedLastError(inputName, err.Error())
		return
	}
	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_elasticsearch"})
	g.Go(func(ctx context.Context) error {
		i.tail.Start()
		return nil
	})
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)

	duration, err := time.ParseDuration(i.Interval)
	if err != nil {
		l.Error("invalid interval, %s", err.Error())
		return
	} else if duration <= 0 {
		l.Error("invalid interval, cannot be less than zero")
		return
	}

	i.duration = config.ProtectedInterval(minInterval, maxInterval, duration)

	i.httpTimeout = Duration{}
	if len(i.HTTPTimeout) > 0 {
		err := i.httpTimeout.UnmarshalTOML([]byte(i.HTTPTimeout))
		if err != nil {
			l.Warnf("invalid http timeout, %s", i.HTTPTimeout)
		}
	}

	client, err := i.createHTTPClient()
	if err != nil {
		l.Error(err)
		io.FeedLastError(inputName, err.Error())
		return
	}
	i.client = client

	defer i.stop()

	tick := time.NewTicker(i.duration)
	defer tick.Stop()

	for {
		if i.pause {
			l.Debugf("not leader, skipped")
		} else {
			start := time.Now()
			if err := i.Collect(); err != nil {
				io.FeedLastError(inputName, err.Error())
				l.Error(err)
			} else if len(i.collectCache) > 0 {
				err := inputs.FeedMeasurement("elasticsearch",
					datakit.Metric,
					i.collectCache,
					&io.Option{CollectCost: time.Since(start)})
				if err != nil {
					io.FeedLastError(inputName, err.Error())
					l.Errorf(err.Error())
				}
				i.collectCache = i.collectCache[:0]
			}
		}
		select {
		case <-datakit.Exit.Wait():
			i.exit()
			l.Info("elasticsearch exit")
			return

		case <-i.semStop.Wait():
			i.exit()
			l.Info("elasticsearch return")
			return

		case <-tick.C:

		case i.pause = <-i.pauseCh:
			// nil
		}
	}
}

func (i *Input) exit() {
	if i.tail != nil {
		i.tail.Close()
		l.Info("elasticsearch log exit")
	}
}

func (i *Input) Terminate() {
	if i.semStop != nil {
		i.semStop.Close()
	}
}

func (i *Input) gatherIndicesStats(url string, clusterName string) error {
	indicesStats := &struct {
		Shards  map[string]interface{} `json:"_shards"`
		All     map[string]interface{} `json:"_all"`
		Indices map[string]indexStat   `json:"indices"`
	}{}

	if err := i.gatherJSONData(url, indicesStats); err != nil {
		return err
	}
	now := time.Now()

	// All Stats
	for m, s := range indicesStats.All {
		// parse Json, ignoring strings and bools
		jsonParser := JSONFlattener{}
		err := jsonParser.FullFlattenJSON(m, s, true, true)
		if err != nil {
			return err
		}

		allFields := make(map[string]interface{})
		for k, v := range jsonParser.Fields {
			_, ok := indicesStatsFields[k]
			if ok {
				allFields[k] = v
			}
		}

		tags := map[string]string{"index_name": "_all", "cluster_name": clusterName}
		setHostTagIfNotLoopback(tags, url)
		i.extendSelfTag(tags)

		metric := &indicesStatsMeasurement{
			elasticsearchMeasurement: elasticsearchMeasurement{
				name:     "elasticsearch_indices_stats",
				tags:     tags,
				fields:   allFields,
				ts:       now,
				election: i.Election,
			},
		}

		if len(metric.fields) > 0 {
			i.collectCache = append(i.collectCache, metric)
		}
	}

	// Individual Indices stats
	for id, index := range indicesStats.Indices {
		indexTag := map[string]string{"index_name": id, "cluster_name": clusterName}
		stats := map[string]interface{}{
			"primaries": index.Primaries,
			"total":     index.Total,
		}
		for m, s := range stats {
			f := JSONFlattener{}
			// parse Json, getting strings and bools
			err := f.FullFlattenJSON(m, s, true, true)
			if err != nil {
				return err
			}

			allFields := make(map[string]interface{})
			for k, v := range f.Fields {
				_, ok := indicesStatsFields[k]
				if ok {
					allFields[k] = v
				}
			}

			setHostTagIfNotLoopback(indexTag, url)
			i.extendSelfTag(indexTag)
			metric := &indicesStatsMeasurement{
				elasticsearchMeasurement: elasticsearchMeasurement{
					name:     "elasticsearch_indices_stats",
					tags:     indexTag,
					fields:   allFields,
					ts:       now,
					election: i.Election,
				},
			}

			if len(metric.fields) > 0 {
				i.collectCache = append(i.collectCache, metric)
			}
		}
	}

	return nil
}

func (i *Input) gatherNodeStats(url string) (string, error) {
	nodeStats := &struct {
		ClusterName string               `json:"cluster_name"`
		Nodes       map[string]*nodeStat `json:"nodes"`
	}{}

	if err := i.gatherJSONData(url, nodeStats); err != nil {
		return "", err
	}

	for id, n := range nodeStats.Nodes {
		sort.Strings(n.Roles)
		tags := map[string]string{
			"node_id":      id,
			"node_host":    n.Host,
			"node_name":    n.Name,
			"cluster_name": nodeStats.ClusterName,
			"node_roles":   strings.Join(n.Roles, ","),
		}

		for k, v := range n.Attributes {
			tags["node_attribute_"+k] = v
		}

		stats := map[string]interface{}{
			"indices":     n.Indices,
			"os":          n.OS,
			"process":     n.Process,
			"jvm":         n.JVM,
			"thread_pool": n.ThreadPool,
			"fs":          n.FS,
			"transport":   n.Transport,
			"http":        n.HTTP,
			"breakers":    n.Breakers,
		}

		//nolint:lll
		const cols = `fs_total_available_in_bytes,fs_total_free_in_bytes,fs_total_total_in_bytes,fs_data_0_available_in_bytes,fs_data_0_free_in_bytes,fs_data_0_total_in_bytes`

		now := time.Now()
		allFields := make(map[string]interface{})
		for p, s := range stats {
			// if one of the individual node stats is not even in the
			// original result
			if s == nil {
				continue
			}
			f := JSONFlattener{}
			// parse Json, ignoring strings and bools
			err := f.FlattenJSON(p, s)
			if err != nil {
				return "", err
			}
			for k, v := range f.Fields {
				filedName := k
				val := v
				// transform bytes to gigabytes
				if p == "fs" {
					if strings.Contains(cols, filedName) {
						if value, ok := v.(float64); ok {
							val = value / (1024 * 1024 * 1024)
							filedName = strings.ReplaceAll(filedName, "in_bytes", "in_gigabytes")
						}
					}
				}
				_, ok := nodeStatsFields[filedName]
				if ok {
					allFields[filedName] = val
				}
			}
		}

		setHostTagIfNotLoopback(tags, url)
		i.extendSelfTag(tags)
		metric := &nodeStatsMeasurement{
			elasticsearchMeasurement: elasticsearchMeasurement{
				name:     "elasticsearch_node_stats",
				tags:     tags,
				fields:   allFields,
				ts:       now,
				election: i.Election,
			},
		}
		if len(metric.fields) > 0 {
			i.collectCache = append(i.collectCache, metric)
		}
	}

	return nodeStats.ClusterName, nil
}

func (i *Input) gatherClusterStats(url string) error {
	clusterStats := &clusterStats{}
	if err := i.gatherJSONData(url, clusterStats); err != nil {
		return err
	}
	now := time.Now()
	tags := map[string]string{
		"node_name":    clusterStats.NodeName,
		"cluster_name": clusterStats.ClusterName,
		"status":       clusterStats.Status,
	}

	stats := map[string]interface{}{
		"nodes":   clusterStats.Nodes,
		"indices": clusterStats.Indices,
	}

	allFields := make(map[string]interface{})
	for p, s := range stats {
		f := JSONFlattener{}
		// parse json, including bools and strings
		err := f.FullFlattenJSON(p, s, true, true)
		if err != nil {
			return err
		}
		for k, v := range f.Fields {
			_, ok := clusterStatsFields[k]
			if ok {
				allFields[k] = v
			}
		}
	}

	setHostTagIfNotLoopback(tags, url)
	i.extendSelfTag(tags)
	metric := &clusterStatsMeasurement{
		elasticsearchMeasurement: elasticsearchMeasurement{
			name:     "elasticsearch_cluster_stats",
			tags:     tags,
			fields:   allFields,
			ts:       now,
			election: i.Election,
		},
	}

	if len(metric.fields) > 0 {
		i.collectCache = append(i.collectCache, metric)
	}
	return nil
}

func (i *Input) isVersion6(url string) bool {
	serverInfo, ok := i.serverInfo[url]
	if !ok { // default
		return false
	} else {
		parts := strings.Split(serverInfo.version, ".")
		if len(parts) >= 2 {
			return parts[0] == "6"
		}
	}

	return false
}

func (i *Input) getLifeCycleErrorCount(url string) (errCount int) {
	errCount = 0
	// default elasticsearch
	if i.Distribution == "elasticsearch" || (len(i.Distribution) == 0) {
		// check privilege
		privilege := i.serverInfo[url].userPrivilege
		if privilege != nil {
			indexPrivilege, ok := privilege.Index["all"]
			if ok {
				if !indexPrivilege.Ilm {
					l.Warn("user has no ilm privilege, ingore collect indices_lifecycle_error_count")
					return 0
				}
			}
		}

		indicesRes := &indexState{}
		if i.isVersion6(url) { // 6.x
			if err := i.gatherJSONData(url+"/*/_ilm/explain", indicesRes); err != nil {
				l.Warn(err)
			} else {
				for _, index := range indicesRes.Indices {
					if index.Managed && index.Step == "ERROR" {
						errCount += 1
					}
				}
			}
		} else {
			if err := i.gatherJSONData(url+"/*/_ilm/explain?only_errors", indicesRes); err != nil {
				l.Warn(err)
			} else {
				errCount = len(indicesRes.Indices)
			}
		}
	}

	// opendistro or opensearch
	if i.Distribution == "opendistro" || i.Distribution == "opensearch" {
		res := map[string]interface{}{}
		pluginName := "_opendistro"

		if i.Distribution == "opensearch" {
			pluginName = "_plugins"
		}

		if err := i.gatherJSONData(url+"/"+pluginName+"/_ism/explain/*", &res); err != nil {
			l.Warn(err)
		} else {
			for _, index := range res {
				indexVal, ok := index.(map[string]interface{})
				if ok {
					if step, ok := indexVal["step"]; ok {
						if stepVal, ok := step.(map[string]interface{}); ok {
							if status, ok := stepVal["step_status"]; ok {
								if statusVal, ok := status.(string); ok {
									if statusVal == "failed" {
										errCount += 1
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return errCount
}

func (i *Input) gatherClusterHealth(url string, serverURL string) error {
	healthStats := &clusterHealth{}
	if err := i.gatherJSONData(url, healthStats); err != nil {
		return err
	}
	indicesErrorCount := i.getLifeCycleErrorCount(serverURL)
	now := time.Now()
	clusterFields := map[string]interface{}{
		"active_primary_shards":            healthStats.ActivePrimaryShards,
		"active_shards":                    healthStats.ActiveShards,
		"active_shards_percent_as_number":  healthStats.ActiveShardsPercentAsNumber,
		"delayed_unassigned_shards":        healthStats.DelayedUnassignedShards,
		"initializing_shards":              healthStats.InitializingShards,
		"number_of_data_nodes":             healthStats.NumberOfDataNodes,
		"number_of_in_flight_fetch":        healthStats.NumberOfInFlightFetch,
		"number_of_nodes":                  healthStats.NumberOfNodes,
		"number_of_pending_tasks":          healthStats.NumberOfPendingTasks,
		"relocating_shards":                healthStats.RelocatingShards,
		"status_code":                      mapHealthStatusToCode(healthStats.Status),
		"task_max_waiting_in_queue_millis": healthStats.TaskMaxWaitingInQueueMillis,
		"timed_out":                        healthStats.TimedOut,
		"unassigned_shards":                healthStats.UnassignedShards,
		"indices_lifecycle_error_count":    indicesErrorCount,
	}

	allFields := make(map[string]interface{})

	for k, v := range clusterFields {
		_, ok := clusterHealthFields[k]
		if ok {
			allFields[k] = v
		}
	}

	tags := map[string]string{
		"name":           healthStats.ClusterName, // depreciated, may be discarded in future
		"cluster_name":   healthStats.ClusterName,
		"cluster_status": healthStats.Status,
	}

	setHostTagIfNotLoopback(tags, url)
	i.extendSelfTag(tags)
	metric := &clusterHealthMeasurement{
		elasticsearchMeasurement: elasticsearchMeasurement{
			name:     "elasticsearch_cluster_health",
			tags:     tags,
			fields:   allFields,
			ts:       now,
			election: i.Election,
		},
	}

	if len(metric.fields) > 0 {
		i.collectCache = append(i.collectCache, metric)
	}

	return nil
}

func (i *Input) gatherNodeID(url string) (string, error) {
	nodeStats := &struct {
		ClusterName string               `json:"cluster_name"`
		Nodes       map[string]*nodeStat `json:"nodes"`
	}{}
	if err := i.gatherJSONData(url, nodeStats); err != nil {
		return "", err
	}

	// Only 1 should be returned
	for id := range nodeStats.Nodes {
		return id, nil
	}
	return "", nil
}

func (i *Input) getVersion(url string) (string, error) {
	clusterInfo := &struct {
		Version struct {
			Number string `json:"number"`
		} `json:"version"`
	}{}
	if err := i.gatherJSONData(url, clusterInfo); err != nil {
		return "", err
	}

	return clusterInfo.Version.Number, nil
}

func (i *Input) getUserPrivilege(url string) *userPrivilege {
	privilege := &userPrivilege{}
	if i.Distribution == "elasticsearch" || len(i.Distribution) == 0 {
		body := strings.NewReader(`{"cluster": ["monitor"],"index":[{"names":["all"], "privileges":["monitor","manage_ilm"]}]}`)
		header := map[string]string{"Content-Type": "application/json"}
		if err := i.requestData("GET", url+"/_security/user/_has_privileges", header, body, privilege); err != nil {
			l.Warnf("get user privilege error: %s", err.Error())
			return nil
		}
	}

	return privilege
}

func (i *Input) nodeStatsURL(baseURL string) string {
	var url string

	if i.Local {
		url = baseURL + statsPathLocal
	} else {
		url = baseURL + statsPath
	}

	if len(i.NodeStats) == 0 {
		return url
	}

	return fmt.Sprintf("%s/%s", url, strings.Join(i.NodeStats, ","))
}

func (i *Input) stop() {
	i.client.CloseIdleConnections()
}

func (i *Input) createHTTPClient() (*http.Client, error) {
	timeout := 10 * time.Second
	if i.httpTimeout.Duration > 0 {
		timeout = i.httpTimeout.Duration
	}
	client := &http.Client{
		Timeout: timeout,
	}

	if i.TLSOpen {
		if i.InsecureSkipVerify {
			client.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // nolint:gosec
			}
		} else {
			tc, err := TLSConfig(i.CacertFile, i.CertFile, i.KeyFile)
			if err != nil {
				return nil, err
			} else {
				client.Transport = &http.Transport{
					TLSClientConfig: tc,
				}
			}
		}
	} else {
		if len(i.Servers) > 0 {
			server := i.Servers[0]
			if strings.HasPrefix(server, "https://") {
				client.Transport = &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // nolint:gosec
				}
			}
		}
	}

	return client, nil
}

func (i *Input) requestData(method string, url string, header map[string]string, body internalIo.Reader, v interface{}) error {
	m := "GET"
	if len(method) > 0 {
		m = method
	}
	req, err := http.NewRequest(m, url, body)
	for k, v := range header {
		req.Header.Add(k, v)
	}
	if err != nil {
		return err
	}

	if i.Username != "" || i.Password != "" {
		req.SetBasicAuth(i.Username, i.Password)
	}

	r, err := i.client.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close() //nolint:errcheck
	if r.StatusCode != http.StatusOK {
		// NOTE: we are not going to read/discard r.Body under the assumption we'd prefer
		// to let the underlying transport close the connection and re-establish a new one for
		// future calls.
		resBodyBytes, err := ioutil.ReadAll(r.Body)
		resBody := ""
		if err != nil {
			l.Debugf("get response body err: %s", err.Error())
		} else {
			resBody = string(resBodyBytes)
		}

		l.Debugf("response body: %s", resBody)
		return fmt.Errorf("elasticsearch: API responded with status-code %d, expected %d, url: %s",
			r.StatusCode, http.StatusOK, mask.ReplaceAllString(url, "http(s)://XXX:XXX@"))
	}

	if err = json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}

	return nil
}

func (i *Input) gatherJSONData(url string, v interface{}) error {
	return i.requestData("GET", url, nil, nil, v)
}

func (i *Input) getCatMaster(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	if i.Username != "" || i.Password != "" {
		req.SetBasicAuth(i.Username, i.Password)
	}

	r, err := i.client.Do(req)
	if err != nil {
		return "", err
	}
	defer r.Body.Close() //nolint:errcheck
	if r.StatusCode != http.StatusOK {
		// NOTE: we are not going to read/discard r.Body under the assumption we'd prefer
		// to let the underlying transport close the connection and re-establish a new one for
		// future calls.
		//nolint:lll
		return "", fmt.Errorf("elasticsearch: Unable to retrieve master node information. API responded with status-code %d, expected %d", r.StatusCode, http.StatusOK)
	}
	response, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	masterID := strings.Split(string(response), " ")[0]

	return masterID, nil
}

func (i *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case i.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (i *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case i.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return NewElasticsearch()
	})
}
