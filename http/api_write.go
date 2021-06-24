package http

import (
	"time"

	"github.com/gin-gonic/gin"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"

	influxdb "github.com/influxdata/influxdb1-client/v2"
)

func apiWrite(c *gin.Context) {
	var body []byte
	var err error

	input := DEFAULT_INPUT

	category := c.Request.URL.Path

	switch category {
	case datakit.Metric,
		datakit.Logging,
		datakit.Object,
		datakit.Tracing,
		datakit.KeyEvent:

	case datakit.CustomObject:
		input = "custom_object"

	case datakit.Rum:
		input = "rum"
	case datakit.Security:
		input = "sechecker"
	default:
		l.Debugf("invalid category: %s", category)
		uhttp.HttpErr(c, ErrInvalidCategory)
		return
	}

	if x := c.Query(INPUT); x != "" {
		input = x
	}

	precision := DEFAULT_PRECISION
	if x := c.Query(PRECISION); x != "" {
		precision = x
	}

	switch precision {
	case "h", "m", "s", "ms", "u", "n":
	default:
		l.Warnf("invalid precision %s", precision)
		uhttp.HttpErr(c, ErrInvalidPrecision)
		return
	}

	tags := extraTags
	if x := c.Query(IGNORE_GLOBAL_TAGS); x != "" {
		tags = nil
	}

	body, err = uhttp.GinRead(c)
	if err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrHttpReadErr, err.Error()))
		return
	}

	l.Debugf("body: %s", string(body))

	if category == datakit.Rum { // RUM 数据单独处理
		handleRUM(c, precision, input, body)
		return
	}

	pts, err := handleWriteBody(body, tags, precision)
	if err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}

	l.Debugf("received %d(%s) points from %s", len(pts), category, input)

	err = io.Feed(input, category, io.WrapPoint(pts), &io.Option{HighFreq: true})

	if err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
	} else {
		ErrOK.HttpBody(c, nil)
	}
}

func handleWriteBody(body []byte, tags map[string]string, precision string) (pts []*influxdb.Point, err error) {

	pts, err = lp.ParsePoints(body, &lp.Option{
		Time:      time.Now(),
		ExtraTags: tags,
		Strict:    true,
		Precision: precision})

	if err != nil {
		return nil, err
	}

	return
}
