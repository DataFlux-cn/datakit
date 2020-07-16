package dataclean

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	influxm "github.com/influxdata/influxdb1-client/models"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"

	"gitlab.jiagouyun.com/cloudcare-tools/ftagent/utils"
)

func getPromb(compressed []byte) (wr *prompb.WriteRequest, err error) {

	reqBuf, err := snappy.Decode(nil, compressed)
	if err != nil {
		moduleLogger.Errorf("snappy decode failed, %s", err.Error())
		err = utils.ErrSnappyDecodeFailed
		return
	}

	wr = &prompb.WriteRequest{}

	err = proto.Unmarshal(reqBuf, wr)
	if err != nil {
		moduleLogger.Errorf("[error] proto unmarshal failed, %s", err.Error())
		err = utils.ErrProtobufDecodeFailed
		return
	}
	return
}

// prom 格式相关数据
func ParsePromToInflux(data []byte, template string) ([]*influxdb.Point, error) {
	req, err := getPromb(data)
	if err != nil {
		return nil, err
	}

	pts := []*influxdb.Point{}

	for _, ts := range req.Timeseries {

		var measurement string
		filedName := "value"
		tags := map[string]string{}

		for _, l := range ts.Labels {
			if l.Name == model.MetricNameLabel {
				measurement = l.Value
				parts := strings.Split(l.Value, "_")
				if len(parts) > 2 {
					measurement = strings.Join(parts[:2], "_")
					filedName = strings.Join(parts[2:], "_")
				}
				continue
			}
			tags[l.Name] = l.Value
		}

		for _, s := range ts.Samples {
			if s.Value >= math.MaxInt64 {
				moduleLogger.Warnf("invalid data type, value %v", s.Value)
				continue
			}

			if math.IsNaN(s.Value) {
				continue
			}

			fields := map[string]interface{}{
				filedName: s.Value,
			}

			pt, err := influxdb.NewPoint(
				measurement,
				tags,
				fields,
				time.Unix(s.Timestamp/1000, 0),
			)

			if err != nil {
				moduleLogger.Warnf("assembly influx failed,  %s", err.Error())
				continue
			}

			pts = append(pts, pt)
		}
	}

	moduleLogger.Debugf("prom points %d", len(pts))
	return pts, nil
}

// telegraf json 格式
func ParseJsonToInflux(data []byte, template string) ([]*influxdb.Point, error) {
	m := make(map[string][]interface{})
	if err := json.Unmarshal(data, &m); err != nil {
		moduleLogger.Warnf("data unmarshal failed,  %s ", string(data))
		return nil, err
	}

	ms := m["metrics"]
	if ms == nil {
		moduleLogger.Warnf("no any metrics")
		return nil, utils.ErrNoMetricAvailable
	}

	pts := []*influxdb.Point{}
	for _, m := range ms {

		namePrefix := ""
		var metric map[string]interface{}
		ok := false
		if metric, ok = m.(map[string]interface{}); !ok {
			continue
		}
		if namePrefix, ok = metric["name"].(string); !ok {
			continue
		}

		var metricTags map[string]interface{}
		if metric["tags"] != nil {
			if metricTags, ok = metric["tags"].(map[string]interface{}); !ok {
				continue
			}
		}

		tags := map[string]string{}
		for k, v := range metricTags {
			strK := fmt.Sprintf("%v", k)
			strV := fmt.Sprintf("%v", v)

			tags[strK] = strV
		}

		var timestamp int64
		if ft, ok := metric["timestamp"].(float64); ok {
			timestamp = int64(ft)
		}

		var fields map[string]interface{}
		if fields, ok = metric["fields"].(map[string]interface{}); !ok {
			continue
		}

		n, err := utils.FormatTimeStamps(timestamp)
		if err != nil {
			moduleLogger.Warnf("timestamp format failed, %s", err.Error())
			return nil, err
		}

		pt, err := influxdb.NewPoint(
			namePrefix,
			tags,
			fields,
			time.Unix(n, 0),
		)

		if err != nil {
			moduleLogger.Errorf("%s", err.Error())
			return nil, err
		}

		pts = append(pts, pt)
	}

	return pts, nil
}

func ParsePoints(data []byte, prec string) ([]*influxdb.Point, error) {
	points, err := influxm.ParsePointsWithPrecision(data, time.Now().UTC(), prec)
	if err != nil {
		moduleLogger.Errorf("%s", err.Error())
		return nil, err
	}

	pts := []*influxdb.Point{}
	for _, pt := range points {
		measurement := string(pt.Name())
		tags := map[string]string{}

		for _, tag := range pt.Tags() {
			tags[string(tag.Key)] = string(tag.Value)
		}

		fields, _ := pt.Fields()
		pt, err := influxdb.NewPoint(
			measurement,
			tags,
			fields,
			pt.Time(),
		)

		if err != nil {
			moduleLogger.Errorf("assembly influx failed,  %s", err.Error())
			return nil, err
		}

		pts = append(pts, pt)
	}

	return pts, nil
}
