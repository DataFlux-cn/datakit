package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-sql-driver/mysql"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	maxInterval = 15 * time.Minute
	minInterval = 10 * time.Second
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

type mysqllog struct {
	Files             []string `toml:"files"`
	Pipeline          string   `toml:"pipeline"`
	IgnoreStatus      []string `toml:"ignore"`
	CharacterEncoding string   `toml:"character_encoding"`
	Match             string   `toml:"match"`
}

type Input struct {
	Host   string   `toml:"host"`
	Port   int      `toml:"port"`
	User   string   `toml:"user"`
	Pass   string   `toml:"pass"`
	Sock   string   `toml:"sock"`
	Tables []string `toml:"tables"`
	Users  []string `toml:"users"`

	Charset string `toml:"charset"`

	Timeout         string        `toml:"connect_timeout"`
	timeoutDuration time.Duration `toml:"-"`

	Tls *tls `toml:"tls"`

	Service  string `toml:"service"`
	Interval datakit.Duration

	Tags map[string]string `toml:"tags"`

	options *options       `toml:"options"`
	Query   []*customQuery `toml:"custom_queries"`
	Addr    string         `toml:"-"`
	InnoDB  bool           `toml:"innodb"`
	Log     *mysqllog      `toml:"log"`

	start      time.Time                `toml:"-"`
	db         *sql.DB                  `toml:"-"`
	response   []map[string]interface{} `toml:"-"`
	tail       *tailer.Tailer           `toml:"-"`
	err        error
	collectors []func() ([]inputs.Measurement, error) `toml:"-"`
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
	if i.timeoutDuration != 0 {
		cfg.Timeout = i.timeoutDuration
	}

	// set Charset
	if i.Charset != "" {
		cfg.Params["charset"] = i.Charset
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

func (i *Input) initCfg() error {
	var err error
	i.timeoutDuration, err = time.ParseDuration(i.Timeout)
	if err != nil {
		i.timeoutDuration = 10 * time.Second
	}

	dsnStr := i.getDsnString()

	db, err := sql.Open("mysql", dsnStr)
	if err != nil {
		l.Errorf("sql.Open(): %s", err.Error())
		return err
	} else {
		i.db = db
	}

	ctx, cancel := context.WithTimeout(context.Background(), i.timeoutDuration)
	defer cancel()

	if err := i.db.PingContext(ctx); err != nil {
		l.Errorf("init config connect error %v", err)
		return err
	}

	i.globalTag()
	return nil
}

func (i *Input) globalTag() {
	i.Tags["server"] = i.Addr
	i.Tags["service_name"] = i.Service
}

func (i *Input) Collect() error {
	ctx, cancel := context.WithTimeout(context.Background(), i.timeoutDuration)
	defer cancel()

	if err := i.db.PingContext(ctx); err != nil {
		l.Errorf("connect error %v", err)
		io.FeedLastError(inputName, err.Error())
		return err
	}

	for idx, f := range i.collectors {
		l.Debugf("collecting %d(%v)...", idx, f)

		if ms, err := f(); err != nil {
			io.FeedLastError(inputName, err.Error())
		} else {
			if len(ms) > 0 {
				if err := inputs.FeedMeasurement(inputName,
					datakit.Metric,
					ms,
					&io.Option{CollectCost: time.Since(i.start)}); err != nil {
					l.Error(err)
				}
			}
		}
	}

	return nil
}

// 获取base指标
func (i *Input) collectBaseMeasurement() ([]inputs.Measurement, error) {
	m := &baseMeasurement{
		i:       i,
		resData: make(map[string]interface{}),
		tags:    make(map[string]string),
		fields:  make(map[string]interface{}),
	}

	m.name = "mysql"
	for key, value := range i.Tags {
		m.tags[key] = value
	}

	if err := m.getStatus(); err != nil {
		return nil, err
	}

	if err := m.getVariables(); err != nil {
		return nil, err
	}

	// 如果没有打开 bin-log，这里可能报错：Error 1381: You are not using binary logging
	// 不过忽略这一错误
	// TODO: if-bin-log-enabled
	if m.resData["log_bin"] == "ON" || m.resData["log_bin"] == "on" {
		_ = m.getLogStats()
	}

	if err := m.submit(); err == nil {
		if len(m.fields) > 0 {
			return []inputs.Measurement{m}, nil
		}
	}

	return nil, nil
}

// 获取innodb指标
func (i *Input) collectInnodbMeasurement() ([]inputs.Measurement, error) {
	return i.getInnodb()
}

// 获取tableSchema指标
func (i *Input) collectTableSchemaMeasurement() ([]inputs.Measurement, error) {
	return i.getTableSchema()
}

// 获取用户指标
func (i *Input) collectUserMeasurement() ([]inputs.Measurement, error) {
	return i.getUserData()
}

// 获取schema指标
func (i *Input) collectSchemaMeasurement() ([]inputs.Measurement, error) {
	x, err := i.getSchemaSize()
	if err != nil {
		return nil, err
	}

	y, err := i.getQueryExecTimePerSchema()
	if err != nil {
		return nil, err
	}

	return append(x, y...), nil
}

func (i *Input) runLog() error {
	if i.Log == nil {
		return nil
	}

	if i.Log.Pipeline == "" {
		i.Log.Pipeline = "mysql.p" // use default
	}

	opt := &tailer.Option{
		Source:            "mysql",
		Service:           "mysql",
		GlobalTags:        i.Tags,
		CharacterEncoding: i.Log.CharacterEncoding,
		Match:             i.Log.Match,
	}

	pl := filepath.Join(datakit.PipelineDir, i.Log.Pipeline)
	if _, err := os.Stat(pl); err != nil {
		l.Warn("%s missing: %s", pl, err.Error())
	} else {
		opt.Pipeline = pl
	}

	var err error
	i.tail, err = tailer.NewTailer(i.Log.Files, opt, i.Log.IgnoreStatus)
	if err != nil {
		l.Error(err)
		return err
	}

	go i.tail.Start()
	return nil
}

// TODO
func (*Input) RunPipeline() {
}

func (i *Input) Run() {
	l = logger.SLogger("mysql")
	i.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, i.Interval.Duration)

	if err := i.runLog(); err != nil {
		io.FeedLastError(inputName, err.Error())
	}

	for { // try until init OK

		select {
		case <-datakit.Exit.Wait():
			return
		default:
		}

		if err := i.initCfg(); err != nil {
			io.FeedLastError(inputName, err.Error())
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	tick := time.NewTicker(i.Interval.Duration)
	defer tick.Stop()

	l.Infof("collecting each %v", i.Interval.Duration)

	i.collectors = []func() ([]inputs.Measurement, error){
		i.collectBaseMeasurement,
		i.collectSchemaMeasurement,
		i.customSchemaMeasurement,
		i.collectTableSchemaMeasurement,
		i.collectUserMeasurement,
	}

	if i.InnoDB {
		i.collectors = append(i.collectors, i.collectInnodbMeasurement)
	}

	for {
		select {
		case <-tick.C:
			l.Debugf("mysql input gathering...")
			i.start = time.Now()
			i.Collect()
		case <-datakit.Exit.Wait():
			if i.tail != nil {
				i.tail.Close()
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
		&innodbMeasurement{},
		&tbMeasurement{},
		&userMeasurement{},
	}
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{Timeout: "10s"}
	})
}
