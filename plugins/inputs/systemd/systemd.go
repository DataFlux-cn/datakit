// +build linux

package systemd

import (
	"strings"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "systemd"

	defaultMeasurement = "systemd"

	sampleCfg = `
[inputs.systemd]
    # valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    # required
    interval = "10s"
    
    # [inputs.systemd.tags]
    # tags1 = "value1"
`
)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Systemd{}
	})
}

type Systemd struct {
	Interval string            `toml:"interval"`
	Tags     map[string]string `toml:"tags"`

	conn     *dbus.Conn
	duration time.Duration
}

func (*Systemd) SampleConfig() string {
	return sampleCfg
}

func (*Systemd) Catalog() string {
	return "host"
}

func (s *Systemd) Run() {
	l = logger.SLogger(inputName)

	if s.loadcfg() {
		return
	}
	ticker := time.NewTicker(s.duration)
	defer ticker.Stop()

	l.Infof("systemd input started.")

	for {
		select {
		case <-datakit.Exit.Wait():
			s.conn.Close()
			l.Info("exit")
			return

		case <-ticker.C:
			data, err := s.getMetrics()
			if err != nil {
				l.Error(err)
				continue
			}
			if err := io.NamedFeed(data, io.Metric, inputName); err != nil {
				l.Error(err)
				continue
			}
			l.Debugf("feed %d bytes to io ok", len(data))
		}
	}
}

func (s *Systemd) loadcfg() bool {
	var err error

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		s.duration, err = time.ParseDuration(s.Interval)
		if err != nil || s.duration <= 0 {
			l.Errorf("invalid interval, err %s", err.Error())
			time.Sleep(time.Second)
			continue
		}

		s.conn, err = dbus.New()
		if err != nil {
			l.Errorf("connect systemd err: %s", err.Error())
			time.Sleep(time.Second)
			continue
		}
		break
	}
	return false
}

type metrics struct {
	total     int
	loaded    int
	active    int
	service   int
	socket    int
	device    int
	mount     int
	automount int
	swap      int
	target    int
	path      int
	timer     int
	slice     int
	scope     int
}

func (s *Systemd) getMetrics() ([]byte, error) {

	units, err := s.conn.ListUnits()
	if err != nil {
		return nil, err
	}

	statusMetrics := unitStatus(units)

	fields := map[string]interface{}{
		"units_total":        statusMetrics.total,
		"units_loaded_count": statusMetrics.loaded,
		"units_active_count": statusMetrics.active,
		"unit_service":       statusMetrics.service,
		"unit_socket":        statusMetrics.socket,
		"unit_device":        statusMetrics.device,
		"unit_mount":         statusMetrics.mount,
		"unit_automount":     statusMetrics.automount,
		"unit_swap":          statusMetrics.swap,
		"unit_target":        statusMetrics.target,
		"unit_path":          statusMetrics.path,
		"unit_timer":         statusMetrics.timer,
		"unit_slice":         statusMetrics.slice,
		"unit_scope":         statusMetrics.scope,
	}
	return io.MakeMetric(defaultMeasurement, s.Tags, fields, time.Now())
}

func unitStatus(units []dbus.UnitStatus) metrics {
	var m metrics
	m.total = len(units)

	for index := range units {
		nameBlocks := strings.Split(units[index].Name, ".")
		switch nameBlocks[len(nameBlocks)-1] {
		case "service":
			m.service++
		case "socket":
			m.socket++
		case "device":
			m.device++
		case "mount":
			m.mount++
		case "automount":
			m.automount++
		case "swap":
			m.swap++
		case "target":
			m.target++
		case "path":
			m.path++
		case "timer":
			m.timer++
		case "slice":
			m.slice++
		case "scope":
			m.scope++
		}
		if units[index].LoadState == "loaded" {
			m.loaded++
		}
		if units[index].ActiveState == "active" {
			m.active++
		}
	}
	return m
}
