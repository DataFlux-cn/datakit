package redis

import (
	"bufio"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type replicaMeasurement struct {
	client  *redis.Client
	name    string
	tags    map[string]string
	fields  map[string]interface{}
	ts      time.Time
	resData map[string]interface{}
}

func (m *replicaMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *replicaMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "redis_client",
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
	}
}

func CollectReplicaMeasurement(cli *redis.Client) *replicaMeasurement {
	m := &replicaMeasurement{
		client:  cli,
		resData: make(map[string]interface{}),
		tags:    make(map[string]string),
		fields:  make(map[string]interface{}),
	}

	m.getData()
	m.submit()

	return m
}

// 数据源获取数据
func (m *replicaMeasurement) getData() error {
	list, err := m.client.Info("commandstats").Result()
	if err != nil {
		return err
	}
	m.parseInfoData(list)

	return nil
}

// 解析返回结果
func (m *replicaMeasurement) parseInfoData(list string) error {
	var masterDownSeconds, masterOffset, slaveOffset float64
	var masterStatus, slaveID, ip, port string
	var err error

	rdr := strings.NewReader(list)
	scanner := bufio.NewScanner(rdr)
	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 || line[0] == '#' {
			continue
		}

		record := strings.SplitN(line, ":", 2)
		if len(record) < 2 {
			continue
		}

		//cmdstat_get:calls=2,usec=16,usec_per_call=8.00
		key, value := record[0], record[1]

		if key == "master_repl_offset" {
			masterOffset, _ = strconv.ParseFloat(value, 64)
		}

		if key == "master_link_down_since_seconds" {
			masterDownSeconds, _ = strconv.ParseFloat(value, 64)
		}

		if key == "master_link_status" {
			masterStatus = value
		}

		if re, _ := regexp.MatchString(`^slave\d+`, key); re {
			slaveID = strings.TrimPrefix(key, "slave")
			kv := strings.SplitN(value, ",", 5)
			if len(kv) != 5 {
				continue
			}

			split := strings.Split(kv[0], "=")
			if len(split) != 2 {
				l.Warnf("Failed to parse slave ip. %s", err)
				continue
			}
			ip = split[1]

			split = strings.Split(kv[1], "=")
			if err != nil {
				l.Warnf("Failed to parse slave port. %s", err)
				continue
			}
			port = split[1]

			split = strings.Split(kv[3], "=")
			if err != nil {
				l.Warnf("Failed to parse slave offset. %s", err)
				continue
			}
			slaveOffset, _ = strconv.ParseFloat(split[1], 64)
		}

		delay := masterOffset - slaveOffset
		m.tags["slave_addr"] = fmt.Sprintf("%s:%s", ip, port)
		m.tags["slave_id"] = slaveID

		if delay >= 0 {
			m.resData["repl_delay"] = delay
		}

		if masterStatus != "" {
			m.resData["master_link_down_since_seconds"] = masterDownSeconds
		}
	}

	return nil
}

// 提交数据
func (m *replicaMeasurement) submit() error {
	metricInfo := m.Info()
	for key, item := range metricInfo.Fields {
		if value, ok := m.resData[key]; ok {
			val, err := Conv(value, item.DataType)
			if err != nil {
				l.Errorf("infoMeasurement metric %v value %v parse error %v", key, value, err)
			} else {
				m.fields[key] = val
			}
		}
	}

	return nil
}
