package datakit

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	bstoml "github.com/BurntSushi/toml"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
)

var (
	IntervalDuration = 10 * time.Second

	DefaultWebsocketPath = "/v1/datakit/ws"

	Cfg = DefaultConfig()
)

func DefaultConfig() *Config {
	return &Config{ //nolint:dupl
		MainCfg: &MainConfig{
			GlobalTags:      map[string]string{},
			flushInterval:   Duration{Duration: time.Second * 10},
			Interval:        "10s",
			MaxPostInterval: "15s", // add 5s plus for network latency
			StrictMode:      false,

			HTTPBind: "0.0.0.0:9529",

			LogLevel:  "info",
			Log:       filepath.Join(InstallDir, "datakit.log"),
			LogRotate: 32,
			LogUpload: false,
			GinLog:    filepath.Join(InstallDir, "gin.log"),

			RoundInterval: false,
			TelegrafAgentCfg: &TelegrafCfg{
				Interval:                   "10s",
				RoundInterval:              true,
				MetricBatchSize:            1000,
				MetricBufferLimit:          100000,
				CollectionJitter:           "0s",
				FlushInterval:              "10s",
				FlushJitter:                "0s",
				Precision:                  "ns",
				Debug:                      false,
				Quiet:                      false,
				LogTarget:                  "file",
				Logfile:                    filepath.Join(TelegrafDir, "agent.log"),
				LogfileRotationMaxArchives: 5,
				LogfileRotationMaxSize:     "32MB",
				OmitHostname:               true, // do not append host tag
			},
		},
	}
}

//用于支持在datakit.conf中加入telegraf的agent配置
type TelegrafCfg struct {
	Interval                   string `toml:"interval"`
	RoundInterval              bool   `toml:"round_interval"`
	Precision                  string `toml:"precision"`
	CollectionJitter           string `toml:"collection_jitter"`
	FlushInterval              string `toml:"flush_interval"`
	FlushJitter                string `toml:"flush_jitter"`
	MetricBatchSize            int    `toml:"metric_batch_size"`
	MetricBufferLimit          int    `toml:"metric_buffer_limit"`
	FlushBufferWhenFull        bool   `toml:"-"`
	UTC                        bool   `toml:"utc"`
	Debug                      bool   `toml:"debug"`
	Quiet                      bool   `toml:"quiet"`
	LogTarget                  string `toml:"logtarget"`
	Logfile                    string `toml:"logfile"`
	LogfileRotationInterval    string `toml:"logfile_rotation_interval"`
	LogfileRotationMaxSize     string `toml:"logfile_rotation_max_size"`
	LogfileRotationMaxArchives int    `toml:"logfile_rotation_max_archives"`
	OmitHostname               bool   `toml:"omit_hostname"`
}

type Config struct {
	MainCfg      *MainConfig
	InputFilters []string
}

type DataWayCfg struct {
	URL     string `toml:"url"`
	Proxy bool   `toml:"proxy,omitempty"`
	WSURL   string `toml:"ws_url"`
	Timeout string `toml:"timeout"`
	Heartbeat string `toml:"heartbeat"`

	DeprecatedHost   string `toml:"host,omitempty"`
	DeprecatedScheme string `toml:"scheme,omitempty"`
	DeprecatedToken  string `toml:"token,omitempty"`

	host      string
	scheme    string
	WSToken   string
	urlValues url.Values

	wspath      string
	wshost      string
	wsscheme    string
	wsUrlValues url.Values
}

func (dc *DataWayCfg) DeprecatedMetricURL() string {
	if dc.Proxy {
		return fmt.Sprintf("%s://%s%s?%s",
			dc.scheme,
			dc.host,
			"/proxy",
			"category=/v1/write/metric")
	}

	return fmt.Sprintf("%s://%s%s?%s",
		dc.scheme,
		dc.host,
		"/v1/write/metrics",
		dc.urlValues.Encode())
}

func (dc *DataWayCfg) MetricURL() string {

	if dc.Proxy {
		return fmt.Sprintf("%s://%s%s?%s",
			dc.scheme,
			dc.host,
			"/proxy",
			"category=/v1/write/metric")
	}

	return fmt.Sprintf("%s://%s%s?%s",
		dc.scheme,
		dc.host,
		"/v1/write/metric",
		dc.urlValues.Encode())
}

func (dc *DataWayCfg) ObjectURL() string {

	if dc.Proxy {
		return fmt.Sprintf("%s://%s%s?%s",
			dc.scheme,
			dc.host,
			"/proxy",
			"category=/v1/write/object")
	}

	return fmt.Sprintf("%s://%s%s?%s",
		dc.scheme,
		dc.host,
		"/v1/write/object",
		dc.urlValues.Encode())
}

func (dc *DataWayCfg) LoggingURL() string {

	if dc.Proxy {
		return fmt.Sprintf("%s://%s%s?%s",
			dc.scheme,
			dc.host,
			"/proxy",
			"category=/v1/write/logging")
	}

	return fmt.Sprintf("%s://%s%s?%s",
		dc.scheme,
		dc.host,
		"/v1/write/logging",
		dc.urlValues.Encode())
}

func (dc *DataWayCfg) TracingURL() string {
	if dc.Proxy {
		return fmt.Sprintf("%s://%s%s?%s",
			dc.scheme,
			dc.host,
			"/proxy",
			"category=/v1/write/tracing")
	}

	return fmt.Sprintf("%s://%s%s?%s",
		dc.scheme,
		dc.host,
		"/v1/write/tracing",
		dc.urlValues.Encode())
}

func (dc *DataWayCfg) KeyEventURL() string {

	if dc.Proxy {
		return fmt.Sprintf("%s://%s%s?%s",
			dc.scheme,
			dc.host,
			"/proxy",
			"category=/v1/write/keyevent")
	}

	return fmt.Sprintf("%s://%s%s?%s",
		dc.scheme,
		dc.host,
		"/v1/write/keyevent",
		dc.urlValues.Encode())
}

func (dc *DataWayCfg) BuildWSURL(mc *MainConfig) *url.URL {

	rawQuery := fmt.Sprintf("id=%s&version=%s&os=%s&arch=%s&token=%s&heartbeatconf=%s",
		mc.UUID, git.Version, runtime.GOOS, runtime.GOARCH, dc.WSToken,dc.Heartbeat)

	return &url.URL{
		Scheme:   dc.wsscheme,
		Host:     dc.wshost,
		Path:     dc.wspath,
		RawQuery: rawQuery,
	}
}

func (dc *DataWayCfg) tcpaddr(scheme, addr string) (string, error) {
	tcpaddr := addr
	if _, _, err := net.SplitHostPort(tcpaddr); err != nil {
		switch scheme {
		case "http", "ws":
			tcpaddr += ":80"
		case "https", "wss":
			tcpaddr += ":443"
		}

		if _, _, err := net.SplitHostPort(tcpaddr); err != nil {
			l.Errorf("net.SplitHostPort(): %s", err)
			return "", err
		}
	}

	return tcpaddr, nil
}

func (dc *DataWayCfg) Test() error {

	wsaddr, err := dc.tcpaddr(dc.wsscheme, dc.wshost)
	if err != nil {
		return err
	}

	httpaddr, err := dc.tcpaddr(dc.scheme, dc.host)
	if err != nil {
		return err
	}

	for _, h := range []string{wsaddr, httpaddr} {
		conn, err := net.DialTimeout("tcp", h, time.Second*5)
		if err != nil {
			l.Errorf("TCP dial host `%s' failed: %s", dc.host, err.Error())
			return err
		}

		if err := conn.Close(); err != nil {
			l.Errorf("Close(): %s, ignored", err.Error())
		}
	}

	return nil
}

func (dc *DataWayCfg) addToken(tkn string) {
	if dc.urlValues == nil {
		dc.urlValues = url.Values{}
	}

	if dc.urlValues.Get("token") == "" {
		l.Debugf("use old token %s", dc.DeprecatedToken)
		dc.urlValues.Set("token", dc.DeprecatedToken)
	}
}

func ParseDataway(httpurl, wsurl string) (*DataWayCfg, error) {

	dwcfg := &DataWayCfg{
		Timeout: "30s",
	}
	if httpurl == "" {
		return nil, fmt.Errorf("empty dataway HTTP endpoint")
	}

	if u, err := url.Parse(httpurl); err == nil {
		dwcfg.scheme = u.Scheme
		dwcfg.urlValues = u.Query()
		dwcfg.host = u.Host
		if u.Path == "/proxy" {
			l.Debugf("datakit proxied by %s", u.Host)
			dwcfg.Proxy = true
		} else {
			u.Path = ""
		}
	}

	var urls []string
	if wsurl == "" {
		urls = []string{httpurl}
	} else {
		urls = []string{httpurl, wsurl}
	}

	for _, s := range urls {
		u, err := url.Parse(s)
		if err != nil {
			l.Errorf("parse url %s failed: %s", s, err.Error())
			return nil, err
		}

		switch u.Scheme {
		case "ws", "wss":
			dwcfg.WSURL = u.String()

			dwcfg.wsscheme = u.Scheme
			dwcfg.wsUrlValues = u.Query()
			dwcfg.wshost = u.Host
			dwcfg.wspath = u.Path
			if dwcfg.wspath == "" {
				dwcfg.wspath = DefaultWebsocketPath
			}

			dwcfg.WSToken = dwcfg.wsUrlValues.Get("token")
			if dwcfg.WSToken == "" {
				l.Warn("ws token missing, ignored")
			}

		case "http", "https":
			dwcfg.URL = u.String()

			dwcfg.scheme = u.Scheme
			dwcfg.urlValues = u.Query()
			dwcfg.host = u.Host

		default:
			l.Errorf("unknown scheme %s", u.Scheme)
			return nil, fmt.Errorf("unknown scheme")
		}
	}

	if wsurl == "" { // for old version dataway, no websocket available
		switch dwcfg.scheme {
		case "http":
			dwcfg.wsscheme = "ws"
		case "https":
			dwcfg.wsscheme = "wss"
		}

		dwcfg.wshost = dwcfg.host
		dwcfg.wsUrlValues = dwcfg.urlValues // ws default share same urlvalues with http
		dwcfg.wspath = DefaultWebsocketPath
		dwcfg.WSURL = fmt.Sprintf("%s://%s%s?%s", dwcfg.wsscheme, dwcfg.wshost, dwcfg.wspath, dwcfg.wsUrlValues.Encode())
	}


	return dwcfg, nil
}

type MainConfig struct {
	UUID     string      `toml:"uuid"`
	Name     string      `toml:"name"`
	DataWay  *DataWayCfg `toml:"dataway,omitempty"`
	HTTPBind string      `toml:"http_server_addr"`

	// For old datakit verison conf, there may exist these fields,
	// if these tags missing, TOML will parse error
	DeprecatedFtGateway        string `toml:"ftdataway,omitempty"`
	DeprecatedIntervalDuration int64  `toml:"interval_duration,omitempty"`
	DeprecatedConfigDir        string `toml:"config_dir,omitempty"`
	DeprecatedOmitHostname     bool   `toml:"omit_hostname,omitempty"`

	Log       string `toml:"log"`
	LogLevel  string `toml:"log_level"`
	LogRotate int    `toml:"log_rotate,omitempty"`
	LogUpload bool   `toml:"log_upload"`

	GinLog               string            `toml:"gin_log"`
	MaxPostInterval      string            `toml:"max_post_interval"`
	GlobalTags           map[string]string `toml:"global_tags"`
	RoundInterval        bool
	StrictMode           bool   `toml:"strict_mode,omitempty"`
	EnablePProf          bool   `toml:"enable_pprof,omitempty"`
	Interval             string `toml:"interval"`
	flushInterval        Duration
	OutputFile           string       `toml:"output_file"`
	Hostname             string       `toml:"hostname,omitempty"`
	DefaultEnabledInputs []string     `toml:"default_enabled_inputs"`
	InstallDate          time.Time    `toml:"install_date,omitempty"`
	TelegrafAgentCfg     *TelegrafCfg `toml:"agent"`
}

func InitDirs() {
	for _, dir := range []string{TelegrafDir, DataDir, LuaDir, ConfdDir} {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			l.Fatalf("create %s failed: %s", dir, err)
		}
	}
}

func (c *Config) LoadMainConfig(p string) error {
	cfgdata, err := ioutil.ReadFile(p)
	if err != nil {
		l.Errorf("read main cfg %s failed: %s", p, err.Error())
		return err
	}

	return c.doLoadMainConfig(cfgdata)
}

func (c *Config) InitCfg(p string) error {

	if c.MainCfg.Hostname == "" {
		c.setHostname()
	}

	if mcdata, err := TomlMarshal(c.MainCfg); err != nil {
		l.Errorf("TomlMarshal(): %s", err.Error())
		return err
	} else {
		if err := ioutil.WriteFile(p, mcdata, 0600); err != nil {
			l.Errorf("error creating %s: %s", p, err)
			return err
		}
	}

	return nil
}

func (c *Config) doLoadMainConfig(cfgdata []byte) error {
	_, err := bstoml.Decode(string(cfgdata), c.MainCfg)
	if err != nil {
		l.Errorf("unmarshal main cfg failed %s", err.Error())
		return err
	}

	if c.MainCfg.TelegrafAgentCfg.LogTarget == "file" && c.MainCfg.TelegrafAgentCfg.Logfile == "" {
		c.MainCfg.TelegrafAgentCfg.Logfile = filepath.Join(InstallDir, "embed", "agent.log")
	}

	if c.MainCfg.OutputFile != "" {
		OutputFile = c.MainCfg.OutputFile
	}

	if c.MainCfg.Hostname == "" {
		c.setHostname()
	}

	if c.MainCfg.DataWay.URL == "" {
		l.Fatal("dataway URL not set")
	}

	dw, err := ParseDataway(c.MainCfg.DataWay.URL, c.MainCfg.DataWay.WSURL)
	if err != nil {
		return err
	}

	heart,err := time.ParseDuration(c.MainCfg.DataWay.Heartbeat)
	if err != nil {
		l.Error("ws heartbeat config err:",err.Error())
		c.MainCfg.DataWay.Heartbeat = "30s"
	}
	maxHeart,_:= time.ParseDuration("5m")
	minHeart,_:= time.ParseDuration("30s")
	if heart > maxHeart{
		c.MainCfg.DataWay.Heartbeat = "5m"
	}
	if heart < minHeart{
		c.MainCfg.DataWay.Heartbeat = "30s"
	}

	dw.Heartbeat = c.MainCfg.DataWay.Heartbeat

	c.MainCfg.DataWay = dw

	if c.MainCfg.DataWay.DeprecatedToken != "" { // compatible with old dataway config
		c.MainCfg.DataWay.addToken(c.MainCfg.DataWay.DeprecatedToken)
	}

	if c.MainCfg.MaxPostInterval != "" {
		du, err := time.ParseDuration(c.MainCfg.MaxPostInterval)
		if err != nil {
			l.Warnf("parse %s failed: %s, set default to 15s", c.MainCfg.MaxPostInterval)
			du = time.Second * 15
		}
		MaxLifeCheckInterval = du
	}

	if c.MainCfg.Interval != "" {
		du, err := time.ParseDuration(c.MainCfg.Interval)
		if err != nil {
			l.Warnf("parse %s failed: %s, set default to 10s", c.MainCfg.Interval)
			du = time.Second * 10
		}
		IntervalDuration = du
	}

	c.MainCfg.TelegrafAgentCfg.Debug = strings.EqualFold(strings.ToLower(c.MainCfg.LogLevel), "debug")

	// reset global tags
	for k, v := range c.MainCfg.GlobalTags {
		switch strings.ToLower(v) {
		case `$datakit_hostname`:
			if c.MainCfg.Hostname == "" {
				c.setHostname()
			}

			c.MainCfg.GlobalTags[k] = c.MainCfg.Hostname
			l.Debugf("set global tag %s: %s", k, c.MainCfg.Hostname)

		case `$datakit_ip`:
			c.MainCfg.GlobalTags[k] = "unavailable"

			if ipaddr, err := LocalIP(); err != nil {
				l.Errorf("get local ip failed: %s", err.Error())
			} else {
				l.Debugf("set global tag %s: %s", k, ipaddr)
				c.MainCfg.GlobalTags[k] = ipaddr
			}

		case `$datakit_uuid`, `$datakit_id`:
			c.MainCfg.GlobalTags[k] = c.MainCfg.UUID
			l.Debugf("set global tag %s: %s", k, c.MainCfg.UUID)

		default:
			// pass
		}
	}

	return nil
}

func (c *Config) setHostname() {
	hn, err := os.Hostname()
	if err != nil {
		l.Errorf("get hostname failed: %s", err.Error())
	} else {
		c.MainCfg.Hostname = hn
		l.Infof("set hostname to %s", hn)
	}
}

func (c *Config) EnableDefaultsInputs(inputlist string) {
	elems := strings.Split(inputlist, ",")
	if len(elems) == 0 {
		return
	}

	for _, name := range elems {
		c.MainCfg.DefaultEnabledInputs = append(c.MainCfg.DefaultEnabledInputs, name)
	}
}

func (c *Config) LoadEnvs(mcp string) error {
	if !Docker { // only accept configs from ENV within docker
		return nil
	}

	enableInputs := os.Getenv("ENV_ENABLE_INPUTS")
	if enableInputs != "" {
		c.EnableDefaultsInputs(enableInputs)
	}

	globalTags := os.Getenv("ENV_GLOBAL_TAGS")
	if globalTags != "" {
		c.MainCfg.GlobalTags = ParseGlobalTags(globalTags)
	}

	loglvl := os.Getenv("ENV_LOG_LEVEL")
	if loglvl != "" {
		c.MainCfg.LogLevel = loglvl
	}

	dwcfg := os.Getenv("ENV_DATAWAY")
	if dwcfg != "" {

		parts := strings.Split(dwcfg, ";")
		if len(parts) != 2 {
			return fmt.Errorf("invalid ENV_DATAWAY")
		}

		dw, err := ParseDataway(parts[0], parts[1])
		if err != nil {
			return err
		}

		if err := dw.Test(); err != nil {
			return err
		}

		c.MainCfg.DataWay = dw
	}

	dkhost := os.Getenv("ENV_HOSTNAME")
	if dkhost != "" {
		l.Debugf("set hostname to %s from ENV", dkhost)
		c.MainCfg.Hostname = dkhost
	} else {
		c.setHostname()
	}

	c.MainCfg.Name = os.Getenv("ENV_NAME")

	if fi, err := os.Stat(mcp); err != nil || fi.Size() == 0 { // create the main config
		if c.MainCfg.UUID == "" { // datakit.conf not exit: we have to create new datakit with new UUID
			c.MainCfg.UUID = cliutils.XID("dkid_")
		}

		c.MainCfg.InstallDate = time.Now()

		cfgdata, err := TomlMarshal(c.MainCfg)
		if err != nil {
			l.Errorf("failed to build main cfg %s", err)
			return err
		}

		l.Debugf("generating datakit.conf...")
		if err := ioutil.WriteFile(mcp, cfgdata, os.ModePerm); err != nil {
			l.Error(err)
			return err
		}
	}

	return nil
}

const (
	tagsKVPartsLen = 2
)

func ParseGlobalTags(s string) map[string]string {
	tags := map[string]string{}

	parts := strings.Split(s, ",")
	for _, p := range parts {
		arr := strings.Split(p, "=")
		if len(arr) != tagsKVPartsLen {
			l.Warnf("invalid global tag: %s, ignored", p)
			continue
		}

		tags[arr[0]] = arr[1]
	}

	return tags
}
