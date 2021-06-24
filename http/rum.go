package http

import (
	"fmt"
	"strings"
	"time"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/geo"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ip2isp"

	influxm "github.com/influxdata/influxdb1-client/models"
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

	l.Debugf("ipinfo(%s): %+#v", ipInfo, srcip)

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

func handleRUMBody(body []byte, precision, srcip string, isjson bool) ([]*io.Point, error) {
	extags := geoTags(srcip)

	if isjson {
		return jsonPoints(body, precision, extags)
	}

	rumpts, err := lp.ParsePoints(body, &lp.Option{
		Time:      time.Now(),
		Precision: precision,
		ExtraTags: extags,
		Strict:    true,

		// 由于 RUM 数据需要分别处理，故用回调函数来区分
		Callback: func(p influxm.Point) (influxm.Point, error) {
			name := string(p.Name())

			if _, ok := rumMetricNames[name]; !ok {
				return nil, fmt.Errorf("unknow RUM data-type %s", name)
			}

			// 移除 message 中可能的换行
			// 在行协议的 tag 上新增字段是比较方便的，而新增 field 则比较麻烦
			// 但奇怪的是，如果 tag-value 中有换行，拼接行协议不会报错，但 dataway
			// 解析行协议就报错了，尴尬

			// TODO: 此处需验证更多其它特殊字符，看啥时候会报错，以及在 tag 或
			// field 中是否会报错
			p.AddTag("message", strings.Replace(p.String(), "\n", "", -1))

			return p, nil
		},
	})

	if err != nil {
		l.Error(err)
		return nil, err
	}

	return io.WrapPoint(rumpts), nil
}

//func handleRUM(srcip, precision, input string, body []byte, json bool) ([]*io.Point, error) {
//
//	rumpts, err := handleRUMBody(body, precision, srcip, json)
//	if err != nil {
//		//uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
//		return nil, err
//	}
//
//	for _, pt := range rumpts {
//		x := pt.String()
//		l.Debugf("%s", x)
//		if err := lp.ParseLineProto([]byte(x), "n"); err != nil {
//			l.Errorf("parse failed: %s", err.Error())
//		} else {
//			l.Debug("parse ok")
//		}
//	}
//
//	if len(rumpts) > 0 {
//		if err = io.Feed(input, datakit.Rum, io.WrapPoint(rumpts), &io.Option{HighFreq: true}); err != nil {
//			uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
//			return
//		}
//	}
//
//	ErrOK.HttpBody(c, nil)
//}
