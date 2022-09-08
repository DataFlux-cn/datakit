// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const dockerContainerName = "docker_containers"

func gatherDockerContainerMetric(client dockerClientX, k8sClient k8sClientX, container *types.Container) (*containerMetric, error) {
	m := &containerMetric{}
	m.tags = getContainerInfo(container, k8sClient)

	f, err := getContainerStats(client, container.ID)
	if err != nil {
		return nil, err
	}
	m.fields = f

	return m, nil
}

func getContainerTags(container *types.Container) tagsType {
	imageName, imageShortName, imageTag := ParseImage(container.Image)

	tags := map[string]string{
		"state":                  container.State,
		"docker_image":           container.Image,
		"image":                  container.Image,
		"image_name":             imageName,
		"image_short_name":       imageShortName,
		"image_tag":              imageTag,
		"container_runtime_name": getContainerName(container.Names),
		"container_id":           container.ID,
		"linux_namespace":        "moby",
	}

	if n := getContainerNameForLabels(container.Labels); n != "" {
		tags["container_name"] = n
	} else {
		tags["container_name"] = tags["container_runtime_name"]
	}

	if !containerIsFromKubernetes(getContainerName(container.Names)) {
		tags["container_type"] = "docker"
	} else {
		tags["container_type"] = "kubernetes"
	}

	if n := getPodNameForLabels(container.Labels); n != "" {
		tags["pod_name"] = n
	}
	if n := getPodNamespaceForLabels(container.Labels); n != "" {
		tags["namespace"] = n
	}

	return tags
}

func getContainerInfo(container *types.Container, k8sClient k8sClientX) tagsType {
	tags := getContainerTags(container)

	podname := getPodNameForLabels(container.Labels)
	podnamespace := getPodNamespaceForLabels(container.Labels)
	podContainerName := getContainerNameForLabels(container.Labels)

	if k8sClient == nil || podname == "" {
		return tags
	}

	meta, err := queryPodMetaData(k8sClient, podname, podnamespace)
	if err != nil {
		// ignore err
		return tags
	}

	if image := meta.containerImage(podContainerName); image != "" {
		// 如果能找到 pod image，则使用它
		imageName, imageShortName, imageTag := ParseImage(image)

		tags["docker_image"] = image
		tags["image"] = image
		tags["image_name"] = imageName
		tags["image_short_name"] = imageShortName
		tags["image_tag"] = imageTag
	}

	if replicaSet := meta.replicaSet(); replicaSet != "" {
		tags["replicaSet"] = replicaSet
	}

	labels := meta.labels()
	if deployment := getDeployment(labels["app"], podnamespace); deployment != "" {
		tags["deployment"] = deployment
	}

	return tags
}

const streamStats = false

func getContainerStats(client dockerClientX, containerID string) (fieldsType, error) {
	resp, err := client.ContainerStats(context.TODO(), containerID, streamStats)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.OSType == datakit.OSWindows {
		return nil, fmt.Errorf("not support windows docker/container")
	}

	var v *types.StatsJSON
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}
	return calculateContainerStats(v), nil
}

func calculateContainerStats(v *types.StatsJSON) map[string]interface{} {
	mem := calculateMemUsageUnixNoCache(v.MemoryStats)
	memPercent := calculateMemPercentUnixNoCache(float64(v.MemoryStats.Limit), float64(mem))
	netRx, netTx := calculateNetwork(v.Networks)
	blkRead, blkWrite := calculateBlockIO(v.BlkioStats)

	return map[string]interface{}{
		"cpu_usage": calculateCPUPercentUnix(v.PreCPUStats.CPUUsage.TotalUsage,
			v.PreCPUStats.SystemUsage, v), /*float64*/
		"cpu_delta":          calculateCPUDelta(v),
		"cpu_system_delta":   calculateCPUSystemDelta(v),
		"cpu_numbers":        calculateCPUNumbers(v),
		"mem_limit":          int64(v.MemoryStats.Limit),
		"mem_usage":          mem,
		"mem_used_percent":   memPercent, /*float64*/
		"mem_failed_count":   int64(v.MemoryStats.Failcnt),
		"network_bytes_rcvd": netRx,
		"network_bytes_sent": netTx,
		"block_read_byte":    blkRead,
		"block_write_byte":   blkWrite,
	}
}

// getDeployment
// 	ex: deployment-func-work-0, namespace is ’func', deployment is 'work-0'
func getDeployment(appStr, namespace string) string {
	if !strings.HasPrefix(appStr, "deployment") {
		return ""
	}

	if !strings.HasPrefix(appStr, "deployment-"+namespace) {
		return ""
	}

	return strings.TrimPrefix(appStr, "deployment-"+namespace+"-")
}

func getContainerName(names []string) string {
	if len(names) > 0 {
		return strings.TrimPrefix(names[0], "/")
	}
	return "unknown"
}

func isRunningContainer(state string) bool {
	return state == "running"
}

func isPauseContainer(command string) bool {
	return command == "/pause"
}

// nolint:lll
// containerIsFromKubernetes 判断该容器是否由kubernetes创建
// 所有kubernetes启动的容器的containerNamePrefix都是k8s，依据链接如下
// https://github.com/rootsongjc/kubernetes-handbook/blob/master/practice/monitor.md#%E5%AE%B9%E5%99%A8%E7%9A%84%E5%91%BD%E5%90%8D%E8%A7%84%E5%88%99
func containerIsFromKubernetes(containerName string) bool {
	const kubernetesContainerNamePrefix = "k8s"
	return strings.HasPrefix(containerName, kubernetesContainerNamePrefix)
}

type containerMetric struct {
	tags   tagsType
	fields fieldsType
}

func (c *containerMetric) LineProto() (*point.Point, error) {
	return point.NewPoint(dockerContainerName, c.tags, c.fields, point.MOpt())
}

//nolint:lll
func (c *containerMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: dockerContainerName,
		Type: "metric",
		Desc: "容器指标数据，只采集正在运行的容器",
		Tags: map[string]interface{}{
			"container_id":           inputs.NewTagInfo(`容器 ID`),
			"container_name":         inputs.NewTagInfo(`k8s 命名的容器名（在 labels 中取 'io.kubernetes.container.name'），如果值为空则跟 container_runtime_name 相同`),
			"container_runtime_name": inputs.NewTagInfo(`由 runtime 命名的容器名（例如 docker ps 查看），如果值为空则默认是 unknown（[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)）`),
			"docker_image":           inputs.NewTagInfo("镜像全称，例如 `nginx.org/nginx:1.21.0` （Depercated, use image）"),
			"linux_namespace":        inputs.NewTagInfo(`该容器所在的 [linux namespace](https://man7.org/linux/man-pages/man7/namespaces.7.html)`),
			"image":                  inputs.NewTagInfo("镜像全称，例如 `nginx.org/nginx:1.21.0`"),
			"image_name":             inputs.NewTagInfo("镜像名称，例如 `nginx.org/nginx`"),
			"image_short_name":       inputs.NewTagInfo("镜像名称精简版，例如 `nginx`"),
			"image_tag":              inputs.NewTagInfo("镜像 tag，例如 `1.21.0`"),
			"container_type":         inputs.NewTagInfo(`容器类型，表明该容器由谁创建，kubernetes/docker/containerd`),
			"state":                  inputs.NewTagInfo(`运行状态，running（containerd 缺少此字段）`),
			"pod_name":               inputs.NewTagInfo(`pod 名称（容器由 k8s 创建时存在）`),
			"namespace":              inputs.NewTagInfo(`pod 的 k8s 命名空间（k8s 创建容器时，会打上一个形如 'io.kubernetes.pod.namespace' 的 label，DataKit 将其命名为 'namespace'）`),
			"deployment":             inputs.NewTagInfo(`deployment 名称（容器由 k8s 创建时存在，containerd 缺少此字段）`),
		},
		Fields: map[string]interface{}{
			"cpu_usage":          &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "CPU 占主机总量的使用率"},
			"cpu_delta":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationNS, Desc: "容器 CPU 增量（containerd 缺少此字段）"},
			"cpu_system_delta":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationNS, Desc: "系统 CPU 增量，仅支持 Linux（containerd 缺少此字段）"},
			"cpu_numbers":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "CPU 核心数（containerd 缺少此字段）"},
			"mem_limit":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "内存可用总量，如果未对容器做内存限制，则为主机内存容量"},
			"mem_usage":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "内存使用量"},
			"mem_used_percent":   &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "内存使用率，使用量除以可用总量"},
			"mem_failed_count":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "内存分配失败的次数（containerd 缺少此字段）"},
			"network_bytes_rcvd": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "从网络接收到的总字节数（containerd 缺少此字段）"},
			"network_bytes_sent": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "向网络发送出的总字节数（containerd 缺少此字段）"},
			"block_read_byte":    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "从容器文件系统读取的总字节数（containerd 缺少此字段）"},
			"block_write_byte":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "向容器文件系统写入的总字节数（containerd 缺少此字段）"},
		},
	}
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&containerMetric{})
}
