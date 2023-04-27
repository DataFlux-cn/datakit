// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package promremote handle promremote remote write data.
package promremote

import (
	"compress/gzip"
	"crypto/subtle"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/golang/snappy"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	iod "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	l                = logger.DefaultSLogger(inputName)
	_ inputs.InputV2 = (*Input)(nil)
)

const (
	body                   = "body"
	query                  = "query"
	defaultRemoteWritePath = "/prom_remote_write"
)

type Input struct {
	Path            string            `toml:"path"`
	Methods         []string          `toml:"methods"`
	DataSource      string            `toml:"data_source"`
	MaxBodySize     int64             `toml:"max_body_size"`
	BasicUsername   string            `toml:"basic_username"`
	BasicPassword   string            `toml:"basic_password"`
	HTTPHeaderTags  map[string]string `toml:"http_header_tags"`
	Tags            map[string]string `toml:"tags"`
	TagsIgnore      []string          `toml:"tags_ignore"`
	TagsIgnoreRegex []string          `toml:"tags_ignore_regex"`
	TagsRename      map[string]string `toml:"tags_rename"`
	Overwrite       bool              `toml:"overwrite"`
	Output          string            `toml:"output"`
	Parser
}

func (h *Input) RegHTTPHandler() {
	if h.Path == "" {
		h.Path = defaultRemoteWritePath
	}
	for _, m := range h.Methods {
		dkhttp.RegHTTPHandler(m, h.Path, h.ServeHTTP)
	}
}

func (h *Input) Catalog() string {
	return catalog
}

func (h *Input) Terminate() {
	// do nothing
}

func (h *Input) Run() {
	l.Infof("%s input started...", inputName)
	for i, m := range h.Methods {
		h.Methods[i] = strings.ToUpper(m)
	}
}

// ServeHTTP accepts prometheus remote writing, then parses received
// metrics, and sends them to datakit io or local disk file.
func (h *Input) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	handler := h.serveWrite

	h.authenticateIfSet(handler, res, req)
}

func (h *Input) authenticateIfSet(handler http.HandlerFunc, res http.ResponseWriter, req *http.Request) {
	if h.BasicUsername != "" && h.BasicPassword != "" {
		reqUsername, reqPassword, ok := req.BasicAuth()
		if !ok ||
			subtle.ConstantTimeCompare([]byte(reqUsername), []byte(h.BasicUsername)) != 1 ||
			subtle.ConstantTimeCompare([]byte(reqPassword), []byte(h.BasicPassword)) != 1 {
			http.Error(res, "Unauthorized.", http.StatusUnauthorized)
			return
		}
	}
	handler(res, req)
}

func (h *Input) serveWrite(res http.ResponseWriter, req *http.Request) {
	t := time.Now()
	// Check that the content length is not too large for us to handle.
	if req.ContentLength > h.MaxBodySize {
		if err := tooLarge(res); err != nil {
			l.Debugf("error in too-large: %v", err)
		}
		return
	}

	// Check if the requested HTTP method was specified in config.
	if !h.isAcceptedMethod(req.Method) {
		if err := methodNotAllowed(res); err != nil {
			l.Debugf("error in method-not-allowed: %v", err)
		}
		return
	}

	var bytes []byte
	var ok bool
	switch strings.ToLower(h.DataSource) {
	case query:
		bytes, ok = h.collectQuery(res, req)
	default:
		bytes, ok = h.collectBody(res, req)
	}
	if !ok {
		return
	}

	// If h.Output is configured, data is written to disk file path specified by h.Output.
	// Data will no more be written to datakit io.
	if h.Output != "" {
		err := h.writeFile(bytes)
		if err != nil {
			l.Warnf("fail to write data to file: %v", err)
		}
		res.WriteHeader(http.StatusNoContent)
		return
	}

	metrics, err := h.Parse(bytes)
	if err != nil {
		l.Debugf("parse error: %s", err.Error())
		if err := badRequest(res); err != nil {
			l.Debugf("error in bad-request: %v", err)
		}
		return
	}

	// Add HTTP header tags and custom tags.
	for i := range metrics {
		m, ok := metrics[i].(*Measurement)
		if !ok {
			l.Warnf("expect to be *Measurement")
			return
		}

		for headerName, tagName := range h.HTTPHeaderTags {
			headerValues := req.Header.Get(headerName)
			if len(headerValues) > 0 {
				m.tags[tagName] = headerValues
			}
		}
		h.SetTags(m)
	}
	if len(metrics) > 0 {
		if err := inputs.FeedMeasurement(inputName,
			datakit.Metric,
			metrics,
			&iod.Option{CollectCost: time.Since(t)}); err != nil {
			l.Warnf("inputs.FeedMeasurement: %s, ignored", err)
		}
	}
	res.WriteHeader(http.StatusNoContent)
}

func (h *Input) isAcceptedMethod(method string) bool {
	for _, m := range h.Methods {
		if method == m {
			return true
		}
	}
	return false
}

func (h *Input) SetTags(m *Measurement) {
	h.addTags(m)
	h.ignoreTags(m)
	h.ignoreTagsRegex(m)
	h.renameTags(m)
}

func (h *Input) addTags(m *Measurement) {
	for k, v := range h.Tags {
		m.tags[k] = v
	}
}

func (h *Input) ignoreTags(m *Measurement) {
	for _, t := range h.TagsIgnore {
		delete(m.tags, t)
	}
}

func (h *Input) ignoreTagsRegex(m *Measurement) {
	if len(h.TagsIgnoreRegex) == 0 {
		return
	}
	for tagKey := range m.tags {
		for _, r := range h.TagsIgnoreRegex {
			match, err := regexp.MatchString(r, tagKey)
			if err != nil {
				continue
			}
			if match {
				delete(m.tags, tagKey)
				break
			}
		}
	}
}

func (h *Input) renameTags(m *Measurement) {
	for oldKey, newKey := range h.TagsRename {
		if _, has := m.tags[oldKey]; !has {
			continue
		}
		_, has := m.tags[newKey]
		if has && h.Overwrite || !has {
			m.tags[newKey] = m.tags[oldKey]
			delete(m.tags, oldKey)
		}
	}
}

// writeFile writes data to path specified by h.Output.
// If file already exists, simply truncate it.
func (h *Input) writeFile(data []byte) error {
	fp := h.Output
	if !path.IsAbs(fp) {
		dir := datakit.InstallDir
		fp = filepath.Join(dir, fp)
	}

	f, err := os.Create(filepath.Clean(fp))
	if err != nil {
		return err
	}

	defer f.Close() //nolint:errcheck,gosec
	if _, err := f.Write(data); err != nil {
		return err
	}
	return nil
}

func (h *Input) collectBody(res http.ResponseWriter, req *http.Request) ([]byte, bool) {
	encoding := req.Header.Get("Content-Encoding")

	switch encoding {
	case "gzip":
		r, err := gzip.NewReader(req.Body)
		if err != nil {
			l.Debug(err.Error())
			if err := badRequest(res); err != nil {
				l.Debugf("error in bad-request: %v", err)
			}
			return nil, false
		}
		defer r.Close() //nolint:errcheck
		maxReader := http.MaxBytesReader(res, r, h.MaxBodySize)
		bytes, err := io.ReadAll(maxReader)
		if err != nil {
			if err := tooLarge(res); err != nil {
				l.Debugf("error in too-large: %v", err)
			}
			return nil, false
		}
		return bytes, true
	case "snappy":
		defer req.Body.Close() //nolint:errcheck
		bytes, err := io.ReadAll(req.Body)
		if err != nil {
			l.Debug(err.Error())
			if err := badRequest(res); err != nil {
				l.Debugf("error in bad-request: %v", err)
			}
			return nil, false
		}
		// snappy block format is only supported by decode/encode not snappy reader/writer
		bytes, err = snappy.Decode(nil, bytes)
		if err != nil {
			l.Debug(err.Error())
			if err := badRequest(res); err != nil {
				l.Debugf("error in bad-request: %v", err)
			}
			return nil, false
		}
		return bytes, true
	default:
		defer req.Body.Close() //nolint:errcheck
		bytes, err := io.ReadAll(req.Body)
		if err != nil {
			l.Debug(err.Error())
			if err := badRequest(res); err != nil {
				l.Debugf("error in bad-request: %v", err)
			}
			return nil, false
		}
		return bytes, true
	}
}

func (h *Input) collectQuery(res http.ResponseWriter, req *http.Request) ([]byte, bool) {
	rawQuery := req.URL.RawQuery

	query, err := url.QueryUnescape(rawQuery)
	if err != nil {
		l.Debugf("Error parsing query: %s", err.Error())
		if err := badRequest(res); err != nil {
			l.Debugf("error in bad-request: %v", err)
		}
		return nil, false
	}

	return []byte(query), true
}

func tooLarge(res http.ResponseWriter) error {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusRequestEntityTooLarge)
	_, err := res.Write([]byte(`{"error":"http: request body too large"}`))
	return err
}

func methodNotAllowed(res http.ResponseWriter) error {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusMethodNotAllowed)
	_, err := res.Write([]byte(`{"error":"http: method not allowed"}`))
	return err
}

func badRequest(res http.ResponseWriter) error {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusBadRequest)
	_, err := res.Write([]byte(`{"error":"http: bad request"}`))
	return err
}

func (h *Input) SampleConfig() string {
	return sample
}

func (h *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&Measurement{}}
}

func (h *Input) AvailableArchs() []string {
	return datakit.AllOS
}

func NewInput() *Input {
	i := Input{
		Methods:        []string{"POST", "PUT"},
		DataSource:     body,
		Tags:           map[string]string{},
		TagsRename:     map[string]string{},
		HTTPHeaderTags: map[string]string{},
		TagsIgnore:     []string{},
		MaxBodySize:    defaultMaxBodySize,
	}
	return &i
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return NewInput()
	})
}
