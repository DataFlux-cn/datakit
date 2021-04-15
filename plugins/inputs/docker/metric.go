package docker

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/docker/docker/api/types"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func (this *Input) gather(option ...*gatherOption) ([]*io.Point, error) {
	var opt *gatherOption
	if len(option) >= 1 {
		opt = option[0]
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, this.timeoutDuration)
	defer cancel()

	cList, err := this.client.ContainerList(ctx, this.opts)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	var pts []*io.Point

	for _, container := range cList {
		tags := this.gatherContainerInfo(container)

		// 区分指标和对象
		// 对象数据需要有 name 标签
		if opt != nil && opt.IsObjectCategory {
			tags["name"] = container.ID
		}

		fields, err := this.gatherStats(container)
		if err != nil {
			l.Error(err)
			continue
		}

		pt, err := io.MakePoint(dockerContainersName, tags, fields, time.Now())
		if err != nil {
			l.Error(err)
			continue
		}
		pts = append(pts, pt)
	}

	return pts, nil
}

func (this *Input) gatherContainerInfo(container types.Container) map[string]string {
	tags := map[string]string{
		"container_id":   container.ID,
		"container_name": getContainerName(container.Names),
		"docker_image":   container.ImageID,
		"image_name":     container.Image,
		"state":          container.State,
	}

	for k, v := range this.Tags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}

	podInfo, err := this.gatherK8sPodInfo(container.ID)
	if err != nil {
		l.Warnf("gather k8s pod error, %s", err)
	}

	for k, v := range podInfo {
		tags[k] = v
	}

	return tags
}

const streamStats = false

func (this *Input) gatherStats(container types.Container) (map[string]interface{}, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, this.timeoutDuration)
	defer cancel()

	resp, err := this.client.ContainerStats(ctx, container.ID, streamStats)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.OSType == "windows" {
		return nil, nil
	}

	var v *types.StatsJSON
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}

	mem := calculateMemUsageUnixNoCache(v.MemoryStats)
	memPercent := calculateMemPercentUnixNoCache(float64(v.MemoryStats.Limit), float64(mem))
	netRx, netTx := calculateNetwork(v.Networks)
	blkRead, blkWrite := calculateBlockIO(v.BlkioStats)

	return map[string]interface{}{
		"cpu_usage_percent":  calculateCPUPercentUnix(v.PreCPUStats.CPUUsage.TotalUsage, v.PreCPUStats.SystemUsage, v), /*float64*/
		"cpu_delta":          calculateCPUDelta(v),
		"cpu_system_delta":   calculateCPUSystemDelta(v),
		"cpu_numbers":        calculateCPUNumbers(v),
		"mem_available":      int64(v.MemoryStats.Limit),
		"mem_used":           mem,
		"mem_usage_percent":  memPercent, /*float64*/
		"mem_failed_count":   int64(v.MemoryStats.Failcnt),
		"network_bytes_rcvd": netRx,
		"network_bytes_sent": netTx,
		"block_read_byte":    blkRead,
		"block_write_byte":   blkWrite,
		"from_kubernetes":    contianerIsFromKubernetes(getContainerName(container.Names)),
	}, nil
}

func (this *Input) gatherK8sPodInfo(id string) (map[string]string, error) {
	if this.kubernetes == nil {
		return nil, nil
	}
	return this.kubernetes.GatherPodInfo(id)
}

func getContainerName(names []string) string {
	if len(names) > 0 {
		return strings.TrimPrefix(names[0], "/")
	}
	return ""
}

// contianerIsFromKubernetes 判断该容器是否由kubernetes创建
// 所有kubernetes启动的容器的containerNamePrefix都是k8s，依据链接如下
// https://github.com/rootsongjc/kubernetes-handbook/blob/master/practice/monitor.md#%E5%AE%B9%E5%99%A8%E7%9A%84%E5%91%BD%E5%90%8D%E8%A7%84%E5%88%99
func contianerIsFromKubernetes(containerName string) bool {
	const kubernetesContainerNamePrefix = "k8s"
	return strings.HasPrefix(containerName, kubernetesContainerNamePrefix)
}
