package lighttpd

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "lighttpd"

	defaultMeasurement = "lighttpd"

	sampleCfg = `
# [[inputs.lighttpd]]
# 	# lighttpd status url
# 	url = "http://127.0.0.1:8080/server-status"
# 	
# 	# lighttpd version is "v1" or "v2"
# 	version = "v1"
# 	
# 	# second
# 	collect_cycle = 60
# 	
# 	# [inputs.tailf.tags]
# 	# tags1 = "tags1"
`
)

var l *zap.SugaredLogger

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Lighttpd{}
	})
}

type Lighttpd struct {
	URL     string            `toml:"url"`
	Version string            `toml:"version"`
	Cycle   time.Duration     `toml:"collect_cycle"`
	Tags    map[string]string `toml:"tags"`

	statusURL     string
	statusVersion Version
}

func (_ *Lighttpd) SampleConfig() string {
	return sampleCfg
}

func (_ *Lighttpd) Catalog() string {
	return inputName
}

func (h *Lighttpd) Run() {
	l = logger.SLogger(inputName)

	if _, ok := h.Tags["url"]; !ok {
		h.Tags["url"] = h.URL
	}
	if _, ok := h.Tags["version"]; !ok {
		h.Tags["version"] = h.Version
	}

	switch h.Version {
	case "v1":
		h.statusURL = fmt.Sprintf("%s?json", h.URL)
		h.statusVersion = v1
	case "v2":
		h.statusURL = fmt.Sprintf("%s?format=plain", h.URL)
		h.statusVersion = v2
	default:
		l.Error("invalid lighttpd version")
		return
	}

	ticker := time.NewTicker(time.Second * h.Cycle)
	defer ticker.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

		case <-ticker.C:
			data, err := LighttpdStatusParse(h.statusURL, h.statusVersion, h.Tags)
			if err != nil {
				l.Error(err)
				continue
			}
			if err := io.Feed(data, io.Metric); err != nil {
				l.Error(err)
				continue
			}
			l.Debugf("feed %d bytes to io ok", len(data))
		}
	}
}
