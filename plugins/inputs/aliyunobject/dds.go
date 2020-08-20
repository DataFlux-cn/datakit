package aliyunobject

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	ddsSampleConfig = `
#[inputs.aliyunobject.dds]

# ## @param - custom tags for dds object - [list of key:value element] - optional
#[inputs.aliyunobject.dds.tags]
# key1 = 'val1'

# ## @param - custom tags - [list of dds instanceid] - optional
#db_instanceids = []

# ## @param - custom tags - [list of excluded dds instanceid] - optional
#exclude_db_instanceids = []
`
)

type Dds struct {
	Tags                 map[string]string `toml:"tags,omitempty"`
	DBInstancesIDs       []string          `toml:"db_instanceids,omitempty"`
	ExcludeDBInstanceIDs []string          `toml:"exclude_db_instanceids,omitempty"`
}

func (r *Dds) run(ag *objectAgent) {
	var cli *dds.Client
	var err error

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err = dds.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
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

		pageNum := 1
		pageSize := 100
		req := dds.CreateDescribeDBInstancesRequest()
		req.Scheme = "https"

		for {
			moduleLogger.Infof("pageNume %v, pagesize %v", pageNum, pageSize)
			if len(r.DBInstancesIDs) > 0 {
				if pageNum <= len(r.DBInstancesIDs) {
					req.DBInstanceId = r.DBInstancesIDs[pageNum-1]
				} else {
					break
				}
			} else {
				req.PageNumber = requests.NewInteger(pageNum)
				req.PageSize = requests.NewInteger(pageSize)
			}
			resp, err := cli.DescribeDBInstances(req)

			select {
			case <-ag.ctx.Done():
				return
			default:
			}

			if err == nil {
				r.handleResponse(resp, ag)
			} else {
				moduleLogger.Errorf("%s", err)
				if len(r.DBInstancesIDs) > 0 {
					pageNum++
					continue
				}
				break
			}

			if len(r.DBInstancesIDs) <= 0 && resp.TotalCount < resp.PageNumber*pageSize {
				break
			}

			pageNum++
			if len(r.DBInstancesIDs) <= 0 {
				req.PageNumber = requests.NewInteger(pageNum)
			}
		}

		internal.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (r *Dds) handleResponse(resp *dds.DescribeDBInstancesResponse, ag *objectAgent) {

	moduleLogger.Debugf("TotalCount=%d, PageSize=%v, PageNumber=%v", resp.TotalCount, resp.PageSize, resp.PageNumber)

	var objs []*map[string]interface{}

	for _, db := range resp.DBInstances.DBInstance {
		//moduleLogger.Debugf("dbinstanceInfo %+#v", db)

		exclude := false
		for _, dbIsId := range ag.Dds.ExcludeDBInstanceIDs {
			if db.DBInstanceId == dbIsId {
				exclude = true
				break
			}
		}

		if exclude {
			continue
		}

		tags := map[string]interface{}{
			"__class":               "dds",
			"__provider":            "aliyun",
			"DBInstanceDescription": db.DBInstanceDescription,
			"DBInstanceId":          db.DBInstanceId,
			"DBInstanceType":        db.DBInstanceType,
			"RegionId":              db.RegionId,
			"DBInstanceStatus":      db.DBInstanceStatus,
			"Engine":                db.Engine,
			"NetworkType":           db.NetworkType,
			"LockMode":              db.LockMode,
			"DBInstanceClass":       db.DBInstanceClass,
			"DBInstanceStorage":     db.DBInstanceStorage,
			"EngineVersion":         db.EngineVersion,
			"ResourceGroupId":       db.ResourceGroupId,
			"VSwitchId":             db.VSwitchId,
			"VpcCloudInstanceId":    db.VPCCloudInstanceIds,
			"VPCId":                 db.VPCId,
			"ZoneId":                db.ZoneId,
		}

		for _, t := range db.Tags.Tag {
			tags[t.Key] = t.Value
		}

		//add dds object custom tags
		for k, v := range r.Tags {
			tags[k] = v
		}

		//add global tags
		for k, v := range ag.Tags {
			if _, have := tags[k]; !have {
				tags[k] = v
			}
		}

		obj := &map[string]interface{}{
			"__name":                fmt.Sprintf(`%s_%s`, db.DBInstanceDescription, db.DBInstanceId),
			"__tags":                tags,
			"ExpireTime":            db.ExpireTime,
			"DestroyTime":           db.DestroyTime,
			"CreationTime":          db.CreationTime,
			"DBInstanceDescription": db.DBInstanceDescription,
			"MaintainStartTime":     db.MaintainStartTime,
			"MaxIOPS":               db.MaxIOPS,
			"MaintainEndTime":       db.MaintainEndTime,
			"LastDowngradeTime":     db.LastDowngradeTime,
			"ChargeType":            db.ChargeType,
			"ReadonlyReplicas":      db.ReadonlyReplicas,
			"VpcAuthMode":           db.VpcAuthMode,
			"MaxConnections":        db.MaxConnections,
			"ReplicationFactor":     db.ReplicationFactor,
			"CurrentKernelVersion":  db.CurrentKernelVersion,
			"ConfigserverList":      db.ConfigserverList,
			"ShardList":             db.ShardList,
			"ReplicaSets":           db.ReplicaSets,
			"MongosList":            db.MongosList,
		}

		objs = append(objs, obj)
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
