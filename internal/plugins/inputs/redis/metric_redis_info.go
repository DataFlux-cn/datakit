// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"bufio"
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type infoMeasurement struct {
	cli         *redis.Client
	name        string
	tags        map[string]string
	fields      map[string]interface{}
	resData     map[string]interface{}
	election    bool
	lastCollect *redisCPUUsage
}

// 生成行协议.
func (m *infoMeasurement) LineProto() (*dkpt.Point, error) {
	return dkpt.NewPoint(m.name, m.tags, m.fields, dkpt.MOptElectionV2(m.election))
}

// 指定指标.
//nolint:lll,funlen
func (m *infoMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: redisInfoM,
		Type: "metric",
		Fields: map[string]interface{}{
			"info_latency_ms": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "The latency of the redis INFO command.",
			},
			"active_defrag_running": &inputs.FieldInfo{
				DataType: inputs.Bool,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Flag indicating if active defragmentation is active",
			},
			"active_defrag_hits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of value reallocations performed by active the defragmentation process",
			},
			"active_defrag_misses": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of aborted value reallocations started by the active defragmentation process",
			},
			"active_defrag_key_hits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of keys that were actively defragmented",
			},
			"active_defrag_key_misses": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of keys that were skipped by the active defragmentation process",
			},
			"aof_last_rewrite_time_sec": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Duration of the last AOF rewrite operation in seconds",
			},
			"aof_rewrite_in_progress": &inputs.FieldInfo{
				DataType: inputs.Bool,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Flag indicating a AOF rewrite operation is on-going",
			},
			"aof_current_size": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "AOF current file size",
			},
			"aof_buffer_length": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "Size of the AOF buffer",
			},
			"loading_total_bytes": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "Total file size",
			},
			"loading_loaded_bytes": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "Number of bytes already loaded",
			},
			"loading_loaded_perc": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.Percent,
				Desc:     "Same value expressed as a percentage",
			},
			"loading_eta_seconds": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationSecond,
				Desc:     "ETA in seconds for the load to be complete",
			},
			"connected_clients": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     " Number of client connections (excluding connections from replicas)",
			},
			"connected_slaves": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of connected replicas",
			},
			"rejected_connections": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of connections rejected because of Max-Clients limit",
			},
			"blocked_clients": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of clients pending on a blocking call (`BLPOP/BRPOP/BRPOPLPUSH/BLMOVE/BZPOPMIN/BZPOPMAX`)",
			},
			"client_biggest_input_buf": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "Biggest input buffer among current client connections",
			},
			"client_longest_output_list": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Longest output list among current client connections",
			},
			"evicted_keys": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of evicted keys due to Max-Memory limit",
			},
			"expired_keys": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Total number of key expiration events",
			},
			"latest_fork_usec": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "Duration of the latest fork operation in microseconds",
			},
			"pubsub_channels": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Global number of pub/sub channels with client subscriptions",
			},
			"pubsub_patterns": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Global number of pub/sub pattern with client subscriptions",
			},
			"rdb_bgsave_in_progress": &inputs.FieldInfo{
				DataType: inputs.Bool,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "Flag indicating a RDB save is on-going",
			},
			"rdb_changes_since_last_save": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Refers to the number of operations that produced some kind of changes in the dataset since the last time either `SAVE` or `BGSAVE` was called.",
			},
			"rdb_last_bgsave_time_sec": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationSecond,
				Desc:     "Duration of the last RDB save operation in seconds",
			},
			"mem_fragmentation_ratio": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.Percent,
				Desc:     "Ratio between used_memory_rss and used_memory",
			},
			"used_memory": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "Total number of bytes allocated by Redis using its allocator (either standard libc, jemalloc, or an alternative allocator such as tcmalloc)",
			},
			"used_memory_lua": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "Number of bytes used by the Lua engine",
			},
			"used_memory_peak": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "Peak memory consumed by Redis (in bytes)",
			},
			"used_memory_rss": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "Number of bytes that Redis allocated as seen by the operating system (a.k.a resident set size)",
			},
			"used_memory_startup": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "Initial amount of memory consumed by Redis at startup in bytes",
			},
			"used_memory_overhead": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "The sum in bytes of all overheads that the server allocated for managing its internal data structures",
			},
			"maxmemory": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "The value of the Max Memory configuration directive",
			},
			"master_last_io_seconds_ago": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationSecond,
				Desc:     "Number of seconds since the last interaction with master",
			},
			"master_sync_in_progress": &inputs.FieldInfo{
				DataType: inputs.Bool,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "Indicate the master is syncing to the replica",
			},
			"master_sync_left_bytes": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "Number of bytes left before syncing is complete (may be negative when master_sync_total_bytes is 0)",
			},
			"repl_backlog_histlen": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "Size in bytes of the data in the replication backlog buffer",
			},
			"master_repl_offset": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The server's current replication offset",
			},
			"slave_repl_offset": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The replication offset of the replica instance",
			},
			"used_cpu_sys": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.DurationSecond,
				Desc:     "System CPU consumed by the Redis server, which is the sum of system CPU consumed by all threads of the server process (main thread and background threads)",
			},
			"used_cpu_sys_percent": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Rate,
				Unit:     inputs.Percent,
				Desc:     "System CPU percentage consumed by the Redis server, which is the sum of system CPU consumed by all threads of the server process (main thread and background threads)",
			},
			"used_cpu_sys_children": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.DurationSecond,
				Desc:     "System CPU consumed by the background processes",
			},
			"used_cpu_user": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.DurationSecond,
				Desc:     "User CPU consumed by the Redis server, which is the sum of user CPU consumed by all threads of the server process (main thread and background threads)",
			},
			"used_cpu_user_percent": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Rate,
				Unit:     inputs.Percent,
				Desc:     "User CPU percentage consumed by the Redis server, which is the sum of user CPU consumed by all threads of the server process (main thread and background threads)",
			},
			"used_cpu_user_children": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Count,
				Unit:     inputs.DurationSecond,
				Desc:     "User CPU consumed by the background processes",
			},
			"keyspace_hits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of successful lookup of keys in the main dictionary",
			},
			"keyspace_misses": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of failed lookup of keys in the main dictionary",
			},

			"total_net_input_bytes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total number of bytes read from the network",
			},

			"total_net_output_bytes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total number of bytes written to the network",
			},
		},
		Tags: map[string]interface{}{
			"host":          &inputs.TagInfo{Desc: "Hostname"},
			"redis_version": &inputs.TagInfo{Desc: "Version of the Redis server"},
			"server":        &inputs.TagInfo{Desc: "Server addr"},
		},
	}
}

// 数据源获取数据.
func (m *infoMeasurement) getData() error {
	start := time.Now()
	ctx := context.Background()

	info, err := m.cli.Info(ctx, "ALL").Result()
	if err != nil {
		l.Error("redis exec command `All`, happen error,", err)
		return err
	}
	elapsed := time.Since(start)
	nextTS := start.Add(elapsed / 2)

	latencyMs := Round(float64(elapsed)/float64(time.Millisecond), 2)

	m.resData["info_latency_ms"] = latencyMs
	if err := m.parseInfoData(info); err != nil {
		l.Error("redis exec command `All` result data, parse error,", err)
		return err
	}

	// Calculate CPU usage.
	if usedCPUSys, has := m.resData["used_cpu_sys"]; has {
		if usedCPUSysStr, ok := usedCPUSys.(string); ok {
			usedCPUSysFloat, err := strconv.ParseFloat(usedCPUSysStr, 64)
			if err != nil {
				l.Errorf("failed to parse used_cpu_sys: %v", err)
			}
			totElapsed := nextTS.Sub(m.lastCollect.usedCPUSysTS)
			if !m.lastCollect.usedCPUSysTS.IsZero() {
				m.resData["used_cpu_sys_percent"] = (usedCPUSysFloat - m.lastCollect.usedCPUSys) / totElapsed.Seconds()
			}
			m.lastCollect.usedCPUSys = usedCPUSysFloat
			m.lastCollect.usedCPUSysTS = nextTS
		} else {
			l.Errorf("failed to cast used_cpu_sys to string, got type = %T", usedCPUSys)
		}
	}
	if usedCPUUser, has := m.resData["used_cpu_user"]; has {
		if usedCPUUserStr, ok := usedCPUUser.(string); ok {
			usedCPUUserFloat, err := strconv.ParseFloat(usedCPUUserStr, 64)
			if err != nil {
				l.Errorf("failed to parse used_cpu_user: %v", err)
			}
			totElapsed := nextTS.Sub(m.lastCollect.usedCPUUserTS)
			if !m.lastCollect.usedCPUUserTS.IsZero() {
				m.resData["used_cpu_user_percent"] = (usedCPUUserFloat - m.lastCollect.usedCPUUser) / totElapsed.Seconds()
			}
			m.lastCollect.usedCPUUser = usedCPUUserFloat
			m.lastCollect.usedCPUUserTS = nextTS
		} else {
			l.Errorf("failed to cast used_cpu_user to string, got type = %T", usedCPUUser)
		}
	}

	return nil
}

// 解析返回结果.
func (m *infoMeasurement) parseInfoData(info string) error {
	rdr := strings.NewReader(info)

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

		key := parts[0]
		val := strings.TrimSpace(parts[1])

		if key == "redis_version" {
			if val == "" {
				val = "unknown"
			}
			m.tags["redis_version"] = val
		} else {
			m.resData[key] = val
		}
	}

	return nil
}

// 提交数据.
func (m *infoMeasurement) submit() error {
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
