// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package http

import (
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

type IApiWrite interface {
	sendToIO(string, string, []*point.Point, *io.Option) error
	geoInfo(string) map[string]string
}

type apiWriteImpl struct{}

type jsonPoint struct {
	Measurement string                 `json:"measurement"`
	Tags        map[string]string      `json:"tags,omitempty"`
	Fields      map[string]interface{} `json:"fields"`
	Time        int64                  `json:"time,omitempty"`
}

// convert json point to lineproto point.
func (jp *jsonPoint) point(opt *lp.Option) (*point.Point, error) {
	p, err := lp.MakeLineProtoPoint(jp.Measurement, jp.Tags, jp.Fields, opt)
	if err != nil {
		return nil, err
	}

	return &point.Point{Point: p}, nil
}

func (x *apiWriteImpl) sendToIO(input, category string, pts []*point.Point, opt *io.Option) error {
	return io.Feed(input, category, pts, opt)
}

func (x *apiWriteImpl) geoInfo(ip string) map[string]string {
	return geoTags(ip)
}

func apiWrite(w http.ResponseWriter, req *http.Request, x ...interface{}) (interface{}, error) {
	var body []byte
	var err error

	if x == nil || len(x) != 1 {
		l.Errorf("invalid handler")
		return nil, ErrInvalidAPIHandler
	}

	h, ok := x[0].(IApiWrite)
	if !ok {
		l.Errorf("not IApiWrite, got %s", reflect.TypeOf(x).String())
		return nil, ErrInvalidAPIHandler
	}

	input := DEFAULT_INPUT

	category := req.URL.Path

	switch category {
	case datakit.Metric,
		datakit.Network,
		datakit.Logging,
		datakit.Object,
		datakit.Tracing,
		datakit.KeyEvent:

	case datakit.CustomObject:
		input = "custom_object"

	case datakit.Security:
		input = "scheck"
	default:
		l.Debugf("invalid category: %s", category)
		return nil, ErrInvalidCategory
	}

	q := req.URL.Query()

	if x := q.Get(INPUT); x != "" {
		input = x
	}

	precision := DEFAULT_PRECISION
	if x := q.Get(PRECISION); x != "" {
		precision = x
	}

	// extraTags comes from global-host-tag or global-env-tags
	extraTags := map[string]string{}
	for _, arg := range []string{
		IGNORE_GLOBAL_HOST_TAGS,
		IGNORE_GLOBAL_TAGS, // deprecated
	} {
		if x := q.Get(arg); x != "" {
			extraTags = map[string]string{}
			break
		} else {
			for k, v := range point.GlobalHostTags() {
				l.Debugf("arg=%s, add host tag %s: %s", arg, k, v)
				extraTags[k] = v
			}
		}
	}

	if x := q.Get(GLOBAL_ELECTION_TAGS); x != "" {
		for k, v := range point.GlobalEnvTags() {
			l.Debugf("add env tag %s: %s", k, v)
			extraTags[k] = v
		}
	}

	var version string
	if x := q.Get(VERSION); x != "" {
		version = x
	}

	var pipelineSource string
	if x := q.Get(PIPELINE_SOURCE); x != "" {
		pipelineSource = x
	}

	switch precision {
	case "h", "m", "s", "ms", "u", "n":
	default:
		l.Warnf("invalid precision %s", precision)
		return nil, ErrInvalidPrecision
	}

	body, err = uhttp.ReadBody(req)
	if err != nil {
		return nil, err
	}

	if len(body) == 0 {
		return nil, ErrEmptyBody
	}

	isjson := (req.Header.Get("Content-Type") == "application/json")

	var pts []*point.Point

	opt := lp.NewDefaultOption()
	opt.Precision = precision
	opt.Time = time.Now()
	opt.ExtraTags = extraTags
	opt.Strict = true
	pts, err = handleWriteBody(body, isjson, opt)
	if err != nil {
		return nil, err
	}

	// check if object is ok
	if category == datakit.Object {
		for _, pt := range pts {
			if err := checkObjectPoint(pt); err != nil {
				return nil, err
			}
		}
	}

	if len(pts) == 0 {
		return nil, ErrNoPoints
	}

	l.Debugf("received %d(%s) points from %s, pipeline source: %v", len(pts), category, input, pipelineSource)

	if category == datakit.Logging && pipelineSource != "" {
		// Currently on logging support pipeline.
		// We try to find some @input.p to split logging, for example, if @input is nginx
		// the default pipeline is nginx.p.
		// If nginx.p missing, pipeline do nothing on incomming logging data.

		// for logging upload, we redirect them to pipeline
		l.Debugf("send pts to pipeline")
		err = h.sendToIO(input, category, pts, &io.Option{
			HighFreq: true, Version: version,
			PlScript: map[string]string{pipelineSource: pipelineSource + ".p"},
		})
	} else {
		err = h.sendToIO(input, category, pts, &io.Option{HighFreq: true, Version: version})
	}

	if err != nil {
		return nil, err
	}

	if q.Get(ECHO_LINE_PROTO) != "" {
		res := []*point.JSONPoint{}
		for _, pt := range pts {
			x, err := pt.ToJSON()
			if err != nil {
				l.Warnf("ToJSON: %s, ignored", err)
				continue
			}
			res = append(res, x)
		}

		return res, nil
	}

	return nil, nil
}

func handleWriteBody(body []byte, isJSON bool, opt *lp.Option) ([]*point.Point, error) {
	switch isJSON {
	case true:
		return jsonPoints(body, opt)

	default:
		pts, err := lp.ParsePoints(body, opt)
		if err != nil {
			return nil, uhttp.Error(ErrInvalidLinePoint, err.Error())
		}

		return point.WrapPoint(pts), nil
	}
}

func jsonPoints(body []byte, opt *lp.Option) ([]*point.Point, error) {
	var jps []jsonPoint
	err := json.Unmarshal(body, &jps)
	if err != nil {
		l.Error(err)
		return nil, ErrInvalidJSONPoint
	}

	if opt == nil {
		opt = lp.DefaultOption
	}

	var pts []*point.Point
	for _, jp := range jps {
		if jp.Time != 0 { // use time from json point
			opt.Time = time.Unix(0, jp.Time)
		}

		if p, err := jp.point(opt); err != nil {
			l.Error(err)
			return nil, uhttp.Error(ErrInvalidJSONPoint, err.Error())
		} else {
			pts = append(pts, p)
		}
	}
	return pts, nil
}

func checkObjectPoint(p *point.Point) error {
	tags := p.Point.Tags()
	if _, ok := tags["name"]; !ok {
		return ErrInvalidObjectPoint
	}
	return nil
}
