package solr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	regex_url_port = `http[s]?://((?:[\d]{1,3}(?:\.[\d]{1,3}){3})|(?:\[[\:\da-zA-Z]*\])|(?:[\.a-zA-Z\d-]+))(?::([\d]{0,5}))?`
)

func createHTTPClient(timeout datakit.Duration) *http.Client {
	tr := &http.Transport{
		ResponseHeaderTimeout: timeout.Duration,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   timeout.Duration,
	}
	return client
}

// log error
func logError(err error) {
	if err != nil {
		l.Error(err)
		io.FeedLastError(inputName, err.Error())
	}
}

func urljoin(server, path string, param [][2]string) string {
	p := ""
	if len(server) < 1 {
		return ""
	}
	if len(path) > 0 {
		if server[len(server)-1:] != "/" {
			server += "/"
		}
		if path[:1] == "/" {
			path = path[1:]
		}
	}
	if len(param) > 0 {
		for _, x := range param {
			p += "&" + x[0] + "=" + x[1]
		}
		p = p[1:]
	}
	return server + path + "?" + p
}

// ------------------------- assert --------------------

const (
	FieldCompare = 1 << iota
	NameCompare
	TagCompare
	TimeCompare
)

// cache
func whichMesaurement(k string) string {
	regex_cache, _ := regexp.Compile(prefix_regex_cache)
	regex_requesttimes, _ := regexp.Compile(prefix_regex_requesttimes)
	regex_searcher, _ := regexp.Compile(prefix_searcher)
	if regex_cache.MatchString(k) {
		return "cache"
	}
	if regex_requesttimes.MatchString(k) {
		return "requesttimes"
	}
	if regex_searcher.MatchString(k) {
		return "searcher"
	}
	return ""
}

// 根据 server url 生成 instance name， 使用正则匹配域名/ip和端口。
// 如 localhost_8983, 127.0.0.1_8983
func instanceName(serv string) (string, error) {
	var err error
	instanceName := ""
	if r, err := regexp.Compile(regex_url_port); err != nil {
	} else {
		l := r.FindAllStringSubmatch(serv, -1)
		if len(l) >= 1 && len(l[0]) == 3 && len(l[0][1]) > 0 {
			instanceName = l[0][1]
			if len(l[0][2]) > 0 {
				instanceName += "_" + l[0][2]
			}
		}
	}
	return instanceName, err
}

type GatherData func(i *Input, url string, v interface{}) error

// gather data
func gatherDataFunc(i *Input, url string, v interface{}) error {
	req, reqErr := http.NewRequest(http.MethodGet, url, nil)
	if reqErr != nil {
		return reqErr
	}

	if i.Username != "" {
		req.SetBasicAuth(i.Username, i.Password)
	}

	// req.Header.Set("User-Agent", "")
	r, err := i.client.Do(req)
	if err != nil {
		return err
	}

	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("solr: API responded with status-code %d, expected %d, url %s",
			r.StatusCode, http.StatusOK, url)
	}
	if err = json.NewDecoder(r.Body).Decode(v); err != nil {
		return nil
	}
	return nil
}
