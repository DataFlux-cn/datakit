package rum

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	httpd "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/geo"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ip2isp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	"github.com/gin-gonic/gin"
	influxm "github.com/influxdata/influxdb1-client/models"
)

const (
	DEFAULT_PRECISION = "ns"
)

var (
	inputName                   = `rum`
	ipheaderName                = ""
	l            *logger.Logger = logger.DefaultSLogger(inputName)
)

func (_ *Rum) Catalog() string {
	return "rum"
}

func (_ *Rum) SampleConfig() string {
	return configSample
}

func (r *Rum) Run() {
}

func (r *Rum) Test() (result *inputs.TestResult, err error) {
	return
}

func (r *Rum) PipelineConfig() map[string]string {
	return map[string]string{
		inputName: pipelineSample,
	}
}

func (r *Rum) RegHttpHandler() {
	l = logger.SLogger(inputName)

	script := r.Pipeline
	if script == "" {
		scriptPath := filepath.Join(datakit.PipelineDir, inputName+".p")
		data, err := ioutil.ReadFile(scriptPath)
		if err == nil {
			script = string(data)
		}
	}

	if script != "" {
		r.pipelinePool = &sync.Pool{
			New: func() interface{} {
				p, err := pipeline.NewPipeline(script)
				if err != nil {
					l.Errorf("%s", err)
				}
				return p
			},
		}
	}

	ipheaderName = r.IPHeader
	httpd.RegGinHandler("POST", io.Rum, r.Handle)
}

func (r *Rum) getPipeline() *pipeline.Pipeline {
	if r.pipelinePool == nil {
		return nil
	}

	pl := r.pipelinePool.Get()
	if pl == nil {
		return nil
	}

	return pl.(*pipeline.Pipeline)
}

func (r *Rum) handleESData(pt influxm.Point, pl *pipeline.Pipeline) (influxm.Point, error) {

	if pl == nil {
		return pt, nil
	}

	pl = pl.RunPoint(pt)
	if err := pl.LastError(); err != nil {
		return nil, err
	}

	res, err := pl.Result()
	if err != nil {
		return nil, err
	}

	// XXX: use origin tag
	newpt, err := influxm.NewPoint(string(pt.Name()), pt.Tags(), res, pt.Time())

	if err != nil {
		return nil, err
	}

	return newpt, nil
}

func (r *Rum) handleBody(body []byte, precision, srcip string) (influxData, esdata []influxm.Point, err error) {

	pts, err := influxm.ParsePointsWithPrecision(body, time.Now().UTC(), precision)
	if err != nil {
		return nil, nil, err
	}

	pl := r.getPipeline()
	defer func() {
		if pl != nil {
			r.pipelinePool.Put(pl)
		}
	}()

	for _, pt := range pts {
		ptname := string(pt.Name())
		ipInfo, err := geo.Geo(srcip)
		if err != nil {
			l.Errorf("geo failed: %s, ignored", err)
		} else {
			// 无脑填充 geo 数据
			pt.AddTag("city", ipInfo.City)
			pt.AddTag("province", ipInfo.Region)
			pt.AddTag("country", ipInfo.Country_short)
			pt.AddTag("isp", ip2isp.SearchIsp(srcip))
		}

		if isMetric(ptname) {
			influxData = append(influxData, pt)
		} else if isES(ptname) {
			newpt, err := r.handleESData(pt, pl)

			// XXX: 只是上传 RUM 原始数据，因为 ES 会将所有 tag/field 拉平
			// 存入 ES，故只是行协议中增加一个长 tag，这个数据不会存入 influxdb。
			//
			// 这里只是将 RUM 原始数据转一下行协议即可，便于在 ES 中做全文检索。
			// NOTE: 如果做了 pipeline，一定不要 drop_origin_data()，即不要删除
			// `message' 字段，不然 RUM 原始数据无法做分词搜索。

			if err != nil {
				pt.AddTag("message", pt.String())
				esdata = append(esdata, pt)
			} else {
				newpt.AddTag("message", newpt.String())
				esdata = append(esdata, newpt)
			}
		} else {
			return nil, nil, fmt.Errorf("unknown RUM metric name `%s'", ptname)
		}
	}

	return
}

func (r *Rum) Handle(c *gin.Context) {

	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 2048)
			n := runtime.Stack(buf, false)
			l.Errorf("panic: %s", err)
			l.Errorf("%s", string(buf[:n]))
		}
	}()

	var precision string = DEFAULT_PRECISION
	var body []byte
	var err error
	srcip := ""

	precision, _ = uhttp.GinGetArg(c, "X-Precision", "precision")

	if ipheaderName != "" {
		srcip = c.Request.Header.Get(ipheaderName)
		if srcip != "" {
			parts := strings.Split(srcip, ",")
			if len(parts) > 0 {
				srcip = parts[0]
			}
		}
	}

	if srcip == "" {
		parts := strings.Split(c.Request.RemoteAddr, ":")
		if len(parts) > 0 {
			srcip = parts[0]
		}
	}

	body, err = uhttp.GinRead(c)
	if err != nil {
		l.Errorf("%s", err)
		uhttp.HttpErr(c, uhttp.Error(httpd.ErrHttpReadErr, err.Error()))
		return
	}

	metricpts, espts, err := r.handleBody(body, precision, srcip)
	if err != nil {
		uhttp.HttpErr(c, uhttp.Error(httpd.ErrBadReq, err.Error()))
		return
	}

	if err = io.NamedFeedPoints(metricpts, io.Metric, inputName); err != nil {
		uhttp.HttpErr(c, uhttp.Error(httpd.ErrBadReq, err.Error()))
		return
	}

	if err = io.NamedFeedPoints(espts, io.Rum, inputName); err != nil {
		uhttp.HttpErr(c, uhttp.Error(httpd.ErrBadReq, err.Error()))
		return
	}

	httpd.ErrOK.HttpBody(c, nil)
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Rum{}
	})
}
