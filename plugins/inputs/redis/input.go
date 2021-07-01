package redis

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-redis/redis/v8"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	maxInterval = 30 * time.Minute
	minInterval = 15 * time.Second
)

var (
	inputName   = "redis"
	catalogName = "db"
	l           = logger.DefaultSLogger("redis")
)

type redislog struct {
	Files             []string `toml:"files"`
	Pipeline          string   `toml:"pipeline"`
	IgnoreStatus      []string `toml:"ignore"`
	CharacterEncoding string   `toml:"character_encoding"`
	Match             string   `toml:"match"`
}

type Input struct {
	Host              string        `toml:"host"`
	Port              int           `toml:"port"`
	UnixSocketPath    string        `toml:"unix_socket_path"`
	DB                int           `toml:"db"`
	Password          string        `toml:"password"`
	Timeout           string        `toml:"connect_timeout"`
	timeoutDuration   time.Duration `toml:"-"`
	Service           string        `toml:"service"`
	SocketTimeout     int           `toml:"socket_timeout"`
	Interval          datakit.Duration
	Keys              []string                               `toml:"keys"`
	WarnOnMissingKeys bool                                   `toml:"warn_on_missing_keys"`
	CommandStats      bool                                   `toml:"command_stats"`
	Slowlog           bool                                   `toml:"slow_log"`
	SlowlogMaxLen     int                                    `toml:"slowlog-max-len"`
	Tags              map[string]string                      `toml:"tags"`
	client            *redis.Client                          `toml:"-"`
	Addr              string                                 `toml:"-"`
	Log               *redislog                              `toml:"log"`
	tail              *tailer.Tailer                         `toml:"-"`
	start             time.Time                              `toml:"-"`
	collectors        []func() ([]inputs.Measurement, error) `toml:"-"`
}

func (i *Input) initCfg() error {
	var err error
	i.timeoutDuration, err = time.ParseDuration(i.Timeout)
	if err != nil {
		i.timeoutDuration = 10 * time.Second
	}

	i.Addr = fmt.Sprintf("%s:%d", i.Host, i.Port)

	client := redis.NewClient(&redis.Options{
		Addr:     i.Addr,
		Password: i.Password, // no password set
		DB:       i.DB,       // use default DB
	})

	if i.SlowlogMaxLen == 0 {
		i.SlowlogMaxLen = 128
	}

	i.client = client

	// ping (todo)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = client.Ping(ctx).Result()

	if err != nil {
		return err
	}

	i.Tags["server"] = i.Addr
	i.Tags["service_name"] = i.Service

	return nil
}

func (i *Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		"redis": pipelineCfg,
	}
	return pipelineMap
}

func (i *Input) Collect() error {
	for _, f := range i.collectors {
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

func (i *Input) collectInfoMeasurement() ([]inputs.Measurement, error) {
	var collectCache []inputs.Measurement

	m := &infoMeasurement{
		cli:     i.client,
		resData: make(map[string]interface{}),
		tags:    make(map[string]string),
		fields:  make(map[string]interface{}),
	}

	m.name = "redis_info"

	for key, value := range i.Tags {
		m.tags[key] = value
	}

	// get data
	if err := m.getData(); err != nil {
		return nil, err
	}

	// build line data
	if err := m.submit(); err != nil {
		return nil, err
	}

	if len(m.fields) > 0 {
		collectCache = append(collectCache, m)
	}

	return collectCache, nil
}

func (i *Input) collectBigKeyMeasurement() ([]inputs.Measurement, error) {
	keys, err := i.getKeys()
	if err != nil {
		return nil, err
	}

	return i.getData(keys)
}

// 数据源获取数据
func (i *Input) collectClientMeasurement() ([]inputs.Measurement, error) {
	ctx := context.Background()
	list, err := i.client.ClientList(ctx).Result()
	if err != nil {
		l.Error("client list get error,", err)
		return nil, err
	}

	return i.parseClientData(list)
}

// 数据源获取数据
func (i *Input) collectCommandMeasurement() ([]inputs.Measurement, error) {
	ctx := context.Background()
	list, err := i.client.Info(ctx, "commandstats").Result()
	if err != nil {
		l.Error("command stats error,", err)
		return nil, err
	}

	return i.parseCommandData(list)
}

func (i *Input) collectSlowlogMeasurement() ([]inputs.Measurement, error) {
	return i.getSlowData()
}

func (i *Input) runLog() error {
	if i.Log == nil {
		return nil
	}

	if i.Log.Pipeline == "" {
		i.Log.Pipeline = "redis.p" // use default
	}

	opt := &tailer.Option{
		Source:            "redis",
		Service:           "redis",
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
	l = logger.SLogger("redis")

	i.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, i.Interval.Duration)

	if err := i.runLog(); err != nil {
		io.FeedLastError(inputName, err.Error())
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			return
		default:
		}

		if err := i.initCfg(); err != nil {
			io.FeedLastError(inputName, err.Error())
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}

	tick := time.NewTicker(i.Interval.Duration)
	defer tick.Stop()

	i.collectors = []func() ([]inputs.Measurement, error){
		i.collectInfoMeasurement,
		i.collectClientMeasurement,
		i.collectCommandMeasurement,
		i.collectDBMeasurement,
		i.collectSlowlogMeasurement,
	}

	if len(i.Keys) > 0 {
		i.collectors = append(i.collectors, i.collectBigKeyMeasurement)
	}

	for {
		select {
		case <-tick.C:
			l.Debugf("redis input gathering...")
			i.start = time.Now()
			i.Collect()

		case <-datakit.Exit.Wait():
			if i.tail != nil {
				i.tail.Close()
				l.Info("redis log exit")
			}
			l.Info("redis exit")
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
	return []inputs.Measurement{
		&infoMeasurement{},
		&clientMeasurement{},
		&commandMeasurement{},
		&slowlogMeasurement{},
		&bigKeyMeasurement{},
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{Timeout: "10s"}
	})
}
