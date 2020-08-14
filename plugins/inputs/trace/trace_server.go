package trace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/ftagent/utils"
)

type Reply struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}
type TraceDecoder interface {
	Decode(octets []byte) error
}

type TraceReqInfo struct {
	Source      string
	Version     string
	ContentType string
}

type ZipkinTracer struct {
	TraceReqInfo
}

type JaegerTracer struct {
	TraceReqInfo
}

type TraceAdapter struct {
	Source string

	Duration    int64
	TimestampUs int64
	Content     string

	Class         string
	ServiceName   string
	OperationName string
	ParentID      string
	TraceID       string
	SpanID        string
	IsError       string
	SpanType      string
	EndPoint      string

	Tags          map[string]string
}

const (
	US_PER_SECOND   int64 = 1000000
	SPAN_TYPE_ENTRY       = "entry"
	SPAN_TYPE_LOCAL       = "local"
	SPAN_TYPE_EXIT        = "exit"
)

var SkyWalkTraceUrl  = map[string]bool{
	SKYWALK_SEGMENT    : true,
	SKYWALK_PROPERTIES : true,
	SKYWALK_KEEPALIVE  : true,
	SKYWALK_SEGMENTS   : true,
}

func (tAdpt *TraceAdapter) MkLineProto() ([]byte, error) {
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["__class"]         = tAdpt.Class
	tags["__operationName"] = tAdpt.OperationName
	tags["__serviceName"]   = tAdpt.ServiceName
	tags["__parentID"]      = tAdpt.ParentID
	tags["__traceID"]       = tAdpt.TraceID
	tags["__spanID"]        = tAdpt.SpanID

	for tag, tagV := range tAdpt.Tags {
		tags[tag] = tagV
	}
	if tAdpt.IsError == "true" {
		tags["__isError"] = "true"
	} else {
		tags["__isError"] = "false"
	}

	if tAdpt.EndPoint != "" {
		tags["__endpoint"] = tAdpt.EndPoint
	} else {
		tags["__endpoint"] = "null"
	}

	if tAdpt.SpanType != "" {
		tags["__spanType"] = tAdpt.SpanType
	} else {
		tags["__spanType"] = SPAN_TYPE_ENTRY
	}

	fields["__duration"] = tAdpt.Duration
	fields["__content"]  = tAdpt.Content

	ts := time.Unix(tAdpt.TimestampUs/US_PER_SECOND, (tAdpt.TimestampUs%US_PER_SECOND)*1000)

	pt, err := io.MakeMetric(tAdpt.Source, tags, fields, ts)
	if err != nil {
		return nil, fmt.Errorf("build metric err: %s", err)
	}

	lineProtoStr := string(pt)
	log.Debugf(lineProtoStr)

	return pt, nil
}

func (t *TraceReqInfo) Decode(octets []byte) error {
	var decoder TraceDecoder
	source := strings.ToLower(t.Source)

	switch source {
	case "zipkin":
		decoder = &ZipkinTracer{*t}
	case "jaeger":
		decoder = &JaegerTracer{*t}
	default:
		return fmt.Errorf("Unsupported trace source %s", t.Source)
	}

	return decoder.Decode(octets)
}

func Handle(w http.ResponseWriter, r *http.Request) {
	log.Debugf("trace handle with path: %s", r.URL.Path)
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Stack crash: %v", r)
			log.Errorf("Stack info :%s", string(debug.Stack()))
		}
	}()

	if err := handleTrace(w, r); err != nil {
		log.Errorf("%v", err)
	}
}

func handleTrace(w http.ResponseWriter, r *http.Request) error {
	source := r.URL.Query().Get("source")
	version := r.URL.Query().Get("version")
	contentType := r.Header.Get("Content-Type")
	contentEncoding := r.Header.Get("Content-Encoding")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		JsonReply(w, http.StatusBadRequest, "Read body err: %s", err)
		return err
	}
	defer r.Body.Close()

	if contentEncoding == "gzip" {
		body, err = utils.ReadCompressed(bytes.NewReader(body), true)
		if err != nil {
			JsonReply(w, http.StatusBadRequest, "Uncompress body err: %s", err)
			return err
		}
	}

	if _, ok := SkyWalkTraceUrl[r.URL.Path]; ok {
		return handleSkyWalking(w, r, r.URL.Path, body)
	} else {
		tInfo := TraceReqInfo{source, version, contentType}
		err = tInfo.Decode(body)
		if err != nil {
			JsonReply(w, http.StatusBadRequest, "Parse trace err: %s", err)
			return err
		}

		JsonReply(w, http.StatusOK, "ok")
		return nil
	}
}

func JsonReply(w http.ResponseWriter, code int, strfmt string, args ...interface{}) {
	msg := fmt.Sprintf(strfmt, args...)
	w.WriteHeader(code)

	r, err := json.Marshal(Reply{
		Code: code,
		Msg:  msg,
	})
	if err == nil {
		w.Write(r)
	}
}
