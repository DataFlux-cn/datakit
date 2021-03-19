package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/go-redis/redis"
	"go.uber.org/zap"
	yaml "gopkg.in/yaml.v2"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/utils"
)

var (
	l = logger.DefaultSLogger("config")

	C = Config{
		Influx: InfluxCfg{
			ReadTimeOut:  30,
			WriteTimeOut: 30,
			EnableGrant:  true,
		},

		Database: DatabaseCfg{
			Dialect: `mysql`,
		},

		Redis: RedisCfg{
			Host: `cloudcare-kodo-redis.db.svc.cluster.local:6379`,
			Db:   0,
		},

		LogConfig: LogCfg{
			LogFile:    `/logdata/log`,
			Level:      `info`,
			JSONFormat: true,
			GinLogFile: `/logdata/gin.log`,
		},

		NSQ: NSQCfg{
			Lookupd:   `http://nsqlookupd.middleware.svc.cluster.local:4161`,
			RefreshAt: 300,
		},

		Stat: StatCfg{
			SlowWrite: 3.0, // second
			SlowQuery: 1.0,
		},

		Global: GlobalCfg{
			EnableInnerApi: true,
			StatsOn:        256,
			Listen:         `:9527`,
			Workers:        8,
			LogWorkers:     8,
			Dataway:        "http://internal-dataway.utils:9528",
			EsConsumer:     false,
			RetryTimes:     3600 * 24,
			MaxWrites:      1000,
			MeterInterval:  60, //minute
			SysWsUUID:      `wksp_system`,
			DataMigration:  false,
		},

		DQL: DQLCfg{
			MaxDuration:  366*24*time.Hour + time.Hour, // 366 day + 1 hour
			MaxLimit:     5000,
			DefaultLimit: 1000,
		},

		SourceEs: SourceEsCfg{
			Host:   ``,
			User:   ``,
			Passwd: ``,
		},

		Ck: CkCfg{
			ReadTimeOut:  60,
			WriteTimeOut: 30,
			Host:         ``,
			User:         ``,
			Passwd:       ``,
			ClusterName:  `default`,
		},

		Es: EsCfg{
			Host:    ``,
			User:    ``,
			Passwd:  ``,
			TimeOut: `30s`,
		},

		RpConfig: map[string][2]string{
			`rp6`:      {`25920h`, `720h`}, // 3 year
			`rp5`:      {`8640h`, `720h`},  // 1 year
			`rp4`:      {`4320h`, `720h`},  // 6 month
			`rp3`:      {`2160h`, `1w`},    // 3 month
			`rp2`:      {`720h`, `1w`},     // 1 month
			`rp1`:      {`168h`, `1d`},     // 1 week
			`rp0`:      {`24h`, `6h`},      // 1 day
			`rp_2160h`: {`2160h`, `1w`},
			`autogen`:  {`25920h`, `720h`},
		},
		Ws: WsConfig{
			Path: "/v1/ws/datakit",
		},
		ShardDurationConfig: map[string]string{
			`25920h`:  `720h`, // 3 year
			`8640h`:   `720h`, // 1 year
			`4320h`:   `720h`, // 6 month
			`2160h`:   `1w`,   // 3 month
			`720h`:    `1w`,   // 1 month
			`336h`:    `2d`,   // 2 weeks
			`168h`:    `1d`,   // 1 week
			`24h`:     `6h`,   // 1 day
			`8760h`:   `720h`,
			`262800h`: `720h`,
			// `1095d`:  `720h`, // 3 year
			// `365d`:   `720h`, // 1 year
			// `180d`:   `720h`, // 6 month
			// `90d`:    `1w`,   // 3 month
			// `30d`:    `1w`,   // 1 month
			// `14d`:    `2d`,   // 2 weeks
			// `7d`:     `1d`,   // 1 week
			// `1d`:     `6h`,   // 1 day
		},
	}

	Redis *redis.Client
)

type DatabaseCfg struct {
	Dialect    string `yaml:"db_dialect"`
	Connection string `yaml:"connection"`
}

type NSQCfg struct {
	Lookupd   string `yaml:"lookupd"`
	RefreshAt int    `yaml:"refresh_at"`
}

type RedisCfg struct {
	Host string `yaml:"host"`
	Pass string `yaml:"password"`
	Db   int    `yaml:"db"`
}

type InfluxCfg struct {
	ReadTimeOut  int64  `yaml:"read_timeout"`
	WriteTimeOut int64  `yaml:"write_timeout"`
	DefaultRP    string `yaml:"default_rp"`
	UserAgent    string `yaml:"user_agent"`
	EnableGrant  bool   `yaml:"enable_grant"`
	EnableGZ     bool   `yaml:"enable_gz"`

	DisableWrite bool `yaml:"disable_write"` // for test
}

type LogCfg struct {
	LogFile    string `yaml:"log_file"`
	Level      string `yaml:"level,omitempty"`
	JSONFormat bool   `yaml:"json_format,omitempty"`
	ShortFile  bool   `yaml:"short_file,omitempty"`
	GinLogFile string `yaml:"gin_log_file,omitempty"`
	Rl         *zap.Logger
}

type StatCfg struct {
	SlowWrite float64 `yaml:"slow_write"`
	SlowQuery float64 `yaml:"slow_query"`
	RP        string  `yaml:"rp"`
}

type GlobalCfg struct {
	EnableInnerApi bool   `yaml:"enable_inner_api,omitempty"`
	StatsOn        int    `yaml:"stats_on,omitempty"`
	Listen         string `yaml:"listen"`
	Workers        int    `yaml:"workers,omitempty"`
	LogWorkers     int    `yaml:"log_workers,omitempty"`
	Dataway        string `yaml:"dataway,omitempty"`
	EsConsumer     bool   `yaml:"es_consumer,omitempty"`
	RetryTimes     int64  `yaml:"retry_time_seconds,omitempty"`
	SysDBUUID      string `yaml:"sys_db_uuid,omitempty"`
	SysWsUUID      string `yaml:"sys_ws_uuid,omitempty"`
	MeterInterval  int    `yaml:"meter_interval,omitempty"`

	MaxWrites int `yaml:"max_writes"`

	DataMigration bool `yaml:"data_migration"`

	// each license should only used on 1 dataway, if any dataway mis-configured
	// license(legal) used on other dataway, kodo will refuse it's request.
	EnableLicenseDataWayBinding bool `yaml:"enable_license_dataWay_binding"`
}

type DQLCfg struct {
	MaxDurationStr string        `yaml:"max_duration"`
	MaxDuration    time.Duration `yaml:"-"`
	MaxLimit       int64         `yaml:"max_limit"`
	DefaultLimit   int64         `yaml:"default_limit"`
}

type CkCfg struct {
	ReadTimeOut  int64 `yaml:"read_timeout"`
	WriteTimeOut int64 `yaml:"write_timeout"`

	Host        string `yaml:"host"`
	User        string `yaml:"user"`
	Passwd      string `yaml:"password"`
	ClusterName string `yaml:"cluster_name"`
}

type SecretCfg struct {
	EncryptKey string `yaml:"encrypt_key"`
}

type Config struct {
	Influx              InfluxCfg            `yaml:"influxdb"`
	Database            DatabaseCfg          `yaml:"database"`
	Redis               RedisCfg             `yaml:"redis"`
	LogConfig           LogCfg               `yaml:"log"`
	RpConfig            map[string][2]string `yaml:"global_rp"`
	ShardDurationConfig map[string]string    `yaml:"shard_duration_cfg"`
	Func                FuncCfg              `yaml:"func"`
	NSQ                 NSQCfg               `yaml:"nsq"`
	Global              GlobalCfg            `yaml:"global"`
	DQL                 DQLCfg               `yaml:"dql"`
	Ck                  CkCfg                `yaml:"ck"`
	Secret              SecretCfg            `yaml:"secret"`
	Stat                StatCfg              `yaml:"stat"`
	Es                  EsCfg                `yaml:"es"`
	Ws                  WsConfig             `yaml:"ws_server"`
	WsKoDo              WsKoDo               `yaml:"kodo"`
	SourceEs            SourceEsCfg          `yaml:"source_es,omitempty"`
}

type EsCfg struct {
	Host    string `yaml:"host"`
	User    string `yaml:"user"`
	Passwd  string `yaml:"password"`
	Enable  bool   `yaml:"enable"`
	TimeOut string `yaml:"timeout"`
}

type FuncCfg struct {
	Host   string `yaml:"host"`
	Enable bool   `yaml:"enable"`
}

type WsConfig struct {
	Bind    string `yaml:"bind"`
	Path    string `yaml:"path"`
	TimeOut string `yaml:"time_out"`
}

type WsKoDo struct {
	Host    string `yaml:"remote_host"`
	TimeOut string `yaml:"time_out"`
}

type SourceEsCfg struct {
	Host   string `yaml:"host"`
	User   string `yaml:"user"`
	Passwd string `yaml:"password"`
}

type SourceInfluxCfg struct {
	Host   string `yaml:"host"`
	User   string `yaml:"user"`
	Passwd string `yaml:"password"`
}

type SourceMysqlCfg struct {
	Dialect    string `yaml:"db_dialect"`
	Connection string `yaml:"connection"`
}

func DumpConfig(cfg interface{}) ([]byte, error) {
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func LoadConfig(f string, cfg interface{}) error {
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return err
	}

	return nil
}

func ApplyConfig() {
	var err error

	// set kodo log
	dir, err := filepath.Abs(filepath.Dir(C.LogConfig.LogFile))
	if err != nil {
		l.Fatal(err)
	}

	exist, err := utils.PathExists(dir)
	if err != nil {
		l.Fatalf("get dir error![%v]\n", err)
	}

	l.Debugf("dir %s  %v", dir, exist)
	if !exist {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			l.Fatalf("create dir error![%v]\n", err)
		}
	}

	// parse duration
	if C.DQL.MaxDurationStr != "" {
		du, err := utils.ParseDuration(C.DQL.MaxDurationStr)
		if err != nil {
			// if parse failed, use default
			l.Error(err)
		} else {
			C.DQL.MaxDuration = du
		}
	}

	// init redis conncetions
	redisOpt := redis.Options{
		Addr: C.Redis.Host,
		DB:   C.Redis.Db,
	}

	if len(C.Redis.Pass) > 0 {
		//l.Debugf("set redis password: %s", C.Redis.Pass)
		redisOpt.Password = C.Redis.Pass
	}

	Redis = redis.NewClient(&redisOpt)
	for {
		_, err := Redis.Ping().Result()
		if err != nil {
			l.Errorf("%s, retry...", err)
			time.Sleep(time.Second)
		} else {
			l.Info("[info] connect to redis ok.")
			break
		}
	}
}
