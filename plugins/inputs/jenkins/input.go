package jenkins

import (
	"net/http"
	"time"

	"encoding/json"
	"fmt"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName   = `jenkins`
	l           = logger.DefaultSLogger(inputName)
	minInterval = time.Second
	maxInterval = time.Second * 30
	sample      = `
[[inputs.jenkins]]
	## The Jenkins URL in the format "schema://host:port",required
	url = "http://my-jenkins-instance:8080"
	## Metric Access Key ,generate in your-jenkins-host:/configure,required
	key = ""
	## Set response_timeout
	# response_timeout = "5s"
	
	## Optional TLS Config
	# tls_ca = "/xx/ca.pem"
	# tls_cert = "/xx/cert.pem"
	# tls_key = "/xx/key.pem"
	## Use SSL but skip chain & host verification
	# insecure_skip_verify = false
	[inputs.jenkins.log]
	#	files = []
	#	# grok pipeline script path
	#	pipeline = "jenkins.p"
	[inputs.jenkins.tags]
	# a = "b"`

	pipelineCfg = `
grok(_, "%{TIMESTAMP_ISO8601:time} \\[id=%{GREEDYDATA:id}\\]\t%{GREEDYDATA:status}\t")
default_time(time)
group_in(status, ["WARNING", "NOTICE"], "warning")
group_in(status, ["SERVER", "ERROR"], "error")
group_in(status, ["INFO"], "info")

`
)

func (_ *Input) SampleConfig() string {
	return sample
}

func (_ *Input) Catalog() string {
	return inputName
}

func (_ *Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		"jenkins": pipelineCfg,
	}
	return pipelineMap
}

func (n *Input) Run() {
	l = logger.SLogger(inputName)
	l.Info("jenkins start")
	n.Interval.Duration = datakit.ProtectedInterval(minInterval, maxInterval, n.Interval.Duration)

	if n.Log != nil {
		go func() {
			inputs.JoinPipelinePath(n.Log, "jenkins.p")
			n.Log.Source = inputName
			n.Log.Match = `^\d{4}-\d{2}-\d{2}`
			n.Log.FromBeginning = true
			n.Log.Tags = map[string]string{}
			for k, v := range n.Tags {
				n.Log.Tags[k] = v
			}
			tail, err := inputs.NewTailer(n.Log)
			if err != nil {
				l.Errorf("init tailf err:%s", err.Error())
				return
			}
			n.tail = tail
			tail.Run()
		}()
	}

	client, err := n.createHttpClient()
	if err != nil {
		l.Errorf("[error] jenkins init client err:%s", err.Error())
		return
	}
	n.client = client

	tick := time.NewTicker(n.Interval.Duration)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			n.getMetric()
			if len(n.collectCache) > 0 {
				err := inputs.FeedMeasurement(inputName, datakit.Metric, n.collectCache, &io.Option{CollectCost: time.Since(n.start)})
				n.collectCache = n.collectCache[:0]
				if err != nil {
					n.lastErr = err
					l.Errorf(err.Error())
					continue
				}
			}
			if n.lastErr != nil {
				io.FeedLastError(inputName, n.lastErr.Error())
				n.lastErr = nil
			}
		case <-datakit.Exit.Wait():
			if n.tail != nil {
				n.tail.Close()
				l.Info("jenkins log exit")
			}
			l.Info("jenkins exit")
			return
		}
	}
}

type MetricFunc func(input *Input)

func (n *Input) getMetric() {
	n.start = time.Now()
	// 此处函数待添加调研
	getFunc := []MetricFunc{getPluginMetric}
	n.wg.Add(len(getFunc))
	for _, v := range getFunc {
		go func(gf MetricFunc) {
			defer n.wg.Done()
			gf(n)
		}(v)
	}
	n.wg.Wait()

}

func (n *Input) requestJSON(u string, target interface{}) error {
	u = fmt.Sprintf("%s%s", n.Url, u)

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}

	//req.SetBasicAuth(n.Username, n.Password)

	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(target)

	return nil
}

func (n *Input) createHttpClient() (*http.Client, error) {
	tlsCfg, err := n.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	if n.ResponseTimeout.Duration < time.Second {
		n.ResponseTimeout.Duration = time.Second * 5
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: n.ResponseTimeout.Duration,
	}

	return client, nil
}

func (_ *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (n *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&Measurement{},
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		s := &Input{
			Interval: datakit.Duration{Duration: time.Second * 30},
		}
		return s
	})
}
