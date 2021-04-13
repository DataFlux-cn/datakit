package mysql

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"github.com/go-sql-driver/mysql"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName   = "mysql"
	catalogName = "db"
	l           = logger.DefaultSLogger("mysql")
)

type tls struct {
	TlsKey  string `toml:"tls_key"`
	TlsCert string `toml:"tls_cert"`
	TlsCA   string `toml:"tls_ca"`
}

type options struct {
	Replication             bool `toml:"replication"`
	GaleraCluster           bool `toml:"galera_cluster"`
	ExtraStatusMetrics      bool `toml:"extra_status_metrics"`
	ExtraInnodbMetrics      bool `toml:"extra_innodb_metrics"`
	DisableInnodbMetrics    bool `toml:"disable_innodb_metrics"`
	SchemaSizeMetrics       bool `toml:"schema_size_metrics"`
	ExtraPerformanceMetrics bool `toml:"extra_performance_metrics"`
}

type customQuery struct {
	sql    string   `toml:"sql"`
	metric string   `toml:"metric"`
	tags   []string `toml:"tags"`
	fields []string `toml:"fields"`
}

type Input struct {
	Host             string                   `toml:"host"`
	Port             int                      `toml:"port"`
	User             string                   `toml:"user"`
	Pass             string                   `toml:"pass"`
	Sock             string                   `toml:"sock"`
	Charset          string                   `toml:"charset"`
	Timeout          string                   `toml:"connect_timeout"`
	TimeoutDuration  time.Duration            `toml:"-"`
	Tls              *tls                     `toml:"tls"`
	Service          string                   `toml:"service"`
	Interval         string                   `toml:"interval"`
	IntervalDuration time.Duration            `toml:"-"`
	Tags             map[string]string        `toml:"tags"`
	options          *options                 `toml:"options"`
	Query            []*customQuery           `toml:"custom_queries"`
	db               *sql.DB                  `toml:"-"`
	Addr             string                   `toml:"-"`
	collectCache     []inputs.Measurement     `toml:"-"`
	response         []map[string]interface{} `toml:"-"`
	Log              *inputs.TailerOption     `toml:"log"`
	tailer           *inputs.Tailer           `toml:"-"`
}

func (i *Input) getDsnString() string {
	cfg := mysql.Config{
		AllowNativePasswords: true,
		CheckConnLiveness:    true,
		User:                 i.User,
		Passwd:               i.Pass,
	}

	// set addr
	if i.Sock != "" {
		cfg.Net = "unix"
		cfg.Addr = i.Sock
	} else {
		addr := fmt.Sprintf("%s:%d", i.Host, i.Port)
		cfg.Net = "tcp"
		cfg.Addr = addr
	}
	i.Addr = cfg.Addr

	// set timeout
	if i.TimeoutDuration != 0 {
		cfg.Timeout = i.TimeoutDuration
	}

	// set Charset
	if i.Charset != "" {
		cfg.Params["charset"] = i.Charset
	}

	// ssl
	if i.Tls != nil {

	}

	// tls (todo)
	return cfg.FormatDSN()
}

func (i *Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		"mysql": pipelineCfg,
	}
	return pipelineMap
}

func (i *Input) initCfg() {
	// 采集频度
	i.IntervalDuration = 5 * time.Minute

	if i.Interval != "" {
		du, err := time.ParseDuration(i.Interval)
		if err != nil {
			l.Errorf("bad interval %s: %s, use default: 1m", i.Interval, err.Error())
		} else {
			i.IntervalDuration = du
		}
	}

	dsnStr := i.getDsnString()
	l.Infof("db build dsn connect str %s", dsnStr)
	db, err := sql.Open("mysql", dsnStr)
	if err != nil {
		l.Errorf("sql.Open(): %s", err.Error())
	} else {
		i.db = db
	}

	i.globalTag()
}

func (i *Input) globalTag() {
	i.Tags["server"] = i.Addr
	i.Tags["service_name"] = i.Service
}

func (i *Input) Collect() error {
	i.collectCache = []inputs.Measurement{}

	i.collectBaseMeasurement()
	i.collectSchemaMeasurement()
	i.customSchemaMeasurement()

	return nil
}

// 获取base指标
func (i *Input) collectBaseMeasurement() {
	m := &baseMeasurement{
		client:  i.db,
		resData: make(map[string]interface{}),
		tags:    make(map[string]string),
		fields:  make(map[string]interface{}),
	}

	m.name = "mysql_base"
	for key, value := range i.Tags {
		m.tags[key] = value
	}

	m.getStatus()
	m.getVariables()
	m.getLogStats()

	m.submit()

	i.collectCache = append(i.collectCache, m)
}

// 获取schema指标
func (i *Input) collectSchemaMeasurement() {
	i.getSchemaSize()
	i.getQueryExecTimePerSchema()
}

func (i *Input) Run() {
	l = logger.SLogger("mysql")
	i.initCfg()

	if i.Log != nil {
		go func() {
			pfile := "redis.p"
			if i.Log.Pipeline != "" {
				pfile = i.Log.Pipeline
			}

			i.Log.Pipeline = filepath.Join(datakit.PipelineDir, pfile)

			i.Log.Source = inputName
			for k, v := range i.Tags {
				i.Log.Tags[k] = v
			}
			tailer, err := inputs.NewTailer(i.Log)
			if err != nil {
				l.Errorf("init tailf err:%s", err.Error())
				return
			}
			i.tailer = tailer
			tailer.Run()
		}()
	}

	tick := time.NewTicker(i.IntervalDuration)
	defer tick.Stop()

	n := 0

	for {
		n++
		select {
		case <-tick.C:
			l.Debugf("redis input gathering...")
			start := time.Now()
			if err := i.Collect(); err != nil {
				l.Error(err)
			} else {
				inputs.FeedMeasurement(inputName, io.Metric, i.collectCache,
					&io.Option{CollectCost: time.Since(start), HighFreq: (n%2 == 0)})

				i.collectCache = i.collectCache[:] // NOTE: do not forget to clean cache
			}

		case <-datakit.Exit.Wait():
			if i.tailer != nil {
				i.tailer.Close()
				l.Info("mysql log exit")
			}
			l.Info("mysql exit")
			return
		}
	}
}

func (i *Input) Catalog() string { return catalogName }

func (i *Input) SampleConfig() string { return configSample }

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&baseMeasurement{},
		&schemaMeasurement{},
	}
}

func (i *Input) AvailableArchs() []string {
	return datakit.UnknownArch
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
