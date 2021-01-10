package aliyunobject

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

const (
	ecsSampleConfig = `
# ##(optional)
#[inputs.aliyunobject.ecs]
    # ##(optional) ignore this object, default is false
    #disable = false

    # ##(optional) list of ecs instanceid
    #instanceids = ['']

    # ##(optional) list of excluded ecs instanceid
    #exclude_instanceids = ['']
`

	ecsPipelineConifg = `

json(_,InstanceName,name);
json(_,RegionId);
json(_,Status);
json(_,InstanceId);
json(_,NetworkType);
json(_,InstanceChargeType);

`
)

type Ecs struct {
	Disable            bool              `toml:"disable"`
	Tags               map[string]string `toml:"tags,omitempty"`
	InstancesIDs       []string          `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string          `toml:"exclude_instanceids,omitempty"`
	PipelinePath       string            `toml:"pipeline,omitempty"`

	p *pipeline.Pipeline
}

func (e *Ecs) disabled() bool {
	return e.Disable
}

func (e *Ecs) run(ag *objectAgent) {
	var cli *ecs.Client
	var err error
	p, err := newPipeline(e.PipelinePath)
	if err != nil {
		moduleLogger.Errorf("%s", err.Error())
		return
	}
	e.p = p

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err = ecs.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
		if err == nil {
			break
		} else {
			moduleLogger.Errorf("%s", err)
			if ag.isTest() {
				ag.testError = err
				return
			}
		}
		datakit.SleepContext(ag.ctx, time.Second*3)
	}

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		pageNum := 1
		req := ecs.CreateDescribeInstancesRequest()
		req.PageNumber = requests.NewInteger(pageNum)
		req.PageSize = requests.NewInteger(100)

		if len(e.InstancesIDs) > 0 {
			data, err := json.Marshal(e.InstancesIDs)
			if err == nil {
				req.InstanceIds = string(data)
			}
		}

		for {
			resp, err := cli.DescribeInstances(req)

			select {
			case <-ag.ctx.Done():
				return
			default:
			}

			if err == nil {
				e.handleResponse(resp, ag)
			} else {
				moduleLogger.Errorf("%s", err)
				if ag.isTest() {
					ag.testError = err
					return
				}
				break
			}

			if resp.TotalCount < resp.PageNumber*resp.PageSize {
				break
			}
			pageNum++
			req.PageNumber = requests.NewInteger(pageNum)
		}

		if ag.isTest() {
			break
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (e *Ecs) handleResponse(resp *ecs.DescribeInstancesResponse, ag *objectAgent) {

	for _, inst := range resp.Instances.Instance {
		ag.parseObject(inst, "aliyun_ecs", fmt.Sprintf(`%s(%s)`, inst.InstanceName, inst.InstanceId), inst.InstanceId, e.p, e.ExcludeInstanceIDs, e.InstancesIDs, e.Tags)
	}

}
