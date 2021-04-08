package redis

import (
	"bufio"
	"strings"
	"time"

	"github.com/go-redis/redis"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type commandMeasurement struct {
	client  *redis.Client
	name    string
	tags    map[string]string
	fields  map[string]interface{}
	ts      time.Time
	resData map[string]interface{}
}

func (m *commandMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *commandMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "redis_command_stat",
		Fields: map[string]*inputs.FieldInfo{
			"calls": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "this is CPU usage",
			},
			"usec": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "this is CPU usage",
			},
			"usec_per_call": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Desc:     "this is CPU usage",
			},
		},
		Tags: map[string]*inputs.TagInfo{
			"server": &inputs.TagInfo{
				Desc: "server addr",
			},
			"method": &inputs.TagInfo{
				Desc: "command type",
			},
		},
	}
}

// 数据源获取数据
func (i *Input) getCommandData() error {
	list, err := i.client.Info("commandstats").Result()
	if err != nil {
		return err
	}

	i.parseCommandData(list)

	return nil
}

// 解析返回结果
func (i *Input) parseCommandData(list string) error {
	rdr := strings.NewReader(list)

	scanner := bufio.NewScanner(rdr)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}

		m := &commandMeasurement{
			name:    "redis_command_stat",
			tags:    make(map[string]string),
			fields:  make(map[string]interface{}),
			resData: make(map[string]interface{}),
		}

		for key, value := range i.Tags {
			m.tags[key] = value
		}

		//cmdstat_get:calls=2,usec=16,usec_per_call=8.00
		method := parts[0]

		m.tags["method"] = method

		itemStrs := strings.Split(parts[1], ",")
		for _, itemStr := range itemStrs {
			item := strings.Split(itemStr, "=")

			key := item[0]
			val := strings.TrimSpace(item[1])

			m.resData[key] = val
		}

		m.submit()
		m.ts = time.Now()

		i.collectCache = append(i.collectCache, m)
	}

	return nil
}

// 提交数据
func (m *commandMeasurement) submit() error {
	metricInfo := m.Info()
	for key, item := range metricInfo.Fields {
		if value, ok := m.resData[key]; ok {
			val, err := Conv(value, item.DataType)
			if err != nil {
				l.Errorf("commandMeasurement metric %v value %v parse error %v", key, value, err)
			} else {
				m.fields[key] = val
			}
		}
	}

	return nil
}
