// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package apache collects Apache metrics.
package apache

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf/plugins/common/tls"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	iod "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var _ inputs.ElectionInput = (*Input)(nil)

var (
	l = logger.DefaultSLogger(inputName)
	g = datakit.G("inputs_apache")
)

type Input struct {
	URLsDeprecated []string `toml:"urls,omitempty"`

	URL      string            `toml:"url"`
	Username string            `toml:"username,omitempty"`
	Password string            `toml:"password,omitempty"`
	Interval datakit.Duration  `toml:"interval,omitempty"`
	Tags     map[string]string `toml:"tags,omitempty"`
	Log      *struct {
		Files             []string `toml:"files"`
		Pipeline          string   `toml:"pipeline"`
		IgnoreStatus      []string `toml:"ignore"`
		CharacterEncoding string   `toml:"character_encoding"`
	} `toml:"log"`

	tls.ClientConfig

	start  time.Time
	tail   *tailer.Tailer
	client *http.Client

	Election bool `toml:"election"`
	pause    bool
	pauseCh  chan bool

	semStop *cliutils.Sem // start stop signal
}

func (n *Input) ElectionEnabled() bool {
	return n.Election
}

//nolint:lll
func (n *Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		inputName: {
			"Apache error log":  `[Tue May 19 18:39:45.272121 2021] [access_compat:error] [pid 9802] [client ::1:50547] AH01797: client denied by server configuration: /Library/WebServer/Documents/server-status`,
			"Apache access log": `127.0.0.1 - - [17/May/2021:14:51:09 +0800] "GET /server-status?auto HTTP/1.1" 200 917`,
		},
	}
}

var maxPauseCh = inputs.ElectionPauseChannelLength

func newInput() *Input {
	return &Input{
		Interval: datakit.Duration{Duration: time.Second * 30},
		pauseCh:  make(chan bool, maxPauseCh),
		Election: true,

		semStop: cliutils.NewSem(),
	}
}

func (*Input) SampleConfig() string { return sample }

func (*Input) Catalog() string { return inputName }

func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (*Input) SampleMeasurement() []inputs.Measurement { return []inputs.Measurement{&Measurement{}} }

func (*Input) PipelineConfig() map[string]string { return map[string]string{"apache": pipeline} }

func (n *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:  inputName,
			Service: inputName,
			Pipeline: func() string {
				if n.Log != nil {
					return n.Log.Pipeline
				}
				return ""
			}(),
		},
	}
}

func (n *Input) RunPipeline() {
	if n.Log == nil || len(n.Log.Files) == 0 {
		return
	}

	if n.Log.Pipeline == "" {
		n.Log.Pipeline = "apache.p" // use default
	}

	opt := &tailer.Option{
		Source:            inputName,
		Service:           inputName,
		Pipeline:          n.Log.Pipeline,
		GlobalTags:        n.Tags,
		IgnoreStatus:      n.Log.IgnoreStatus,
		CharacterEncoding: n.Log.CharacterEncoding,
		MultilinePatterns: []string{`^\[\w+ \w+ \d+`},
	}

	var err error
	n.tail, err = tailer.NewTailer(n.Log.Files, opt)
	if err != nil {
		l.Error(err)
		iod.FeedLastError(inputName, err.Error())
		return
	}

	g.Go(func(ctx context.Context) error {
		n.tail.Start()
		return nil
	})
}

func (n *Input) Run() {
	l = logger.SLogger(inputName)
	l.Info("apache start")
	n.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, n.Interval.Duration)

	client, err := n.createHTTPClient()
	if err != nil {
		l.Errorf("[error] apache init client err:%s", err.Error())
		return
	}
	n.client = client

	tick := time.NewTicker(n.Interval.Duration)
	defer tick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			n.exit()
			l.Info("apache exit")
			return

		case <-n.semStop.Wait():
			n.exit()
			l.Info("apache return")
			return

		case <-tick.C:
			if n.pause {
				l.Debugf("not leader, skipped")
				continue
			}

			m, err := n.getMetric()
			if err != nil {
				iod.FeedLastError(inputName, err.Error())
			}

			if m != nil {
				if err := inputs.FeedMeasurement(inputName,
					datakit.Metric,
					[]inputs.Measurement{m},
					&iod.Option{CollectCost: time.Since(n.start)}); err != nil {
					l.Warnf("inputs.FeedMeasurement: %s, ignored", err)
				}
			}

		case n.pause = <-n.pauseCh:
			// nil
		}
	}
}

func (n *Input) exit() {
	if n.tail != nil {
		n.tail.Close()
		l.Info("apache log exit")
	}
}

func (n *Input) Terminate() {
	if n.semStop != nil {
		n.semStop.Close()
	}
}

func (n *Input) createHTTPClient() (*http.Client, error) {
	tlsCfg, err := n.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: time.Second * 5,
	}

	return client, nil
}

func (n *Input) getMetric() (*Measurement, error) {
	n.start = time.Now()
	req, err := http.NewRequest("GET", n.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("error on new request to %s : %w", n.URL, err)
	}

	if len(n.Username) != 0 && len(n.Password) != 0 {
		req.SetBasicAuth(n.Username, n.Password)
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error on request to %s : %w", n.URL, err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned HTTP status %s", n.URL, resp.Status)
	}
	return n.parse(resp.Body)
}

func (n *Input) parse(body io.Reader) (*Measurement, error) {
	sc := bufio.NewScanner(body)

	tags := map[string]string{
		"url": n.URL,
	}
	for k, v := range n.Tags {
		tags[k] = v
	}
	metric := &Measurement{
		name:     inputName,
		fields:   map[string]interface{}{},
		ts:       time.Now(),
		election: n.Election,
	}

	for sc.Scan() {
		line := sc.Text()
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			key, part := strings.ReplaceAll(parts[0], " ", ""), strings.TrimSpace(parts[1])
			if tagKey, ok := tagMap[key]; ok {
				tags[tagKey] = part
			}

			fieldKey, ok := filedMap[key]
			if !ok {
				continue
			}
			switch key {
			case "CPULoad":
				value, err := strconv.ParseFloat(part, 64)
				if err != nil {
					l.Error(err.Error())
					continue
				}
				metric.fields[fieldKey] = value
			case "Scoreboard":
				scoreboard := map[string]int{
					WaitingForConnection: 0,
					StartingUp:           0,
					ReadingRequest:       0,
					SendingReply:         0,
					KeepAlive:            0,
					DNSLookup:            0,
					ClosingConnection:    0,
					Logging:              0,
					GracefullyFinishing:  0,
					IdleCleanup:          0,
					OpenSlot:             0,
				}
				for _, c := range part {
					switch c {
					case '_':
						scoreboard[WaitingForConnection]++
					case 'S':
						scoreboard[StartingUp]++
					case 'R':
						scoreboard[ReadingRequest]++
					case 'W':
						scoreboard[SendingReply]++
					case 'K':
						scoreboard[KeepAlive]++
					case 'D':
						scoreboard[DNSLookup]++
					case 'C':
						scoreboard[ClosingConnection]++
					case 'L':
						scoreboard[Logging]++
					case 'G':
						scoreboard[GracefullyFinishing]++
					case 'I':
						scoreboard[IdleCleanup]++
					case '.':
						scoreboard[OpenSlot]++
					}
				}
				for k, v := range scoreboard {
					metric.fields[k] = v
				}
			default:
				value, err := strconv.ParseInt(part, 10, 64)
				if err != nil {
					l.Error(err.Error())
					continue
				}
				if fieldKey == "Total kBytes" {
					// kbyte to byte
					metric.fields[fieldKey] = value * 1024
					continue
				}
				metric.fields[fieldKey] = value
			}
		}
	}
	metric.tags = tags

	return metric, nil
}

func (n *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case n.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (n *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case n.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return newInput()
	})
}
