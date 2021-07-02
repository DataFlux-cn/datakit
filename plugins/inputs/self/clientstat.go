package self

import (
	"runtime"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

var (
	StartTime time.Time
)

type ClientStat struct {
	HostName string

	PID    int
	Uptime int64
	OS     string
	Arch   string

	NumGoroutines int64
	HeapAlloc     int64
	HeapSys       int64
	HeapObjects   int64

	MinNumGoroutines int64
	MinHeapAlloc     int64
	MinHeapSys       int64
	MinHeapObjects   int64

	MaxNumGoroutines int64
	MaxHeapAlloc     int64
	MaxHeapSys       int64
	MaxHeapObjects   int64
}

func setMax(prev, cur int64) int64 {
	if prev == 0 || prev < cur {
		return cur
	} else {
		return prev
	}
}

func setMin(prev, cur int64) int64 {
	if prev == 0 || prev > cur {
		return cur
	} else {
		return prev
	}
}

func (s *ClientStat) Update() {
	s.HostName = config.Cfg.Hostname

	var memStatus runtime.MemStats
	runtime.ReadMemStats(&memStatus)

	s.NumGoroutines = int64(runtime.NumGoroutine())
	s.HeapAlloc = int64(memStatus.HeapAlloc)
	s.HeapSys = int64(memStatus.HeapSys)
	s.HeapObjects = int64(memStatus.HeapObjects)

	s.MaxNumGoroutines = setMax(s.MaxNumGoroutines, s.NumGoroutines)
	s.MinNumGoroutines = setMin(s.MinNumGoroutines, s.NumGoroutines)

	s.MaxHeapAlloc = setMax(s.MaxHeapAlloc, s.HeapAlloc)
	s.MinHeapAlloc = setMin(s.MinHeapAlloc, s.HeapAlloc)

	s.MaxHeapSys = setMax(s.MaxHeapSys, s.HeapSys)
	s.MinHeapSys = setMin(s.MinHeapSys, s.HeapSys)

	s.MaxHeapObjects = setMax(s.MaxHeapObjects, s.HeapObjects)
	s.MinHeapObjects = setMin(s.MinHeapObjects, s.HeapObjects)
}

func (s *ClientStat) ToMetric() *io.Point {

	s.Uptime = int64(time.Now().Sub(StartTime) / time.Second)

	measurement := "datakit"

	tags := map[string]string{
		"uuid":    config.Cfg.UUID,
		"vserion": git.Version,
		"os":      s.OS,
		"arch":    s.Arch,
		"host":    s.HostName,
	}

	fields := map[string]interface{}{
		"pid":    s.PID,
		"uptime": s.Uptime,

		"num_goroutines": s.NumGoroutines,
		"heap_alloc":     s.HeapAlloc,
		"heap_sys":       s.HeapSys,
		"heap_objects":   s.HeapObjects,

		"min_num_goroutines": s.MinNumGoroutines,
		"min_heap_alloc":     s.MinHeapAlloc,
		"min_heap_sys":       s.MinHeapSys,
		"min_heap_objects":   s.MinHeapObjects,

		"max_num_goroutines": s.MaxNumGoroutines,
		"max_heap_alloc":     s.MaxHeapAlloc,
		"max_heap_sys":       s.MaxHeapSys,
		"max_heap_objects":   s.MaxHeapObjects,
	}

	pt, err := io.MakePoint(measurement, tags, fields)
	if err != nil {
		l.Error(err)
	}

	return pt
}
