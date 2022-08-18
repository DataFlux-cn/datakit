// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package mongodb collects MongoDB metrics.
package mongodb

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gopkg.in/mgo.v2"
)

var _ inputs.ElectionInput = (*Input)(nil)

var (
	defInterval      = datakit.Duration{Duration: 10 * time.Second}
	defMongoURL      = "mongodb://127.0.0.1:27017"
	defTLSCaCert     = "/etc/ssl/certs/mongod.cert.pem"
	defTLSCert       = "/etc/ssl/certs/mongo.cert.pem"
	defTLSCertKey    = "/etc/ssl/certs/mongo.key.pem"
	defMongodLogPath = "/var/log/mongodb/mongod.log"
	defPipeline      = "mongod.p"
	defTags          map[string]string
)

var (
	inputName    = "mongodb"
	catalogName  = "db"
	sampleConfig = `
[[inputs.mongodb]]
  ## Gathering interval
  # interval = "` + defInterval.UnitString(time.Second) + `"

  ## An array of URLs of the form:
  ##   "mongodb://" [user ":" pass "@"] host [ ":" port]
  ## For example:
  ##   mongodb://user:auth_key@10.10.3.30:27017,
  ##   mongodb://10.10.3.33:18832,
  # servers = ["` + defMongoURL + `"]

  ## When true, collect replica set stats
  # gather_replica_set_stats = false

  ## When true, collect cluster stats
  ## Note that the query that counts jumbo chunks triggers a COLLSCAN, which may have an impact on performance.
  # gather_cluster_stats = false

  ## When true, collect per database stats
  # gather_per_db_stats = true

  ## When true, collect per collection stats
  # gather_per_col_stats = true

  ## List of db where collections stats are collected, If empty, all dbs are concerned.
  # col_stats_dbs = []

  ## When true, collect top command stats.
  # gather_top_stat = true

  ## Set true to enable election
  election = true

  ## TLS connection config
  # ca_certs = ["` + defTLSCaCert + `"]
  # cert = "` + defTLSCert + `"
  # cert_key = "` + defTLSCertKey + `"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = true
  # server_name = ""

  ## Mongod log
  # [inputs.mongodb.log]
  # #Log file path check your mongodb config path usually under '/var/log/mongodb/mongod.log'.
  # files = ["` + defMongodLogPath + `"]
  # #Grok pipeline script file.
  # pipeline = "` + defPipeline + `"

  ## Customer tags, if set will be seen with every metric.
  [inputs.mongodb.tags]
    # "key1" = "value1"
    # "key2" = "value2"
		# ...
`
	pipelineConfig = `
  json(_, t, "tmp")
  json(tmp, ` + "`" + "$date" + "`" + `, "time")
  json(_, s, "status")
  json(_, c, "component")
  json(_, msg, "msg")
  json(_, ctx, "context")
  drop_key(tmp)
  default_time(time)
`
	l = logger.DefaultSLogger(inputName)
)

type mongodblog struct {
	Files             []string `toml:"files"`
	Pipeline          string   `toml:"pipeline"`
	IgnoreStatus      []string `toml:"ignore"`
	CharacterEncoding string   `toml:"character_encoding"`
	MultilineMatch    string   `toml:"multiline_match"`
}

type Input struct {
	Interval              datakit.Duration `toml:"interval"`
	Servers               []string         `toml:"servers"`
	GatherReplicaSetStats bool             `toml:"gather_replica_set_stats"`
	GatherClusterStats    bool             `toml:"gather_cluster_stats"`
	GatherPerDBStats      bool             `toml:"gather_per_db_stats"`
	GatherPerColStats     bool             `toml:"gather_per_col_stats"`
	ColStatsDBs           []string         `toml:"col_stats_dbs"`
	GatherTopStat         bool             `toml:"gather_top_stat"`

	TLSConf *dknet.TLSClientConfig `toml:"tlsconf"` // deprecated

	EnableTLSDeprecated bool `toml:"enable_tls,omitempty"`

	dknet.TLSClientConfig

	Log  *mongodblog       `toml:"log"`
	Tags map[string]string `toml:"tags"`

	mongos map[string]*Server
	tail   *tailer.Tailer

	Election bool `toml:"election"`
	pause    bool
	pauseCh  chan bool

	semStop *cliutils.Sem // start stop signal
}

func (m *Input) ElectionEnabled() bool {
	return m.Election
}

func (*Input) Catalog() string { return catalogName }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) PipelineConfig() map[string]string {
	return map[string]string{inputName: pipelineConfig}
}

//nolint:lll
func (m *Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		inputName: {
			"MongoDB log": `{"t":{"$date":"2021-06-03T09:12:19.977+00:00"},"s":"I",  "c":"STORAGE",  "id":22430,   "ctx":"WTCheckpointThread","msg":"WiredTiger message","attr":{"message":"[1622711539:977142][1:0x7f1b9f159700], WT_SESSION.checkpoint: [WT_VERB_CHECKPOINT_PROGRESS] saving checkpoint snapshot min: 653, snapshot max: 653 snapshot count: 0, oldest timestamp: (0, 0) , meta checkpoint timestamp: (0, 0)"}}`,
		},
	}
}

func (m *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:  inputName,
			Service: inputName,
			Pipeline: func() string {
				if m.Log != nil {
					return m.Log.Pipeline
				}
				return ""
			}(),
		},
	}
}

func (m *Input) AvailableArchs() []string { return datakit.AllOS }

func (m *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&mongodbMeasurement{},
		&mongodbDBMeasurement{},
		&mongodbColMeasurement{},
		&mongodbShardMeasurement{},
		&mongodbTopMeasurement{},
	}
}

func (m *Input) RunPipeline() {
	if m.Log == nil || len(m.Log.Files) == 0 {
		return
	}

	if m.Log.Pipeline == "" {
		m.Log.Pipeline = "mongod.p" // use default
	}

	opt := &tailer.Option{
		Source:            inputName,
		Service:           inputName,
		Pipeline:          m.Log.Pipeline,
		GlobalTags:        m.Tags,
		IgnoreStatus:      m.Log.IgnoreStatus,
		CharacterEncoding: m.Log.CharacterEncoding,
		MultilinePatterns: []string{m.Log.MultilineMatch},
	}

	var err error
	m.tail, err = tailer.NewTailer(m.Log.Files, opt)
	if err != nil {
		l.Errorf("NewTailer: %s", err)

		io.FeedLastError(inputName, err.Error())
		return
	}

	go m.tail.Start()
}

func (m *Input) Run() {
	l = logger.SLogger(inputName)
	l.Info("mongodb input started")

	defTags = m.Tags

	tick := time.NewTicker(m.Interval.Duration)

	for {
		if m.pause {
			l.Debugf("not leader, skipped")
		} else if err := m.gather(); err != nil {
			l.Errorf("gather: %s", err.Error())
			io.FeedLastError(inputName, err.Error())
		}

		select {
		case <-datakit.Exit.Wait():
			m.exit()
			l.Info("mongodb input exit")

			return
		case <-m.semStop.Wait():
			m.exit()
			l.Info("mongodb input return")

			return
		case <-tick.C:
		case m.pause = <-m.pauseCh:
			// nil
		}
	}
}

func (m *Input) exit() {
	if m.tail != nil {
		m.tail.Close()
		l.Info("mongodb log exits")
	}
}

func (m *Input) Terminate() {
	if m.semStop != nil {
		m.semStop.Close()
	}
}

func (m *Input) getMongoServer(url *url.URL) *Server {
	if _, ok := m.mongos[url.Host]; !ok {
		m.mongos[url.Host] = &Server{URL: url, election: m.Election}
	}

	return m.mongos[url.Host]
}

// Reads stats from all configured servers.
// Returns one of the errors encountered while gather stats (if any).
func (m *Input) gather() error {
	if len(m.Servers) == 0 {
		return m.gatherServer(m.getMongoServer(&url.URL{Host: defMongoURL}))
	}

	var wg sync.WaitGroup
	for i, serv := range m.Servers {
		if !strings.HasPrefix(serv, "mongodb://") {
			serv = "mongodb://" + serv
			l.Warnf("using %q as connection URL; please update your configuration to use an URL", serv)
			m.Servers[i] = serv
		}

		u, err := url.Parse(serv)
		if err != nil {
			l.Errorf("unable to parse address %q: %s", serv, err.Error())
			continue
		}
		if u.Host == "" {
			l.Errorf("unable to parse address %q", serv)
			continue
		}

		wg.Add(1)
		go func(srv *Server) {
			defer wg.Done()

			if err := m.gatherServer(srv); err != nil {
				l.Errorf("error in plugin: %s,%v", srv.URL.String(), err)
			}
		}(m.getMongoServer(u))
	}
	wg.Wait()

	return nil
}

func (m *Input) gatherServer(server *Server) error {
	if server.Session == nil {
		var dialAddrs []string
		if server.URL.User != nil {
			dialAddrs = []string{server.URL.String()}
		} else {
			dialAddrs = []string{server.URL.Host}
		}

		dialInfo, err := mgo.ParseURL(dialAddrs[0])
		if err != nil {
			return fmt.Errorf("unable to parse URL %q: %w", dialAddrs[0], err)
		}

		tlscnf := m.TLSConf // prefer deprecated TLS conf
		if tlscnf == nil {
			tlscnf = &m.TLSClientConfig
		}

		if tlsConfig, err := tlscnf.TLSConfig(); err != nil {
			return err
		} else if tlsConfig != nil {
			// TLS is configured
			dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
				return tls.Dial("tcp", addr.String(), tlsConfig)
			}
		}

		dialInfo.Direct = true
		dialInfo.Timeout = 5 * time.Second

		sess, err := mgo.DialWithInfo(dialInfo)
		if err != nil {
			return fmt.Errorf("unable to connect to MongoDB: %w", err)
		}
		server.Session = sess
	}

	return server.gatherData(m.GatherReplicaSetStats,
		m.GatherClusterStats,
		m.GatherPerDBStats,
		m.GatherPerColStats,
		m.ColStatsDBs,
		m.GatherTopStat)
}

func (m *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case m.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (m *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case m.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Interval:              defInterval,
			Servers:               []string{defMongoURL},
			GatherReplicaSetStats: false,
			GatherClusterStats:    false,
			GatherPerDBStats:      true,
			GatherPerColStats:     true,
			ColStatsDBs:           []string{},
			GatherTopStat:         true,
			Log:                   &mongodblog{Files: []string{defMongodLogPath}, Pipeline: defPipeline},
			mongos:                make(map[string]*Server),
			pauseCh:               make(chan bool, inputs.ElectionPauseChannelLength),
			Election:              true,

			semStop: cliutils.NewSem(),
		}
	})
}
