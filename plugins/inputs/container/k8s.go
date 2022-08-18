// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"fmt"
	"os"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

//nolint:deadcode
// const k8sBearerToken = "/run/secrets/k8s.io/serviceaccount/token"

type kubernetesInput struct {
	client *k8sClient
	cfg    *kubernetesInputConfig
}

type kubernetesInputConfig struct {
	url               string
	bearerToken       string
	bearerTokenString string
	extraTags         map[string]string

	enablePodMetric bool
	enableK8sMetric bool
	election        bool
}

func newKubernetesInput(cfg *kubernetesInputConfig) (*kubernetesInput, error) {
	k := &kubernetesInput{cfg: cfg}
	var err error

	//nolint:gocritic
	if cfg.bearerTokenString != "" {
		k.client, err = newK8sClientFromBearerTokenString(cfg.url, cfg.bearerTokenString)
	} else if cfg.bearerToken != "" {
		k.client, err = newK8sClientFromBearerToken(cfg.url, cfg.bearerToken)
	} else {
		err = fmt.Errorf("invalid bearerToken or bearerTokenString, cannot be empty")
	}
	if err != nil {
		return nil, err
	}
	return k, nil
}

type inputsMeas []inputs.Measurement

func (k *kubernetesInput) gatherResourceMetric() (inputsMeas, error) {
	var (
		res     inputsMeas
		counts  = make(map[string]map[string]int)
		lastErr error
	)

	for _, fn := range k8sResourceMetricList {
		x := fn(k.client, k.cfg.extraTags)

		if m, err := x.metric(k.cfg.election); err != nil {
			lastErr = err
		} else {
			switch x.name() {
			case "pod":
				if k.cfg.enablePodMetric {
					res = append(res, m...)
				}
			default:
				if k.cfg.enableK8sMetric {
					res = append(res, m...)
				}
			}
		}

		nsCount, err := x.count()
		if err != nil {
			lastErr = err
			continue
		}
		for ns, count := range nsCount {
			if c := counts[ns]; c == nil {
				counts[ns] = make(map[string]int)
			}
			counts[ns][x.name()] = count
		}
	}

	for ns, ct := range counts {
		count := &count{
			tags:     map[string]string{"namespace": ns},
			fields:   map[string]interface{}{},
			election: k.cfg.election,
		}

		for name, n := range ct {
			count.fields[name] = n
		}
		res = append(res, count)
	}

	return res, lastErr
}

func (k *kubernetesInput) gatherResourceObject() (inputsMeas, error) {
	var (
		res     inputsMeas
		lastErr error
	)

	for _, fn := range k8sResourceObjectList {
		x := fn(k.client, k.cfg.extraTags)
		if m, err := x.object(k.cfg.election); err == nil {
			res = append(res, m...)
		} else {
			lastErr = err
		}
	}

	return res, lastErr
}

func (k *kubernetesInput) watchingEventLog(stop <-chan interface{}) {
	watchingEvent(k.client, k.cfg.extraTags, stop, k.cfg.election)
}

type k8sResourceMetricInterface interface {
	name() string
	metric(election bool) (inputsMeas, error)
	count() (map[string]int, error)
}

type k8sResourceObjectInterface interface {
	name() string
	object(election bool) (inputsMeas, error)
}

type count struct {
	tags     tagsType
	fields   fieldsType
	election bool
}

func (c *count) LineProto() (*point.Point, error) {
	return point.NewPoint("kubernetes", c.tags, c.fields, point.MOptElectionV2(c.election))
}

//nolint:lll
func (*count) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kubernetes",
		Desc: "Kubernetes count 指标数据",
		Type: "metric",
		Tags: map[string]interface{}{
			"namespace": &inputs.TagInfo{Desc: "namespace"},
		},
		Fields: map[string]interface{}{
			"cluster_role": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "RBAC cluster role count"},
			"deployment":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "deployment count"},
			"node":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "node count"},
			"pod":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "pod count"},
			"cronjob":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "cronjob count"},
			"job":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "job count"},
			"service":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "service count"},
			"replica_set":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "replica_set count"},
		},
	}
}

func defaultNamespace(ns string) string {
	if ns == "" {
		return "default"
	}
	return ns
}

func defaultClusterName(name string) string {
	if name != "" {
		return name
	}
	if e := os.Getenv("ENV_K8S_CLUSTER_NAME"); e != "" {
		return e
	}
	return "kubernetes"
}

type (
	newK8sResourceMetricHandle func(k8sClientX, map[string]string) k8sResourceMetricInterface
	newK8sResourceObjectHandle func(k8sClientX, map[string]string) k8sResourceObjectInterface
)

var (
	k8sResourceMetricList []newK8sResourceMetricHandle
	k8sResourceObjectList []newK8sResourceObjectHandle
)

func registerK8sResourceMetric(newfn newK8sResourceMetricHandle) {
	k8sResourceMetricList = append(k8sResourceMetricList, newfn)
}

func registerK8sResourceObject(newfn newK8sResourceObjectHandle) {
	k8sResourceObjectList = append(k8sResourceObjectList, newfn)
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&count{})
}
