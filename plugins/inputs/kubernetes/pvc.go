package kubernetes

import (
	"context"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	corev1 "k8s.io/api/core/v1"
	"strings"
	"time"
)

var pvcMeasurement = "kube_pvc"

type pvcM struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *pvcM) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *pvcM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: pvcMeasurement,
		Desc: "kubernetes pvc 对象",
		Tags: map[string]interface{}{
			"pvc_name":     &inputs.TagInfo{Desc: "pvc name"},
			"namespace":    &inputs.TagInfo{Desc: "namespace"},
			"phase":        &inputs.TagInfo{Desc: "phase"},
			"storageclass": &inputs.TagInfo{Desc: "storage class"},
		},
		Fields: map[string]interface{}{
			"phase_type": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "phase type, bound:0, lost:1, pending:2, unknown:3",
			},
		},
	}
}

func (i *Input) collectPersistentVolumeClaims(ctx context.Context) error {
	list, err := i.client.getPersistentVolumeClaims(ctx)
	if err != nil {
		return err
	}
	for _, pvc := range list.Items {
		i.gatherPersistentVolumeClaim(pvc)
	}

	return err
}

func (i *Input) gatherPersistentVolumeClaim(pvc corev1.PersistentVolumeClaim) {
	phaseType := 3
	switch strings.ToLower(string(pvc.Status.Phase)) {
	case "bound":
		phaseType = 0
	case "lost":
		phaseType = 1
	case "pending":
		phaseType = 2
	}
	fields := map[string]interface{}{
		"phase_type": phaseType,
	}
	tags := map[string]string{
		"pvc_name":     pvc.Name,
		"namespace":    pvc.Namespace,
		"phase":        string(pvc.Status.Phase),
		"storageclass": *pvc.Spec.StorageClassName,
	}

	// for key, val := range pvc.Spec.Selector.MatchLabels {
	// 	if i.selectorFilter.Match(key) {
	// 		tags["selector_"+key] = val
	// 	}
	// }

	m := &pvcM{
		name:   pvcMeasurement,
		tags:   tags,
		fields: fields,
		ts:     time.Now(),
	}

	i.collectCache = append(i.collectCache, m)
}
