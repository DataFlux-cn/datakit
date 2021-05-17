package http

import (
	"fmt"
	"strings"
	"time"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/geo"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ip2isp"

	"github.com/gin-gonic/gin"
	influxm "github.com/influxdata/influxdb1-client/models"
	influxdb "github.com/influxdata/influxdb1-client/v2"
)

var (
	rumMetricNames = map[string]bool{
		`view`:      true,
		`resource`:  true,
		`error`:     true,
		`long_task`: true,
		`action`:    true,
	}
)

func geoTags(srcip string) (tags map[string]string) {
	tags = map[string]string{}

	ipInfo, err := geo.Geo(srcip)

	l.Debugf("ipinfo: %+#v", ipInfo)

	if err != nil {
		l.Errorf("geo failed: %s, ignored", err)
		return
	} else {
		// 无脑填充 geo 数据
		tags = map[string]string{
			"city":     ipInfo.City,
			"province": ipInfo.Region,
			"country":  ipInfo.Country_short,
			"isp":      ip2isp.SearchIsp(srcip),
			"ip":       srcip,
		}
	}

	return
}

func handleRUMBody(body []byte, precision, srcip string) (rumpts []*influxdb.Point, err error) {
	extraTags := geoTags(srcip)

	rumpts, err = lp.ParsePoints(body, &lp.Option{
		Time:      time.Now(),
		Precision: precision,
		ExtraTags: extraTags,
		Strict:    true,

		// 由于 RUM 数据需要分别处理，故用回调函数来区分
		Callback: func(p influxm.Point) (influxm.Point, error) {
			name := string(p.Name())

			if _, ok := rumMetricNames[name]; !ok {
				return nil, fmt.Errorf("unknow RUM data-type %s", name)
			}

			p.AddTag("message", p.String())

			return p, nil
		},
	})

	if err != nil {
		l.Error(err)
		return nil, err
	}

	return rumpts, nil
}

func handleRUM(c *gin.Context, precision, input string, body []byte) {

	srcip := c.Request.Header.Get(datakit.Cfg.HTTPAPI.RUMOriginIPHeader)
	if srcip != "" {
		parts := strings.Split(srcip, ",")
		if len(parts) > 0 {
			srcip = parts[0] // 注意：此处只取第一个 IP 作为源 IP
		}
	} else { // 默认取 gin 框架带进来的 IP
		parts := strings.Split(c.Request.RemoteAddr, ":")
		if len(parts) > 0 {
			srcip = parts[0]
		}
	}

	rumpts, err := handleRUMBody(body, precision, srcip)
	if err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}

	if input == DEFAULT_INPUT { // RUM 默认源不好直接用 datakit，故单独以 `rum' 标记之
		input = "rum"
	}

	if len(rumpts) > 0 {
		if err = io.Feed(input, datakit.Rum, io.WrapPoint(rumpts), &io.Option{HighFreq: true}); err != nil {
			uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
			return
		}
	}

	ErrOK.HttpBody(c, nil)
}
