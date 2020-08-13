package aliyunobject

import (
	"encoding/json"
	"fmt"
	waf "github.com/aliyun/alibaba-cloud-sdk-go/services/waf-openapi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"time"
)

const (
	wafSampleConfig = `
#[inputs.aliyunobject.waf]

# ## @param - custom tags for waf object - [list of key:value element] - optional
#[inputs.aliyunobject.waf.tags]
# key1 = 'val1'
`
)

type Waf struct {
	Tags               map[string]string `toml:"tags,omitempty"`
}

func (e *Waf) run(ag *objectAgent) {
	var cli *waf.Client
	var err error

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err = waf.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
		if err == nil {
			break
		}
		moduleLogger.Errorf("%s", err)
		internal.SleepContext(ag.ctx, time.Second*3)
	}

	for {
		select {
		case <-ag.ctx.Done():
			return
		default:
		}
		req := waf.CreateDescribeInstanceInfoRequest()
		resp, err := cli.DescribeInstanceInfo(req)
		if err != nil {
			moduleLogger.Errorf("%s", err)
			break
		}
		e.handleResponse(resp, ag)
		internal.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (e *Waf) handleResponse(resp *waf.DescribeInstanceInfoResponse, ag *objectAgent) {
	if resp.InstanceInfo.PayType == 0 {
		moduleLogger.Warnf("%s", "waf payType 0")
		return
	}
	var objs []map[string]interface{}
	tags := map[string]interface{}{
		"__class":    "WAF",
		"__provider": "aliyun",
	}

	obj := map[string]interface{}{
		"__name": resp.InstanceInfo.InstanceId,

	}
	for k, v := range e.Tags {
		tags[k] = v
	}
	for k, v := range ag.Tags {
		if _, have := tags[k]; !have {
			tags[k] = v
		}
	}
	obj["__tags"] = tags
	objs = append(objs, obj)
	fmt.Println(objs)
	data, err := json.Marshal(&objs)
	if err == nil {
		io.NamedFeed(data, io.Object, inputName)
	} else {
		moduleLogger.Errorf("%s", err)
	}
}
