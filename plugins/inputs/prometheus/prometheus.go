package prometheus

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tls"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	acceptHeader = `application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited;q=0.7,text/plain;version=0.0.4;q=0.3,*/*;q=0.1`
)

var (
	promAgent *Prometheus

	sampleConfig = `
  ## An array of urls to scrape metrics from.
  #urls = ["http://localhost:9100/metrics"]

  ## Url tag name (tag containing scrapped url. optional, default is "url")
  # url_tag = "scrapeUrl"

  #[metric_name_map]
  #node_cpu_seconds_total = 'node_cpu'

  ## An array of Kubernetes services to scrape metrics from.
  # kubernetes_services = ["http://my-service-dns.my-namespace:9100/metrics"]

  ## Kubernetes config file to create client from.
  # kube_config = "/path/to/kubernetes.config"

  ## Scrape Kubernetes pods for the following prometheus annotations:
  ## - prometheus.io/scrape: Enable scraping for this pod
  ## - prometheus.io/scheme: If the metrics endpoint is secured then you will need to
  ##     set this to 'https' & most likely set the tls config.
  ## - prometheus.io/path: If the metrics path is not /metrics, define it with this annotation.
  ## - prometheus.io/port: If port is not 9102 use this annotation
  # monitor_kubernetes_pods = true
  ## Restricts Kubernetes monitoring to a single namespace
  ##   ex: monitor_kubernetes_pods_namespace = "default"
  # monitor_kubernetes_pods_namespace = ""

  ## Use bearer token for authorization. ('bearer_token' takes priority)
  # bearer_token = "/path/to/bearer/token"
  ## OR
  # bearer_token_string = "abc_123"

  ## HTTP Basic Authentication username and password. ('bearer_token' and
  ## 'bearer_token_string' take priority)
  # username = ""
  # password = ""

  ## Specify timeout duration for slower prometheus clients (default is 3s)
  # response_timeout = "3s"

  ## Optional TLS Config
  # tls_ca = /path/to/cafile
  # tls_cert = /path/to/certfile
  # tls_key = /path/to/keyfile
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`
)

type Prometheus struct {
	// An array of urls to scrape metrics from.
	URLs []string `toml:"urls"`

	Interval internal.Duration

	MetricNameMap map[string]string `toml:"metric_name_map"`

	// An array of Kubernetes services to scrape metrics from.
	KubernetesServices []string

	// Location of kubernetes config file
	KubeConfig string

	// Bearer Token authorization file path
	BearerToken       string `toml:"bearer_token"`
	BearerTokenString string `toml:"bearer_token_string"`

	// Basic authentication credentials
	Username string `toml:"username"`
	Password string `toml:"password"`

	ResponseTimeout internal.Duration `toml:"response_timeout"`

	MetricVersion int `toml:"metric_version"`

	URLTag string `toml:"url_tag"`

	tls.ClientConfig

	client *http.Client

	// Should we scrape Kubernetes services for prometheus annotations
	MonitorPods    bool   `toml:"monitor_kubernetes_pods"`
	PodNamespace   string `toml:"monitor_kubernetes_pods_namespace"`
	lock           sync.Mutex
	kubernetesPods map[string]URLAndAddress
	cancel         context.CancelFunc
	wg             sync.WaitGroup

	stopCtx       context.Context
	stopCancelFun context.CancelFunc
}

func (p *Prometheus) Catalog() string {
	return `prometheus`
}

func (p *Prometheus) SampleConfig() string {
	return sampleConfig
}

// func (p *Prometheus) Description() string {
// 	return "Read metrics from one or many prometheus clients"
// }

var ErrProtocolError = errors.New("prometheus protocol error")

func (p *Prometheus) AddressToURL(u *url.URL, address string) *url.URL {
	host := address
	if u.Port() != "" {
		host = address + ":" + u.Port()
	}
	reconstructedURL := &url.URL{
		Scheme:     u.Scheme,
		Opaque:     u.Opaque,
		User:       u.User,
		Path:       u.Path,
		RawPath:    u.RawPath,
		ForceQuery: u.ForceQuery,
		RawQuery:   u.RawQuery,
		Fragment:   u.Fragment,
		Host:       host,
	}
	return reconstructedURL
}

type URLAndAddress struct {
	OriginalURL *url.URL
	URL         *url.URL
	Address     string
	Tags        map[string]string
}

func (p *Prometheus) GetAllURLs() (map[string]URLAndAddress, error) {
	allURLs := make(map[string]URLAndAddress, 0)
	for _, u := range p.URLs {
		URL, err := url.Parse(u)
		if err != nil {
			log.Printf("E! Could not parse %q, skipping it. Error: %s", u, err.Error())
			continue
		}
		allURLs[URL.String()] = URLAndAddress{URL: URL, OriginalURL: URL}
	}

	p.lock.Lock()
	defer p.lock.Unlock()
	// loop through all pods scraped via the prometheus annotation on the pods
	for k, v := range p.kubernetesPods {
		allURLs[k] = v
	}

	for _, service := range p.KubernetesServices {
		URL, err := url.Parse(service)
		if err != nil {
			return nil, err
		}

		resolvedAddresses, err := net.LookupHost(URL.Hostname())
		if err != nil {
			log.Printf("E! Could not resolve %q, skipping it. Error: %s", URL.Host, err.Error())
			continue
		}
		for _, resolved := range resolvedAddresses {
			serviceURL := p.AddressToURL(URL, resolved)
			allURLs[serviceURL.String()] = URLAndAddress{
				URL:         serviceURL,
				Address:     resolved,
				OriginalURL: URL,
			}
		}
	}
	return allURLs, nil
}

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (p *Prometheus) gather() error {

	log.Printf("[prometheus] gather start")

	if p.client == nil {
		client, err := p.createHTTPClient()
		if err != nil {
			log.Printf("E! [prometheus] fail to createHTTPClient, %s", err)
			return err
		}
		p.client = client
	}

	allURLs, err := p.GetAllURLs()
	if err != nil {
		log.Printf("E! [prometheus] fail to GetAllURLs, %s", err)
		return err
	}

	for _, URL := range allURLs {

		select {
		case <-p.stopCtx.Done():
			return nil
		default:
			break
		}

		p.gatherURL(URL)

	}

	return nil
}

func (p *Prometheus) createHTTPClient() (*http.Client, error) {
	tlsCfg, err := p.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   tlsCfg,
			DisableKeepAlives: true,
		},
		Timeout: p.ResponseTimeout.Duration,
	}

	return client, nil
}

func (p *Prometheus) gatherURL(u URLAndAddress) error {
	var req *http.Request
	var err error
	var uClient *http.Client
	var metrics []*influxdb.Point
	if u.URL.Scheme == "unix" {
		path := u.URL.Query().Get("path")
		if path == "" {
			path = "/metrics"
		}
		req, err = http.NewRequest("GET", "http://localhost"+path, nil)

		// ignore error because it's been handled before getting here
		tlsCfg, _ := p.ClientConfig.TLSConfig()
		uClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig:   tlsCfg,
				DisableKeepAlives: true,
				Dial: func(network, addr string) (net.Conn, error) {
					c, err := net.Dial("unix", u.URL.Path)
					return c, err
				},
			},
			Timeout: p.ResponseTimeout.Duration,
		}
	} else {
		if u.URL.Path == "" {
			u.URL.Path = "/metrics"
		}
		req, err = http.NewRequest("GET", u.URL.String(), nil)
	}

	req.Header.Add("Accept", acceptHeader)

	if p.BearerToken != "" {
		token, err := ioutil.ReadFile(p.BearerToken)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+string(token))
	} else if p.BearerTokenString != "" {
		req.Header.Set("Authorization", "Bearer "+p.BearerTokenString)
	} else if p.Username != "" || p.Password != "" {
		req.SetBasicAuth(p.Username, p.Password)
	}

	var resp *http.Response
	if u.URL.Scheme != "unix" {
		resp, err = p.client.Do(req)
	} else {
		resp, err = uClient.Do(req)
	}
	if err != nil {
		log.Printf("E! [prometheus] get failed, %s", err)
		return fmt.Errorf("error making HTTP request to %s: %s", u.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", u.URL, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading body: %s", err)
	}

	if p.MetricVersion == 2 {
		metrics, err = ParseV2(body, resp.Header)
	} else {
		metrics, err = Parse(body, resp.Header)
	}

	if err != nil {
		return fmt.Errorf("error reading metrics for %s: %s",
			u.URL, err)
	}

	for _, metric := range metrics {
		tags := metric.Tags()
		// strip user and password from URL
		u.OriginalURL.User = nil
		if p.URLTag != "" {
			tags[p.URLTag] = u.OriginalURL.String()
		}
		if u.Address != "" {
			tags["address"] = u.Address
		}
		for k, v := range u.Tags {
			tags[k] = v
		}

		fields, _ := metric.Fields()

		pt, err := influxdb.NewPoint(metric.Name(), tags, fields, metric.Time())
		if err == nil {
			io.Feed([]byte(pt.String()), io.Metric)
		}

		// switch metric.Type() {
		// case telegraf.Counter:
		// 	acc.AddCounter(metric.Name(), metric.Fields(), tags, metric.Time())
		// case telegraf.Gauge:
		// 	acc.AddGauge(metric.Name(), metric.Fields(), tags, metric.Time())
		// case telegraf.Summary:
		// 	acc.AddSummary(metric.Name(), metric.Fields(), tags, metric.Time())
		// case telegraf.Histogram:
		// 	acc.AddHistogram(metric.Name(), metric.Fields(), tags, metric.Time())
		// default:
		// 	acc.AddFields(metric.Name(), metric.Fields(), tags, metric.Time())
		// }
	}

	return nil
}

//Start will start the Kubernetes scraping if enabled in the configuration
func (p *Prometheus) Run() {

	if p.Interval.Duration == 0 {
		p.Interval.Duration = time.Second * 10
	}
	// if p.MetricVersion != 2 {
	// 	p.Log.Warnf("Use of deprecated configuration: 'metric_version = 1'; please update to 'metric_version = 2'")
	// }

	go func() {
		<-config.Exit.Wait()
		p.stopCancelFun()
		if p.MonitorPods {
			p.cancel()
		}
	}()

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for {

			select {
			case <-p.stopCtx.Done():
				return
			default:
			}

			p.gather()

			internal.SleepContext(p.stopCtx, p.Interval.Duration)
		}
	}()

	if p.MonitorPods {
		var ctx context.Context
		ctx, p.cancel = context.WithCancel(context.Background())
		p.start(ctx)
	}

	p.wg.Wait()
}

func init() {
	inputs.Add("prometheus", func() inputs.Input {
		promAgent = &Prometheus{
			ResponseTimeout: internal.Duration{Duration: time.Second * 3},
			kubernetesPods:  map[string]URLAndAddress{},
			URLTag:          "url",
			MetricVersion:   1,
		}
		promAgent.stopCtx, promAgent.stopCancelFun = context.WithCancel(context.Background())
		return promAgent
	})
}
