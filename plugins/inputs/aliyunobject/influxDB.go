package aliyunobject

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/tidwall/gjson"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	influxDBSampleConfig = `
# ##(optional)
#[inputs.aliyunobject.influxdb]
    # ##(optional) ignore this object, default is false
    #disable = false

    # ##(optional) list of influxdb instanceid
    #instanceids = []

    # ##(optional) list of excluded influxdb instanceid
    #exclude_instanceids = []
`
)

type InfluxDB struct {
	Disable            bool              `toml:"disable"`
	Tags               map[string]string `toml:"tags,omitempty"`
	InstancesIDs       []string          `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string          `toml:"exclude_instanceids,omitempty"`
}

func (e *InfluxDB) disabled() bool {
	return e.Disable
}

func (e *InfluxDB) run(ag *objectAgent) {
	var cli *sdk.Client
	var err error

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err = sdk.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
		if err == nil {
			break
		}
		moduleLogger.Errorf("%s", err)
		datakit.SleepContext(ag.ctx, time.Second*3)
	}
	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}
		pageNum := 1
		pageSize := 100
		for {
			resp, err := DescribeHiTSDBInstanceList(*cli, pageSize, pageNum)

			select {
			case <-ag.ctx.Done():
				return
			default:
			}
			result := resp.GetHttpContentString()
			if err == nil {
				e.handleResponse(result, ag)
			} else {
				moduleLogger.Errorf("%s", err)
				break
			}

			if gjson.Get(result, "Total").Int() < gjson.Get(result, "PageSize").Int()*gjson.Get(result, "PageNumber").Int() {
				break
			}
			pageNum++
		}
		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func DescribeHiTSDBInstanceList(client sdk.Client, pageSize int, pageNumber int) (response *responses.CommonResponse, err error) {
	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Scheme = "https"
	request.Domain = "hitsdb.aliyuncs.com"
	request.Version = "2017-06-01"
	request.ApiName = "DescribeHiTSDBInstanceList"
	request.QueryParams["PageNumber"] = strconv.Itoa(pageNumber)
	request.QueryParams["PageSize"] = strconv.Itoa(pageSize)
	return client.ProcessCommonRequest(request)
}

func (e *InfluxDB) handleResponse(resp string, ag *objectAgent) {
	var objs []map[string]interface{}
	for _, inst := range gjson.Get(resp, "InstanceList").Array() {

		if len(e.ExcludeInstanceIDs) > 0 {
			exclude := false
			for _, v := range e.ExcludeInstanceIDs {
				if v == inst.Get("InstanceId").String() {
					exclude = true
					break
				}
			}
			if exclude {
				continue
			}
		}
		if len(e.InstancesIDs) > 0 {
			contain := false
			for _, v := range e.InstancesIDs {
				if v == inst.Get("InstanceId").String() {
					contain = true
					break
				}
			}
			if !contain {
				continue
			}
		}

		content := map[string]interface{}{
			`GmtCreated`:      inst.Get("GmtCreated").String(),
			`GmtExpire`:       inst.Get("GmtExpire").String(),
			`InstanceStorage`: inst.Get("InstanceStorage").String(),
			`UserId`:          inst.Get("UserId").String(),
			`InstanceId`:      inst.Get("InstanceId").String(),
			`ZoneId`:          inst.Get("ZoneId").String(),
			`ChargeType`:      inst.Get("ChargeType").String(),
			`InstanceStatus`:  inst.Get("InstanceStatus").String(),
			`NetworkType`:     inst.Get("NetworkType").String(),
			`RegionId`:        inst.Get("RegionId").String(),
			`EngineType`:      inst.Get("EngineType").String(),
			`InstanceClass`:   inst.Get("InstanceClass").String(),
		}

		jd, err := json.Marshal(content)
		if err != nil {
			moduleLogger.Errorf("%s", err)
			continue
		}

		obj := map[string]interface{}{
			`__name`:    inst.Get("InstanceAlias").String(),
			`__class`:   `aliyun_influxdb`,
			`__content`: string(jd),
		}

		objs = append(objs, obj)
	}

	if len(objs) <= 0 {
		return
	}
	data, err := json.Marshal(&objs)
	if err != nil {
		moduleLogger.Errorf("%s", err)
		return
	}
	if ag.isDebug() {
		fmt.Printf("%s\n", string(data))
	} else {
		io.NamedFeed(data, io.Object, inputName)
	}

}
