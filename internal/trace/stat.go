// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package trace

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

const (
	tracingStatName = "tracing_stat"
)

var g = goroutine.NewGroup(goroutine.Option{Name: "internal_trace"})

type TracingInfo struct {
	Service      string
	Resource     string
	Source       string
	Project      string
	Version      string
	RequestCount int
	ErrCount     int
	DurationAvg  int64
	key          string
	reCalc       bool
}

var ErrSendSpanInfoFailed = errors.New("send span information failed")

var (
	statOnce        = sync.Once{}
	statUnit        map[string]*TracingInfo
	tracingInfoChan chan *TracingInfo
	calcInterval                  = 30 * time.Second
	sendTimeout     time.Duration = time.Second
	retry           int           = 3
	isWorkerReady   bool          = false
)

func StartTracingStatistic() {
	statOnce.Do(func() {
		statUnit = make(map[string]*TracingInfo)
		tracingInfoChan = make(chan *TracingInfo, 100)
		startTracingStatWorker(calcInterval)
		isWorkerReady = true
	})
}

func startTracingStatWorker(interval time.Duration) {
	log.Info("tracing statistic worker started")

	g.Go(func(ctx context.Context) error {
		tick := time.NewTicker(interval)
		defer tick.Stop()
		for range tick.C {
			sendTracingInfo(&TracingInfo{
				key:    "recalc",
				reCalc: true,
			})
		}

		return nil
	})

	g.Go(func(ctx context.Context) error {
		for tinfo := range tracingInfoChan {
			if tinfo.reCalc {
				if len(statUnit) == 0 {
					continue
				}

				pts := makeTracingInfoPoint(statUnit)
				if len(pts) == 0 {
					log.Warn("empty tracing stat unit")
				} else if err := dkioFeed(tracingStatName, datakit.Tracing, pts, nil); err != nil {
					log.Error(err.Error())
				}
				statUnit = make(map[string]*TracingInfo)
			} else {
				if tunit, ok := statUnit[tinfo.key]; !ok {
					statUnit[tinfo.key] = tinfo
				} else {
					tunit.RequestCount += tinfo.RequestCount
					tunit.ErrCount += tinfo.ErrCount
					tunit.DurationAvg += tinfo.DurationAvg
				}
			}
		}
		return nil
	})
}

func StatTracingInfo(dktrace DatakitTrace) {
	if !isWorkerReady || len(dktrace) == 0 {
		return
	}

	tracingStatUnit := make(map[string]*TracingInfo)
	for i := range dktrace {
		var (
			key   = fmt.Sprintf("%s-%s", dktrace[i].Service, dktrace[i].Resource)
			tinfo *TracingInfo
			ok    bool
		)
		if tinfo, ok = tracingStatUnit[key]; !ok {
			tinfo = &TracingInfo{
				Source:   dktrace[i].Source,
				Project:  dktrace[i].Tags[TAG_PROJECT],
				Version:  dktrace[i].Tags[TAG_VERSION],
				Service:  dktrace[i].Service,
				Resource: dktrace[i].Resource,
				key:      key,
			}
			tracingStatUnit[key] = tinfo
		}
		if dktrace[i].SpanType == SPAN_TYPE_ENTRY {
			tinfo.RequestCount++
			if dktrace[i].Status == STATUS_ERR {
				tinfo.ErrCount++
			}
		}
		tinfo.DurationAvg += dktrace[i].Duration
	}

	for _, info := range tracingStatUnit {
		sendTracingInfo(info)
	}
}

func sendTracingInfo(tinfo *TracingInfo) {
	if tinfo == nil || tinfo.key == "" {
		return
	}

	timeout := time.NewTimer(sendTimeout)
	for retry > 0 {
		select {
		case <-timeout.C:
			retry--
			timeout.Reset(sendTimeout)
		case tracingInfoChan <- tinfo:
			return
		}
	}
}

func makeTracingInfoPoint(tinfos map[string]*TracingInfo) []*point.Point {
	var pts []*point.Point
	for _, tinfo := range tinfos {
		var (
			tags   = make(map[string]string)
			fields = make(map[string]interface{})
		)
		tags["source"] = tinfo.Source
		tags["project"] = tinfo.Project
		tags["version"] = tinfo.Version
		tags["service"] = tinfo.Service
		tags["resource"] = tinfo.Resource

		fields["request_count"] = tinfo.RequestCount
		fields["err_count"] = tinfo.ErrCount
		if tinfo.RequestCount == 0 {
			fields["duration_avg"] = tinfo.DurationAvg
		} else {
			fields["duration_avg"] = tinfo.DurationAvg / int64(tinfo.RequestCount)
		}

		if pt, err := point.NewPoint(tracingStatName, tags, fields, &point.PointOption{
			Time:     time.Now(),
			Category: datakit.Tracing,
			Strict:   false,
		}); err != nil {
			log.Error(err.Error())
		} else {
			pts = append(pts, pt)
		}
	}

	return pts
}
