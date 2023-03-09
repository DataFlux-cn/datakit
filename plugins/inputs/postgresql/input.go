// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package postgresql collects PostgreSQL metrics.
package postgresql

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName                        = "postgresql"
	catalogName                      = "db"
	l                                = logger.DefaultSLogger(inputName)
	_           inputs.ElectionInput = (*Input)(nil)
)

//nolint:lll
const sampleConfig = `
[[inputs.postgresql]]
  ## 服务器地址
  # URI格式
  # postgres://[pqgotest[:password]]@localhost[/dbname]?sslmode=[disable|verify-ca|verify-full]
  # 简单字符串格式
  # host=localhost user=pqgotest password=... sslmode=... dbname=app_production

  address = "postgres://postgres@localhost/test?sslmode=disable"

  ## 配置采集的数据库，默认会采集所有的数据库，当同时设置ignored_databases和databases会忽略databases
  # ignored_databases = ["db1"]
  # databases = ["db1"]

  ## 设置服务器Tag，默认是基于服务器地址生成
  # outputaddress = "db01"

  ## 采集间隔
  # 单位 "ns", "us" (or "µs"), "ms", "s", "m", "h"
  interval = "10s"

  ## Set true to enable election
  election = true

  ## 日志采集
  # [inputs.postgresql.log]
  # files = []
  # pipeline = "postgresql.p"

  ## 自定义Tag
  [inputs.postgresql.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
`

//nolint:lll
const pipelineCfg = `
add_pattern("log_date", "%{YEAR}-%{MONTHNUM}-%{MONTHDAY}%{SPACE}%{HOUR}:%{MINUTE}:%{SECOND}%{SPACE}(?:CST|UTC)")
add_pattern("status", "(LOG|ERROR|FATAL|PANIC|WARNING|NOTICE|INFO)")
add_pattern("session_id", "([.0-9a-z]*)")
add_pattern("application_name", "(\\[%{GREEDYDATA:application_name}?\\])")
add_pattern("remote_host", "(\\[\\[?%{HOST:remote_host}?\\]?\\])")
grok(_, "%{log_date:time}%{SPACE}\\[%{INT:process_id}\\]%{SPACE}(%{WORD:db_name}?%{SPACE}%{application_name}%{SPACE}%{USER:user}?%{SPACE}%{remote_host}%{SPACE})?%{session_id:session_id}%{SPACE}(%{status:status}:)?")

# default
grok(_, "%{log_date:time}%{SPACE}\\[%{INT:process_id}\\]%{SPACE}%{status:status}")

nullif(remote_host, "")
nullif(session_id, "")
nullif(application_name, "")
nullif(user, "")
nullif(db_name, "")

group_in(status, [""], "INFO")

default_time(time)
`

type Rows interface {
	Close() error
	Columns() ([]string, error)
	Next() bool
	Scan(...interface{}) error
}

type Service interface {
	Start() error
	Stop() error
	Query(string) (Rows, error)
	SetAddress(string)
	GetColumnMap(scanner, []string) (map[string]*interface{}, error)
}

type scanner interface {
	Scan(dest ...interface{}) error
}

type Input struct {
	Address          string            `toml:"address"`
	Outputaddress    string            `toml:"outputaddress"`
	IgnoredDatabases []string          `toml:"ignored_databases"`
	Databases        []string          `toml:"databases"`
	Interval         string            `toml:"interval"`
	Tags             map[string]string `toml:"tags"`
	Log              *postgresqllog    `toml:"log"`

	MaxLifetimeDeprecated string `toml:"max_lifetime,omitempty"`

	service      Service
	tail         *tailer.Tailer
	duration     time.Duration
	collectCache []inputs.Measurement
	host         string

	Election bool `toml:"election"`
	pause    bool
	pauseCh  chan bool

	semStop *cliutils.Sem // start stop signal
}

type postgresqllog struct {
	Files             []string `toml:"files"`
	Pipeline          string   `toml:"pipeline"`
	IgnoreStatus      []string `toml:"ignore"`
	CharacterEncoding string   `toml:"character_encoding"`
	MultilineMatch    string   `toml:"multiline_match"`
}

type inputMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	ts       time.Time
	election bool
}

func (m inputMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOptElection())
}

//nolint:lll
func (m inputMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   inputName,
		Fields: postgreFields,
		Tags: map[string]interface{}{
			"server": inputs.NewTagInfo("The server address"),
			"db":     inputs.NewTagInfo("The database name"),
		},
	}
}

func (*Input) Catalog() string {
	return catalogName
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&inputMeasurement{},
	}
}

func (*Input) PipelineConfig() map[string]string {
	return map[string]string{
		"postgresql": pipelineCfg,
	}
}

//nolint:lll
func (ipt *Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		"postgresql": {
			"PostgreSQL log": `2021-05-31 15:23:45.110 CST [74305] test [pgAdmin 4 - DB:postgres] postgres [127.0.0.1] 60b48f01.12241 LOG: statement: 		SELECT psd.*, 2^31 - age(datfrozenxid) as wraparound, pg_database_size(psd.datname) as pg_database_size 		FROM pg_stat_database psd 		JOIN pg_database pd ON psd.datname = pd.datname 		WHERE psd.datname not ilike 'template%' AND psd.datname not ilike 'rdsadmin' 		AND psd.datname not ilike 'azure_maintenance' AND psd.datname not ilike 'postgres'`,
		},
	}
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (ipt *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:  inputName,
			Service: inputName,
			Pipeline: func() string {
				if ipt.Log != nil {
					return ipt.Log.Pipeline
				}
				return ""
			}(),
		},
	}
}

func (ipt *Input) SanitizedAddress() (sanitizedAddress string, err error) {
	var canonicalizedAddress string

	kvMatcher := regexp.MustCompile(`(password|sslcert|sslkey|sslmode|sslrootcert)=\S+ ?`)

	if ipt.Outputaddress != "" {
		return ipt.Outputaddress, nil
	}

	if strings.HasPrefix(ipt.Address, "postgres://") || strings.HasPrefix(ipt.Address, "postgresql://") {
		if canonicalizedAddress, err = parseURL(ipt.Address); err != nil {
			return sanitizedAddress, err
		}
	} else {
		canonicalizedAddress = ipt.Address
	}

	sanitizedAddress = kvMatcher.ReplaceAllString(canonicalizedAddress, "")

	return sanitizedAddress, err
}

func (ipt *Input) executeQuery(query string) error {
	var (
		columns []string
		err     error
	)

	rows, err := ipt.service.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close() //nolint:errcheck

	if columns, err = rows.Columns(); err != nil {
		return err
	}

	for rows.Next() {
		columnMap, err := ipt.service.GetColumnMap(rows, columns)
		if err != nil {
			return err
		}
		err = ipt.accRow(columnMap)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ipt *Input) getDBMetrics() error {
	//nolint:lll
	query := `
	SELECT psd.*, 2^31 - age(datfrozenxid) as wraparound, pg_database_size(psd.datname) as pg_database_size
	FROM pg_stat_database psd
	JOIN pg_database pd ON psd.datname = pd.datname
	WHERE psd.datname not ilike 'template%'   AND psd.datname not ilike 'rdsadmin'
	AND psd.datname not ilike 'azure_maintenance'   AND psd.datname not ilike 'postgres'
	`
	if len(ipt.IgnoredDatabases) != 0 {
		query += fmt.Sprintf(` AND psd.datname NOT IN ('%s')`, strings.Join(ipt.IgnoredDatabases, "','"))
	} else if len(ipt.Databases) != 0 {
		query += fmt.Sprintf(` AND psd.datname IN ('%s')`, strings.Join(ipt.Databases, "','"))
	}

	err := ipt.executeQuery(query)

	return err
}

func (ipt *Input) getBgwMetrics() error {
	query := `
		select * FROM pg_stat_bgwriter
	`
	err := ipt.executeQuery(query)
	return err
}

func (ipt *Input) getConnectionMetrics() error {
	//nolint:lll
	query := `
		WITH max_con AS (SELECT setting::float FROM pg_settings WHERE name = 'max_connections')
		SELECT MAX(setting) AS max_connections, SUM(numbackends)/MAX(setting) AS percent_usage_connections
		FROM pg_stat_database, max_con
	`

	err := ipt.executeQuery(query)
	return err
}

func (ipt *Input) Collect() error {
	var err error

	ipt.service.SetAddress(ipt.Address)
	defer ipt.service.Stop() //nolint:errcheck
	err = ipt.service.Start()
	if err != nil {
		return err
	}

	g := goroutine.NewGroup(goroutine.Option{Name: goroutine.GetInputName(inputName)})

	// collect db metrics
	g.Go(func(ctx context.Context) error {
		err := ipt.getDBMetrics()
		return err
	})

	// collect bgwriter
	g.Go(func(ctx context.Context) error {
		err := ipt.getBgwMetrics()
		return err
	})

	// connection
	g.Go(func(ctx context.Context) error {
		err := ipt.getConnectionMetrics()
		return err
	})

	return g.Wait()
}

func (ipt *Input) accRow(columnMap map[string]*interface{}) error {
	var tagAddress string
	tagAddress, err := ipt.SanitizedAddress()
	if err != nil {
		return err
	}

	tags := map[string]string{"server": tagAddress, "db": "postgres"}
	if ipt.host != "" {
		tags["host"] = ipt.host
	}

	if ipt.Tags != nil {
		for k, v := range ipt.Tags {
			tags[k] = v
		}
	}

	fields := make(map[string]interface{})
	for col, val := range columnMap {
		if col != "datname" {
			if _, isValidCol := postgreFields[col]; !isValidCol {
				continue
			}
		}

		if *val != nil {
			value := *val
			switch trueVal := value.(type) {
			case []uint8:
				if col == "datname" {
					tags["db"] = string(trueVal)
				} else {
					fields[col] = string(trueVal)
				}
			default:
				fields[col] = trueVal
			}
		}
	}
	if len(fields) > 0 {
		ipt.collectCache = append(ipt.collectCache, &inputMeasurement{
			name:     inputName,
			fields:   fields,
			tags:     tags,
			ts:       time.Now(),
			election: ipt.Election,
		})
	}

	return nil
}

func (ipt *Input) RunPipeline() {
	if ipt.Log == nil || len(ipt.Log.Files) == 0 {
		return
	}

	opt := &tailer.Option{
		Source:            inputName,
		Service:           inputName,
		Pipeline:          ipt.Log.Pipeline,
		GlobalTags:        ipt.Tags,
		IgnoreStatus:      ipt.Log.IgnoreStatus,
		CharacterEncoding: ipt.Log.CharacterEncoding,
		MultilinePatterns: []string{ipt.Log.MultilineMatch},
		Done:              ipt.semStop.Wait(),
	}

	var err error
	ipt.tail, err = tailer.NewTailer(ipt.Log.Files, opt)
	if err != nil {
		l.Error(err)
		io.FeedLastError(inputName, err.Error())
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_postgresql"})
	g.Go(func(ctx context.Context) error {
		ipt.tail.Start()
		return nil
	})
}

const (
	maxInterval = 1 * time.Minute
	minInterval = 1 * time.Second
)

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	duration, err := time.ParseDuration(ipt.Interval)
	if err != nil {
		l.Error("invalid interval, %s", err.Error())
	} else if duration <= 0 {
		l.Error("invalid interval, cannot be less than zero")
	}

	if err := ipt.setHostIfNotLoopback(); err != nil {
		l.Errorf("failed to set host: %v", err)
	}

	ipt.duration = config.ProtectedInterval(minInterval, maxInterval, duration)

	tick := time.NewTicker(ipt.duration)

	for {
		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Info("postgresql exit")
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Info("postgresql return")
			return

		case <-tick.C:
			if ipt.pause {
				l.Debugf("not leader, skipped")
				continue
			}

			start := time.Now()
			if err := ipt.Collect(); err != nil {
				io.FeedLastError(inputName, err.Error())
				l.Error(err)
			}

			if len(ipt.collectCache) > 0 {
				err := inputs.FeedMeasurement(inputName, datakit.Metric, ipt.collectCache,
					&io.Option{CollectCost: time.Since(start)})
				if err != nil {
					io.FeedLastError(inputName, err.Error())
					l.Error(err.Error())
				}
				ipt.collectCache = ipt.collectCache[:0]
			}

		case ipt.pause = <-ipt.pauseCh:
			// nil
		}
	}
}

func (ipt *Input) exit() {
	if ipt.tail != nil {
		ipt.tail.Close()
		l.Info("postgresql log exit")
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (ipt *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func (ipt *Input) setHostIfNotLoopback() error {
	uu, err := url.Parse(ipt.Address)
	if err != nil {
		return err
	}
	var host string
	h, _, err := net.SplitHostPort(uu.Host)
	if err == nil {
		host = h
	} else {
		host = uu.Host
	}
	if host != "localhost" && !net.ParseIP(host).IsLoopback() {
		ipt.host = host
	}
	return nil
}

func parseURL(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return "", fmt.Errorf("invalid connection protocol: %s", u.Scheme)
	}

	var kvs []string
	escaper := strings.NewReplacer(` `, `\ `, `'`, `\'`, `\`, `\\`)
	accrue := func(k, v string) {
		if v != "" {
			kvs = append(kvs, k+"="+escaper.Replace(v))
		}
	}

	if u.User != nil {
		v := u.User.Username()
		accrue("user", v)

		v, _ = u.User.Password()
		accrue("password", v)
	}

	if host, port, err := net.SplitHostPort(u.Host); err != nil {
		accrue("host", u.Host)
	} else {
		accrue("host", host)
		accrue("port", port)
	}

	if u.Path != "" {
		accrue("dbname", u.Path[1:])
	}

	q := u.Query()
	for k := range q {
		accrue(k, q.Get(k))
	}

	sort.Strings(kvs)
	return strings.Join(kvs, " "), nil
}

var maxPauseCh = inputs.ElectionPauseChannelLength

func NewInput(service Service) *Input {
	input := &Input{
		Interval: "10s",
		pauseCh:  make(chan bool, maxPauseCh),
		Election: true,

		semStop: cliutils.NewSem(),
	}
	input.service = service
	return input
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		service := &SQLService{
			MaxIdle:     1,
			MaxOpen:     1,
			MaxLifetime: time.Duration(0),
		}
		input := NewInput(service)
		return input
	})
}
