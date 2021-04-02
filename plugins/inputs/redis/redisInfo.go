package redis

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type infoMeasurement struct {
	client  *redis.Client
	name    string
	tags    map[string]string
	fields  map[string]interface{}
	ts      time.Time
	resData map[string]interface{}
}

// 生成行协议
func (m *infoMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

// 指定指标
func (m *infoMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "redis_info",
		Fields: map[string]*inputs.FieldInfo{
			"info_latency_ms": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "The latency of the redis INFO command.",
			},
			"active_defrag_running": &inputs.FieldInfo{
				DataType: inputs.Bool,
				Type:     inputs.Gauge,
				Desc:     "Flag indicating if active defragmentation is active",
			},
			"redis_version": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Desc:     "Version of the Redis server",
			},
			"active_defrag_hits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "Number of value reallocations performed by active the defragmentation process",
			},
			"active_defrag_misses": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "Number of aborted value reallocations started by the active defragmentation process",
			},
			"active_defrag_key_hits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "Number of keys that were actively defragmented",
			},
			"active_defrag_key_misses": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "Number of keys that were skipped by the active defragmentation process",
			},
			"aof_last_rewrite_time_sec": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "Duration of the last AOF rewrite operation in seconds",
			},
			"aof_rewrite_in_progress": &inputs.FieldInfo{
				DataType: inputs.Bool,
				Type:     inputs.Gauge,
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
				Desc:     " Number of client connections (excluding connections from replicas)",
			},
			"connected_slaves": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "and cluster_connections.",
			},
			"rejected_connections": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "Number of connections rejected because of maxclients limit",
			},
			"blocked_clients": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "Number of clients pending on a blocking call (BLPOP, BRPOP, BRPOPLPUSH, BLMOVE, BZPOPMIN, BZPOPMAX)",
			},
			"client_biggest_input_buf": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "Biggest input buffer among current client connections",
			},
			"client_longest_output_list": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "Longest output list among current client connections",
			},
			"evicted_keys": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "Number of evicted keys due to maxmemory limit",
			},
			"expired_keys": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
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
				Desc:     "Global number of pub/sub channels with client subscriptions",
			},
			"pubsub_patterns": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "Global number of pub/sub pattern with client subscriptions",
			},
			"rdb_bgsave_in_progress": &inputs.FieldInfo{
				DataType: inputs.Bool,
				Type:     inputs.Gauge,
				Desc:     "Flag indicating a RDB save is on-going",
			},
			"rdb_changes_since_last_save": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "refers to the number of operations that produced some kind of changes in the dataset since the last time either SAVE or BGSAVE was called.",
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
				Desc: `The sum in bytes of all overheads that the server allocated for managing its internal data structures
                       used_memory_startup: Initial amount of memory consumed by Redis at startup in bytes`,
			},
			"maxmemory": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "The value of the maxmemory configuration directive",
			},
			"master_last_io_seconds_ago": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "",
			},
			"master_sync_in_progress": &inputs.FieldInfo{
				DataType: inputs.Bool,
				Type:     inputs.Gauge,
				Desc:     "ndicate the master is syncing to the replica",
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
				Desc:     "The server's current replication offset",
			},
			"slave_repl_offset": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "The replication offset of the replica instance",
			},
			"used_cpu_sys": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Rate,
				Desc:     "System CPU consumed by the Redis server, which is the sum of system CPU consumed by all threads of the server process (main thread and background threads)",
			},
			"used_cpu_sys_children": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Rate,
				Desc:     "System CPU consumed by the background processes",
			},
			"used_cpu_user": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Rate,
				Desc:     "User CPU consumed by the Redis server, which is the sum of user CPU consumed by all threads of the server process (main thread and background threads)",
			},
			"used_cpu_user_children": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Rate,
				Desc:     "User CPU consumed by the background processes",
			},
			"keyspace_hits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "Number of successful lookup of keys in the main dictionary",
			},
			"keyspace_misses": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Desc:     "Number of failed lookup of keys in the main dictionary",
			},
		},
	}
}

func CollectInfoMeasurement(cli *redis.Client, tags map[string]string) *infoMeasurement {
	m := &infoMeasurement{
		client:  cli,
		resData: make(map[string]interface{}),
		tags:    make(map[string]string),
		fields:  make(map[string]interface{}),
	}

	m.name = "redis_info"
	m.tags = tags

	m.getData()
	m.submit()

	fmt.Println("redis info", m)

	return m
}

// 数据源获取数据
func (m *infoMeasurement) getData() error {
	start := time.Now()
	info, err := m.client.Info("ALL").Result()
	if err != nil {
		return err
	}
	elapsed := time.Since(start)

	latencyMs := Round(float64(elapsed)/float64(time.Millisecond), 2)

	m.resData["info_latency_ms"] = latencyMs
	m.parseInfoData(info)

	return nil
}

// 解析返回结果
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

		m.resData[key] = val
	}

	return nil
}

// 提交数据
func (m *infoMeasurement) submit() error {
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
