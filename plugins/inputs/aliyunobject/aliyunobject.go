package aliyunobject

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName    = `aliyunobject`
	moduleLogger = logger.DefaultSLogger("aliyunobject")
	sampleConf   = ""
)

type subModule interface {
	run(*objectAgent)
	disabled() bool
}

func (_ *objectAgent) SampleConfig() string {
	var buf bytes.Buffer
	buf.WriteString(sampleConfig)
	buf.WriteString(ecsSampleConfig)
	buf.WriteString(slbSampleConfig)
	buf.WriteString(ossSampleConfig)
	buf.WriteString(rdsSampleConfig)
	buf.WriteString(redisSampleConfig)
	buf.WriteString(cdnSampleConfig)
	buf.WriteString(wafSampleConfig)
	buf.WriteString(elasticsearchSampleConfig)
	buf.WriteString(influxDBSampleConfig)
	buf.WriteString(onsSampleConfig)
	buf.WriteString(domainSampleConfig)
	buf.WriteString(ddsSampleConfig)
	return buf.String()
}

func (_ *objectAgent) Catalog() string {
	return `aliyun`
}

func (_ *objectAgent) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		"aliyun_redis":         redisPipelineConifg,
		"aliyun_waf":           wafPipelineConfig,
		"aliyun_cdn":           cdnPipelineConifg,
		"aliyun_elasticsearch": elasticsearchPipelineConifg,
		"aliyun_influxdb":      influxDBPipelineConfig,
		"aliyun_ecs":           ecsPipelineConifg,
	}
	return pipelineMap
}

func (ag *objectAgent) Test() (*inputs.TestResult, error) {
	ag.mode = "test"
	ag.testResult = &inputs.TestResult{}
	ag.Run()
	return ag.testResult, ag.testError
}

func (ag *objectAgent) Run() {

	moduleLogger = logger.SLogger(inputName)

	ag.ctx, ag.cancelFun = context.WithCancel(context.Background())

	go func() {
		<-datakit.Exit.Wait()
		ag.cancelFun()
	}()

	if ag.Interval.Duration == 0 {
		ag.Interval.Duration = time.Minute * 5
	}

	if ag.Ecs == nil {
		ag.Ecs = &Ecs{}
	}
	if ag.Slb != nil {
		ag.Slb = &Slb{}
	}
	if ag.Oss == nil {
		ag.Oss = &Oss{}
	}
	if ag.Rds == nil {
		ag.Rds = &Rds{}
	}
	if ag.Ons == nil {
		ag.Ons = &Ons{}
	}
	if ag.Dds == nil {
		ag.Dds = &Dds{}
	}
	if ag.Domain == nil {
		ag.Domain = &Domain{}
	}
	if ag.Redis == nil {
		ag.Redis = &Redis{
			PipelinePath: "aliyun_redis.p",
		}
	}
	if ag.Cdn == nil {
		ag.Cdn = &Cdn{
			PipelinePath: "aliyun_cdn.p",
		}
	}
	if ag.Waf == nil {
		ag.Waf = &Waf{
			PipelinePath: "aliyun_waf.p",
		}
	}
	if ag.Es == nil {
		ag.Es = &Elasticsearch{
			PipelinePath: "aliyun_elasticsearch.p",
		}
	}
	if ag.InfluxDB != nil {
		ag.InfluxDB = &InfluxDB{
			PipelinePath: "aliyun_influxdb.p",
		}
	}

	ag.addModule(ag.Ecs)
	ag.addModule(ag.Slb)
	ag.addModule(ag.Oss)
	ag.addModule(ag.Rds)
	ag.addModule(ag.Ons)
	ag.addModule(ag.Dds)
	ag.addModule(ag.Domain)
	ag.addModule(ag.Redis)
	ag.addModule(ag.Cdn)
	ag.addModule(ag.Waf)
	ag.addModule(ag.Es)
	ag.addModule(ag.InfluxDB)

	for _, s := range ag.subModules {
		ag.wg.Add(1)
		go func(s subModule) {
			defer ag.wg.Done()
			s.run(ag)
		}(s)
	}

	ag.wg.Wait()

	moduleLogger.Debugf("done")
}

func newAgent() *objectAgent {
	ag := &objectAgent{}
	return ag
}

func newPipeline(pipelinePath string) (*pipeline.Pipeline, error) {
	scriptPath := filepath.Join(datakit.PipelineDir, pipelinePath)
	data, err := ioutil.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}
	p, err := pipeline.NewPipeline(string(data))
	return p, err
}

func (ag *objectAgent) parseObject(obj interface{}, class, name, id string, pipeline *pipeline.Pipeline, blacklist, whitelist []string, tags map[string]string) {
	if datakit.CheckExcluded(id, blacklist, whitelist) {
		return
	}
	data, err := json.Marshal(obj)
	if err != nil {
		moduleLogger.Errorf("[error] json marshal err:%s", err.Error())
		return
	}
	if tags == nil {
		tags = map[string]string{}
	}
	for k, v := range ag.Tags {
		if _, ok := tags[k]; ok {
			continue
		} else {
			tags[k] = v
		}
	}
	fields, err := pipeline.Run(string(data)).Result()
	if err != nil {
		moduleLogger.Errorf("[error] pipeline run err:%s", err.Error())
		return
	}
	fields["content"] = string(data)

	tags["class"] = class
	tags["name"] = name

	io.NamedFeedEx(inputName, io.Object, class, tags, fields, time.Now().UTC())
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newAgent()
	})
}
