// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

//nolint:unused
import (
	"bufio"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/go-redis/redis/v8"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type replicaMeasurement struct {
	client   *redis.Client
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	resData  map[string]interface{}
	election bool
}

//nolint:lll
func (m *replicaMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: redisReplica,
		Type: "metric",
		Fields: map[string]interface{}{
			"repl_delay": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "replica delay",
			},
			"master_link_down_since_seconds": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "Number of seconds since the link is down",
			},
		},
		Tags: map[string]interface{}{
			"host":         &inputs.TagInfo{Desc: "Hostname"},
			"server":       &inputs.TagInfo{Desc: "Server addr"},
			"service_name": &inputs.TagInfo{Desc: "Service name"},
			"slave_id":     &inputs.TagInfo{Desc: "Slave ID"},
		},
	}
}

func (ipt *Input) collectReplicaMeasurement() ([]*point.Point, error) {
	m := &replicaMeasurement{
		client:   ipt.client,
		resData:  make(map[string]interface{}),
		tags:     make(map[string]string),
		fields:   make(map[string]interface{}),
		election: ipt.Election,
	}

	m.name = redisReplica
	setHostTagIfNotLoopback(m.tags, ipt.Host)

	if err := m.getData(); err != nil {
		return nil, err
	}

	if err := m.submit(); err != nil {
		l.Errorf("submit: %s", err)
	}
	var collectCache []*point.Point
	var opts []point.Option

	if m.election {
		m.tags = inputs.MergeTagsWrapper(m.tags, ipt.Tagger.ElectionTags(), ipt.Tags, ipt.Host)
	} else {
		m.tags = inputs.MergeTagsWrapper(m.tags, ipt.Tagger.HostTags(), ipt.Tags, ipt.Host)
	}

	pt := point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
	collectCache = append(collectCache, pt)
	return collectCache, nil
}

// 数据源获取数据.
func (m *replicaMeasurement) getData() error {
	ctx := context.Background()
	list, err := m.client.Info(ctx, "commandstats").Result()
	if err != nil {
		l.Error("redis exec `commandstats`, happen error,", err)
		return err
	}

	m.parseInfoData(list)

	return nil
}

var slaveMatch = regexp.MustCompile(`^slave\d+`)

// 解析返回.
func (m *replicaMeasurement) parseInfoData(list string) {
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

		// cmdstat_get:calls=2,usec=16,usec_per_call=8.00
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

		if slaveMatch.MatchString(key) {
			slaveID = strings.TrimPrefix(key, "slave")
			kv := strings.SplitN(value, ",", 5)
			if len(kv) != 5 {
				continue
			}

			split := strings.Split(kv[0], "=")
			if len(split) != 2 {
				l.Warnf("Failed to parse slave ip, got %s", kv[0])
				continue
			}
			ip = split[1]

			split = strings.Split(kv[1], "=")
			if len(split) != 2 {
				l.Warnf("Failed to parse slave port, got %s", kv[1])
				continue
			}
			port = split[1]

			split = strings.Split(kv[3], "=")
			if len(split) != 2 {
				l.Warnf("Failed to parse slave offset, got %s", kv[3])
				continue
			}

			if slaveOffset, err = strconv.ParseFloat(split[1], 64); err != nil {
				l.Warnf("ParseFloat: %s, slaveOffset expect to be int, got %s", err, split[1])
				continue
			}
		}

		delay := masterOffset - slaveOffset
		addr := fmt.Sprintf("%s:%s", ip, port)
		if addr != ":" {
			m.tags["slave_addr"] = fmt.Sprintf("%s:%s", ip, port)
		}

		m.tags["slave_id"] = slaveID

		if delay >= 0 {
			m.resData["repl_delay"] = delay
		}

		if masterStatus != "" {
			m.resData["master_link_down_since_seconds"] = masterDownSeconds
		}
	}
}

// 提交数据.
func (m *replicaMeasurement) submit() error {
	metricInfo := m.Info()
	for key, item := range metricInfo.Fields {
		if value, ok := m.resData[key]; ok {
			val, err := Conv(value, item.(*inputs.FieldInfo).DataType)
			if err != nil {
				l.Errorf("infoMeasurement metric %v value %v parse error %v", key, value, err)
				return err
			} else {
				m.fields[key] = val
			}
		}
	}

	return nil
}
