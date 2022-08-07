// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package config manage datakit's configurations, include all inputs TOML configure.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	bstoml "github.com/BurntSushi/toml"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/tracer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cgroup"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

var (
	Cfg = DefaultConfig()
	l   = logger.DefaultSLogger("config")
)

func SetLog() {
	l = logger.SLogger("config")
}

func DefaultConfig() *Config {
	c := &Config{ //nolint:dupl
		GlobalHostTags:       map[string]string{},
		GlobalTagsDeprecated: map[string]string{},
		GlobalEnvTags:        map[string]string{
			// "project": "",
			// "cluster": "",
			// "site":    "",
		},

		Environments: map[string]string{
			"ENV_HOSTNAME": "", // not set
		}, // default nothing

		IOConf: &dkio.IOConfig{
			FeedChanSize:         128,
			MaxCacheCount:        64,
			MaxDynamicCacheCount: 64,
			FlushInterval:        "10s",
			OutputFileInputs:     []string{},

			EnableCache: false,
			CacheSizeGB: 1,

			BlockingMode: false,

			Filters: map[string][]string{},
		},

		DataWayCfg: &dataway.DataWayCfg{
			URLs: []string{},
		},

		ProtectMode: true,

		HTTPAPI: &dkhttp.APIConfig{
			RUMOriginIPHeader:   "X-Forwarded-For",
			Listen:              "localhost:9529",
			RUMAppIDWhiteList:   []string{},
			PublicAPIs:          []string{},
			Timeout:             "30s",
			CloseIdleConnection: false,
		},

		DCAConfig: &dkhttp.DCAConfig{
			Enable:    false,
			Listen:    "0.0.0.0:9531",
			WhiteList: []string{},
		},
		Pipeline: &pipeline.PipelineCfg{
			IPdbType:           "-",
			RemotePullInterval: "1m",
		},
		Logging: &LoggerCfg{
			Level:  "info",
			Rotate: 32,
			Log:    filepath.Join("/var/log/datakit", "log"),
			GinLog: filepath.Join("/var/log/datakit", "gin.log"),
		},

		Cgroup: &cgroup.CgroupOptions{
			Path:   "/datakit",
			Enable: true,
			CPUMax: 20.0,
			CPUMin: 5.0,
			MemMax: 4096, // MB
		},

		GitRepos: &GitRepost{
			PullInterval: "1m",
			Repos: []*GitRepository{
				{
					Enable:                false,
					URL:                   "",
					SSHPrivateKeyPath:     "",
					SSHPrivateKeyPassword: "",
					Branch:                "master",
				},
			},
		},

		Sinks: &Sinker{
			Sink: []map[string]interface{}{{}},
		},

		EnableElection:    false,
		EnableElectionTag: false,
		ElectionNamespace: "default",

		Ulimit: 64000,
	}

	// windows 下，日志继续跟 datakit 放在一起
	if runtime.GOOS == datakit.OSWindows {
		c.Logging.Log = filepath.Join(datakit.InstallDir, "log")
		c.Logging.GinLog = filepath.Join(datakit.InstallDir, "gin.log")
	}

	return c
}

type LoggerCfg struct {
	Log          string `toml:"log"`
	GinLog       string `toml:"gin_log"`
	Level        string `toml:"level"`
	DisableColor bool   `toml:"disable_color"`
	Rotate       int    `toml:"rotate,omitzero"`
}

type GitRepository struct {
	Enable                bool   `toml:"enable"`
	URL                   string `toml:"url"`
	SSHPrivateKeyPath     string `toml:"ssh_private_key_path"`
	SSHPrivateKeyPassword string `toml:"ssh_private_key_password"`
	Branch                string `toml:"branch"`
}

type GitRepost struct {
	PullInterval string           `toml:"pull_interval"`
	Repos        []*GitRepository `toml:"repo"`
}

type Sinker struct {
	Sink []map[string]interface{} `toml:"sink"`
}

type Config struct {
	DefaultEnabledInputs []string  `toml:"default_enabled_inputs,omitempty"`
	InstallDate          time.Time `toml:"install_date,omitempty"`
	UpgradeDate          time.Time `toml:"upgrade_date,omitempty"`

	BlackList []*inputHostList `toml:"black_lists,omitempty"`
	WhiteList []*inputHostList `toml:"white_lists,omitempty"`

	UUID    string `toml:"-"`
	RunMode int    `toml:"-"`

	Name     string `toml:"name,omitempty"`
	Hostname string `toml:"-"`

	// http config: TODO: merge into APIConfig
	HTTPBindDeprecated   string `toml:"http_server_addr,omitempty"`
	HTTPListenDeprecated string `toml:"http_listen,omitempty"`

	IntervalDeprecated   string `toml:"interval,omitempty"`
	OutputFileDeprecated string `toml:"output_file,omitempty"`
	UUIDDeprecated       string `toml:"uuid,omitempty"` // deprecated

	// pprof
	EnablePProf bool   `toml:"enable_pprof"`
	PProfListen string `toml:"pprof_listen"`

	// DCA config
	DCAConfig *dkhttp.DCAConfig `toml:"dca"`

	// pipeline
	Pipeline *pipeline.PipelineCfg `toml:"pipeline"`

	// logging config
	LogDeprecated      string `toml:"log,omitempty"`
	LogLevelDeprecated string `toml:"log_level,omitempty"`
	GinLogDeprecated   string `toml:"gin_log,omitempty"`

	InstallVer string `toml:"install_version,omitempty"`

	HTTPAPI *dkhttp.APIConfig `toml:"http_api"`
	IOConf  *dkio.IOConfig    `toml:"io"`

	DataWayCfg *dataway.DataWayCfg `toml:"dataway,omitempty"`
	DataWay    dataway.DataWay     `toml:"-"`

	Sinks   *Sinker    `toml:"sinks"`
	Logging *LoggerCfg `toml:"logging"`

	LogRotateDeprecated    int `toml:"log_rotate,omitzero"`
	IOCacheCountDeprecated int `toml:"io_cache_count,omitzero"`

	GlobalHostTags       map[string]string `toml:"global_host_tags"`
	GlobalEnvTags        map[string]string `toml:"global_env_tags"`
	GlobalTagsDeprecated map[string]string `toml:"global_tags,omitempty"`

	Environments map[string]string     `toml:"environments"`
	Cgroup       *cgroup.CgroupOptions `toml:"cgroup"`

	Disable404PageDeprecated bool `toml:"disable_404page,omitempty"`
	ProtectMode              bool `toml:"protect_mode"`

	// TODO: we should group election-related conf into specified field
	EnableElection      bool   `toml:"enable_election"`
	EnableElectionTag   bool   `toml:"enable_election_tag"`
	ElectionNamespace   string `toml:"election_namespace"`
	NamespaceDeprecated string `toml:"namespace,omitempty"` // 避免跟 k8s 的 namespace 概念混淆

	// 是否已开启自动更新，通过 dk-install --ota 来开启
	AutoUpdate bool `toml:"auto_update,omitempty"`

	EnableUncheckedInputs bool `toml:"enable_unchecked_inputs,omitempty"`

	Tracer *tracer.Tracer `toml:"tracer,omitempty"`

	GitRepos *GitRepost `toml:"git_repos"`

	Ulimit uint64 `toml:"ulimit"`
}

func (c *Config) String() string {
	buf := new(bytes.Buffer)
	if err := bstoml.NewEncoder(buf).Encode(c); err != nil {
		return ""
	}

	return buf.String()
}

func (c *Config) SetUUID() error {
	if c.Hostname == "" {
		hn, err := os.Hostname()
		if err != nil {
			l.Errorf("get hostname failed: %s", err.Error())
			return err
		}

		c.UUID = hn
	} else {
		c.UUID = c.Hostname
	}
	return nil
}

func (c *Config) LoadMainTOML(p string) error {
	cfgdata, err := ioutil.ReadFile(filepath.Clean(p))
	if err != nil {
		return fmt.Errorf("ioutil.ReadFile: %w", err)
	}

	_, err = bstoml.Decode(string(cfgdata), c)
	if err != nil {
		return fmt.Errorf("bstoml.Decode: %w", err)
	}

	_ = c.SetUUID()

	return nil
}

type inputHostList struct {
	Hosts  []string `toml:"hosts"`
	Inputs []string `toml:"inputs"`
}

func (i *inputHostList) MatchHost(host string) bool {
	for _, hostname := range i.Hosts {
		if hostname == host {
			return true
		}
	}

	return false
}

func (i *inputHostList) MatchInput(input string) bool {
	for _, name := range i.Inputs {
		if name == input {
			return true
		}
	}

	return false
}

func (c *Config) InitCfg(p string) error {
	if c.Hostname == "" {
		if err := c.setHostname(); err != nil {
			return err
		}
	}

	if mcdata, err := datakit.TomlMarshal(c); err != nil {
		l.Errorf("TomlMarshal(): %s", err.Error())
		return err
	} else if err := ioutil.WriteFile(p, mcdata, datakit.ConfPerm); err != nil {
		l.Errorf("error creating %s: %s", p, err)
		return err
	}

	return nil
}

func initCfgSample(p string) error {
	if err := ioutil.WriteFile(p, []byte(DatakitConfSample), datakit.ConfPerm); err != nil {
		l.Errorf("error creating %s: %s", p, err)
		return err
	}
	l.Debugf("create datakit sample conf ok, %s!", p)
	return nil
}

func (c *Config) parseGlobalHostTags() {
	// why?
	if c.GlobalHostTags == nil {
		c.GlobalHostTags = map[string]string{}
	}

	// setup global tags
	for k, v := range c.GlobalHostTags {
		// NOTE: accept `__` and `$` as tag-key prefix, to keep compatible with old prefix `$`
		// by using `__` as prefix, avoid escaping `$` in Powershell and shell

		switch strings.ToLower(v) {
		case `__datakit_hostname`, `$datakit_hostname`:
			if c.Hostname == "" {
				if err := c.setHostname(); err != nil {
					l.Warnf("setHostname: %s, ignored", err)
				}
			}

			c.GlobalHostTags[k] = c.Hostname
			l.Debugf("set global tag %s: %s", k, c.Hostname)

		case `__datakit_ip`, `$datakit_ip`:
			c.GlobalHostTags[k] = "unavailable"

			if ipaddr, err := datakit.LocalIP(); err != nil {
				l.Errorf("get local ip failed: %s", err.Error())
			} else {
				l.Infof("set global tag %s: %s", k, ipaddr)
				c.GlobalHostTags[k] = ipaddr
			}

		case `__datakit_uuid`, `__datakit_id`, `$datakit_uuid`, `$datakit_id`:
			c.GlobalHostTags[k] = c.UUID
			l.Debugf("set global tag %s: %s", k, c.UUID)

		default:
			// pass
		}
	}
}

func (c *Config) setLogging() {
	// set global log root
	switch c.Logging.Log {
	case "stdout", "":
		l.Info("set log to stdout, rotate disabled")

		optflags := (logger.OPT_DEFAULT | logger.OPT_STDOUT)
		if !c.Logging.DisableColor {
			optflags |= logger.OPT_COLOR
		}

		if err := logger.InitRoot(
			&logger.Option{
				Level: c.Logging.Level,
				Flags: optflags,
			}); err != nil {
			l.Errorf("set root log faile: %s", err.Error())
		}
	default:

		if c.Logging.Rotate > 0 {
			logger.MaxSize = c.Logging.Rotate
		}

		if err := logger.InitRoot(&logger.Option{
			Path:  c.Logging.Log,
			Level: c.Logging.Level,
			Flags: logger.OPT_DEFAULT,
		}); err != nil {
			l.Panicf("set root log to %s faile: %s", c.Logging.Log, err.Error())
		}

		l.Infof("set root logger to %s ok", c.Logging.Log)
	}
}

// setup global host/env tags.
func (c *Config) setupGlobalTags() {
	c.parseGlobalHostTags()

	if len(c.GlobalTagsDeprecated) != 0 { // c.GlobalTags deprecated, move them to GlobalHostTags
		for k, v := range c.GlobalTagsDeprecated {
			c.GlobalHostTags[k] = v
		}
	}

	for k, v := range c.GlobalHostTags {
		point.SetGlobalHostTags(k, v)
	}

	// 开启选举且开启开关的情况下，将选举的命名空间塞到 global-env-tags 中
	if c.EnableElection && c.EnableElectionTag {
		c.GlobalEnvTags["election_namespace"] = c.ElectionNamespace
	}
	for k, v := range c.GlobalEnvTags {
		point.SetGlobalEnvTags(k, v)
	}

	// 此处不将 host 计入 c.GlobalHostTags，因为 c.GlobalHostTags 是读取的用户配置，而 host
	// 是不允许修改的, 故单独添加这个 tag 到 io 模块
	point.SetGlobalHostTags("host", c.Hostname)
}

func (c *Config) ApplyMainConfig() error {
	c.setLogging()

	l = logger.SLogger("config")

	// Set up ulimit.
	if err := setUlimit(c.Ulimit); err != nil {
		return fmt.Errorf("fail to set ulimit to %d: %w", c.Ulimit, err)
	} else {
		soft, hard, err := getUlimit()
		if err != nil {
			l.Warnf("fail to get ulimit: %v", err)
		} else {
			l.Infof("ulimit set to softLimit = %d, hardLimit = %d", soft, hard)
		}
	}

	if c.NamespaceDeprecated != "" && c.ElectionNamespace == "" {
		c.ElectionNamespace = c.NamespaceDeprecated
	}

	if c.EnableUncheckedInputs {
		datakit.EnableUncheckInputs = true
	}

	if c.Hostname == "" {
		if err := c.setHostname(); err != nil {
			return err
		}
	}

	if c.DataWayCfg != nil && len(c.DataWayCfg.URLs) > 0 {
		if err := c.setupDataway(); err != nil {
			return err
		}
	}

	datakit.AutoUpdate = c.AutoUpdate
	point.EnableElection = c.EnableElection

	// config default io
	if c.IOConf != nil {
		if c.IOConf.MaxCacheCount == 0 && c.IOCacheCountDeprecated != 0 {
			c.IOConf.MaxCacheCount = c.IOCacheCountDeprecated
		}
		if c.IOConf.OutputFile == "" && c.OutputFileDeprecated != "" {
			c.IOConf.OutputFile = c.OutputFileDeprecated
		}

		dkio.ConfigDefaultIO(c.IOConf)
		dkio.SetDataway(c.DataWay)
	}

	if err := c.setupSinks(); err != nil {
		l.Errorf("setup sinks failed: %s", err)
		return err
	}

	c.setupGlobalTags()

	// remove deprecated UUID field in main configure
	if c.UUIDDeprecated != "" {
		c.UUIDDeprecated = "" // clear deprecated UUID field
		buf := new(bytes.Buffer)
		if err := bstoml.NewEncoder(buf).Encode(c); err != nil {
			l.Fatalf("encode main configure failed: %s", err.Error())
		}
		if err := ioutil.WriteFile(datakit.MainConfPath, buf.Bytes(), datakit.ConfPerm); err != nil {
			l.Fatalf("refresh main configure failed: %s", err.Error())
		}

		l.Info("refresh main configure ok")
	}

	InitGitreposDir()

	return nil
}

func (c *Config) setHostname() error {
	// try get hostname from configure
	if v, ok := c.Environments["ENV_HOSTNAME"]; ok && v != "" {
		c.Hostname = v
		l.Infof("set hostname to %s from config ENV_HOSTNAME", v)
		datakit.DatakitHostName = c.Hostname
		return nil
	}

	// get real hostname
	hn, err := os.Hostname()
	if err != nil {
		l.Errorf("get hostname failed: %s", err.Error())
		return err
	}

	c.Hostname = hn

	l.Infof("hostname: %s", c.Hostname)
	datakit.DatakitHostName = c.Hostname
	return nil
}

func (c *Config) EnableDefaultsInputs(inputlist string) {
	inputs := []string{}
	inputsUnique := make(map[string]bool)

	for _, name := range c.DefaultEnabledInputs {
		if _, ok := inputsUnique[name]; !ok {
			inputsUnique[name] = true
			inputs = append(inputs, name)
		}
	}

	elems := strings.Split(inputlist, ",")
	for _, name := range elems {
		if _, ok := inputsUnique[name]; !ok {
			inputsUnique[name] = true
			inputs = append(inputs, name)
		}
	}

	c.DefaultEnabledInputs = inputs
}

func ParseGlobalTags(s string) map[string]string {
	if s == "" {
		return map[string]string{}
	}

	tags := map[string]string{}

	parts := strings.Split(s, ",")
	for _, p := range parts {
		arr := strings.Split(p, "=")
		if len(arr) != 2 {
			l.Warnf("invalid global tag: %s, ignored", p)
			continue
		}

		tags[arr[0]] = arr[1]
	}

	return tags
}

func CreateUUIDFile(f, uuid string) error {
	return ioutil.WriteFile(f, []byte(uuid), datakit.ConfPerm)
}

func LoadUUID(f string) (string, error) {
	if data, err := ioutil.ReadFile(filepath.Clean(f)); err != nil {
		return "", err
	} else {
		return string(data), nil
	}
}

func emptyDir(fp string) bool {
	fd, err := os.Open(filepath.Clean(fp))
	if err != nil {
		l.Error(err)
		return false
	}

	defer fd.Close() //nolint:errcheck,gosec

	_, err = fd.ReadDir(1)
	return errors.Is(err, io.EOF)
}

// remove all xxx.conf.sample.
func removeSamples() {
	l.Debugf("searching samples under %s", datakit.ConfdDir)

	fps := SearchDir(datakit.ConfdDir, ".conf.sample")

	l.Debugf("searched %d samples", len(fps))

	for _, fp := range fps {
		if err := os.Remove(fp); err != nil {
			l.Error(err)
			continue
		}

		l.Debugf("remove sample %s", fp)

		// check if directory empty
		pwd := filepath.Dir(fp)
		if emptyDir(pwd) {
			if err := os.RemoveAll(pwd); err != nil {
				l.Error(err)
			}
		}

		l.Debugf("remove dir %s", pwd)
	}
}

func MoveDeprecatedCfg() {
	if _, err := os.Stat(datakit.MainConfPathDeprecated); err == nil {
		if err := os.Rename(datakit.MainConfPathDeprecated, datakit.MainConfPath); err != nil {
			l.Fatal("move deprecated main configure failed: %s", err.Error())
		}
		l.Infof("move %s to %s", datakit.MainConfPathDeprecated, datakit.MainConfPath)
	}
}

func CreateSymlinks() error {
	var x [][2]string

	if runtime.GOOS == datakit.OSWindows {
		x = [][2]string{
			{
				filepath.Join(datakit.InstallDir, "datakit.exe"),
				`C:\WINDOWS\system32\datakit.exe`,
			},
		}
	} else {
		x = [][2]string{
			{
				filepath.Join(datakit.InstallDir, "datakit"),
				"/usr/local/bin/datakit",
			},

			{
				filepath.Join(datakit.InstallDir, "datakit"),
				"/usr/local/sbin/datakit",
			},

			{
				filepath.Join(datakit.InstallDir, "datakit"),
				"/sbin/datakit",
			},

			{
				filepath.Join(datakit.InstallDir, "datakit"),
				"/usr/sbin/datakit",
			},

			{
				filepath.Join(datakit.InstallDir, "datakit"),
				"/usr/bin/datakit",
			},
		}
	}

	ok := 0
	for _, item := range x {
		if err := os.MkdirAll(filepath.Dir(item[1]), os.ModePerm); err != nil {
			l.Warnf("create dir %s failed: %s, ignored", err.Error())
			continue
		}

		if err := symlink(item[0], item[1]); err != nil {
			l.Warnf("create datakit symlink: %s -> %s: %s, ignored", item[1], item[0], err.Error())
			continue
		}
		ok++
	}

	if ok == 0 {
		return fmt.Errorf("create symlink failed")
	}

	return nil
}

func symlink(src, dst string) error {
	l.Debugf("remove link %s...", dst)
	if err := os.Remove(dst); err != nil {
		l.Warnf("%s, ignored", err)
	}

	return os.Symlink(src, dst)
}

func GetToken() string {
	if Cfg.DataWay == nil {
		return ""
	}

	tokens := Cfg.DataWay.GetTokens()

	if len(tokens) > 0 {
		return tokens[0]
	}

	return ""
}

func GitHasEnabled() bool {
	return datakit.GitReposRepoName != "" && datakit.GitReposRepoFullPath != ""
}

func ProtectedInterval(min, max, cur time.Duration) time.Duration {
	if Cfg.ProtectMode {
		if cur >= max {
			return max
		}

		if cur <= min {
			return min
		}
	}

	return cur
}
