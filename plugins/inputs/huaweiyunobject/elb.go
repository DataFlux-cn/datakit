package huaweiyunobject

import (
	"encoding/json"
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/sdk/huaweicloud"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/sdk/huaweicloud/elb"
)

const (
	classicType          = `经典型`
	sharedType           = `共享型`
	sharedTypeEnterprise = `共享型_企业项目`
	elbSampleConfig      = `
#[inputs.huaweiyunobject.elb]

## elb type: 经典型、共享型、共享型_企业项目 - requried
type=""

## 地区和终端节点 https://developer.huaweicloud.com/endpoint?ELB
endpoint=""

# ## @param - [list of Elb instanceid] - optional
#instanceids = []

# ## @param - [list of excluded Elb instanceid] - optional
#exclude_instanceids = []

# ## @param - custom tags for Elb object - [list of key:value element] - optional
#[inputs.huaweiyunobject.elb.tags]
# key1 = 'val1'
`
)

type Elb struct {
	Type     string `toml:"type"`
	EndPoint string `toml:"endpoint"`
	//	ProjectID          string            `toml:"project_id"`
	Tags               map[string]string `toml:"tags,omitempty"`
	InstancesIDs       []string          `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string          `toml:"exclude_instanceids,omitempty"`
}

func (e *Elb) run(ag *objectAgent) {

	if e.EndPoint == `` {
		e.EndPoint = fmt.Sprintf(`elb.%s.myhuaweicloud.com`, ag.RegionID)
	}

	cli := huaweicloud.NewHWClient(ag.AccessKeyID, ag.AccessKeySecret, e.EndPoint, ag.ProjectID, moduleLogger)

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		switch e.Type {
		case classicType:
			elistV1, err := cli.ElbV1List(nil)
			if err != nil {
				moduleLogger.Errorf(`get elblist v1, %v`, err)
				return
			}

			e.handResponseV1(elistV1, ag)

		case sharedType, sharedTypeEnterprise:
			//			moduleLogger.Debugf(`cli %+#v`, cli)
			e.doV2Action(cli, ag)

		default:
			moduleLogger.Warnf(`wrong type`)
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (e *Elb) doV2Action(cli *huaweicloud.HWClient, ag *objectAgent) {

	var marker string
	limit := 100
	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		opt := map[string]string{
			"limit":        fmt.Sprintf("%d", limit),
			"page_reverse": fmt.Sprintf("%v", true),
			"marker":       marker,
		}

		switch e.Type {
		case sharedType:
			elbsV20, err := cli.ElbV20List(opt)
			if err != nil {
				moduleLogger.Errorf(`get elblist v2.0, %v`, err)
				return
			}

			length := len(elbsV20.Loadbalancers)
			e.handResponseV2(elbsV20.Loadbalancers, ag)

			if length < limit {
				return
			}

			marker = elbsV20.Loadbalancers[length-1].ID
		case sharedTypeEnterprise:
			elbsV2, err := cli.ElbV2List(opt)
			if err != nil {
				moduleLogger.Errorf(`get elblist v2, %v`, err)
				return
			}

			length := len(elbsV2.Loadbalancers)
			e.handResponseV2(elbsV2.Loadbalancers, ag)

			if length < limit {
				return
			}

			marker = elbsV2.Loadbalancers[length-1].ID

		default:
			moduleLogger.Warnf(`wrong type`)
		}

	}

}

func (e *Elb) handResponseV1(resp *elb.ListLoadbalancersV1, ag *objectAgent) {

	moduleLogger.Debugf("Elb TotalCount=%d", resp.InstanceNum)
	var objs []map[string]interface{}

	for _, lb := range resp.Loadbalancers {

		if obj, err := datakit.CloudObject2Json(fmt.Sprintf(`%s(%s)`, lb.Name, lb.ID), `huaweiyun_elb`, lb, lb.ID, e.ExcludeInstanceIDs, e.InstancesIDs); obj != nil {
			objs = append(objs, obj)
		} else {
			if err != nil {
				moduleLogger.Errorf("%s", err)
			}
		}
	}

	if len(objs) <= 0 {
		return
	}

	data, err := json.Marshal(&objs)
	if err == nil {
		io.NamedFeed(data, io.Object, inputName)
	} else {
		moduleLogger.Errorf("%s", err)
	}
}

func (e *Elb) handResponseV2(lbs []elb.LoadbalancerV2, ag *objectAgent) {
	moduleLogger.Debugf("Elb TotalCount=%d", len(lbs))
	var objs []map[string]interface{}

	for _, lb := range lbs {

		if obj, err := datakit.CloudObject2Json(fmt.Sprintf(`%s(%s)`, lb.Name, lb.ID), `huaweiyun_elb`, lb, lb.ID, e.ExcludeInstanceIDs, e.InstancesIDs); obj != nil {
			objs = append(objs, obj)
		} else {
			if err != nil {
				moduleLogger.Errorf("%s", err)
			}
		}
	}

	if len(objs) <= 0 {
		return
	}

	data, err := json.Marshal(&objs)
	if err != nil {
		moduleLogger.Errorf("%s", err)
		return
	}
	io.NamedFeed(data, io.Object, inputName)
}
