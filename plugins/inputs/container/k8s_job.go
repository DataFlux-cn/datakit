// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	v1 "k8s.io/api/batch/v1"
	"sigs.k8s.io/yaml"
)

var (
	_ k8sResourceMetricInterface = (*job)(nil)
	_ k8sResourceObjectInterface = (*job)(nil)
)

type job struct {
	client    k8sClientX
	extraTags map[string]string
	items     []v1.Job
}

func newJob(client k8sClientX, extraTags map[string]string) *job {
	return &job{
		client:    client,
		extraTags: extraTags,
	}
}

func (j *job) name() string {
	return "job"
}

func (j *job) pullItems() error {
	list, err := j.client.getJobs().List(context.Background(), metaV1ListOption)
	if err != nil {
		return fmt.Errorf("failed to get jobs resource: %w", err)
	}
	j.items = list.Items
	return nil
}

func (j *job) metric(election bool) (inputsMeas, error) {
	if err := j.pullItems(); err != nil {
		return nil, err
	}
	var res inputsMeas

	for _, item := range j.items {
		met := &jobMetric{
			tags: map[string]string{
				"job":       item.Name,
				"namespace": defaultNamespace(item.Namespace),
			},
			fields: map[string]interface{}{
				"failed":               item.Status.Failed,
				"succeeded":            item.Status.Succeeded,
				"completion_succeeded": 0,
				"completion_failed":    0,
				// "active":item.Status.Active,
			},
			election: election,
		}

		var succeeded, failed int
		for _, condition := range item.Status.Conditions {
			switch condition.Type {
			case v1.JobFailed:
				failed++
			case v1.JobComplete:
				succeeded++
			case v1.AlphaNoCompatGuaranteeJobFailureTarget:
				// nil
			case v1.JobSuspended:
				// nil
			}
		}

		met.fields["completion_succeeded"] = succeeded
		met.fields["completion_failed"] = failed

		met.tags.append(j.extraTags)
		res = append(res, met)
	}

	count, _ := j.count()
	for ns, c := range count {
		met := &jobMetric{
			tags:     map[string]string{"namespace": ns},
			fields:   map[string]interface{}{"count": c},
			election: election,
		}
		met.tags.append(j.extraTags)
		res = append(res, met)
	}

	return res, nil
}

func (j *job) object(election bool) (inputsMeas, error) {
	if err := j.pullItems(); err != nil {
		return nil, err
	}
	var res inputsMeas

	for _, item := range j.items {
		obj := &jobObject{
			tags: map[string]string{
				"name":     fmt.Sprintf("%v", item.UID),
				"job_name": item.Name,

				"namespace": defaultNamespace(item.Namespace),
			},
			fields: map[string]interface{}{
				"age":             int64(time.Since(item.CreationTimestamp.Time).Seconds()),
				"active":          item.Status.Active,
				"succeeded":       item.Status.Succeeded,
				"failed":          item.Status.Failed,
				"parallelism":     0,
				"completions":     0,
				"active_deadline": 0,
				"backoff_limit":   0,
			},
			election: election,
		}

		// 因为原数据类型（例如 item.Spec.Parallelism）就是 int32，所以此处也用 int32
		if item.Spec.Parallelism != nil {
			obj.fields["parallelism"] = *item.Spec.Parallelism
		}
		if item.Spec.Completions != nil {
			obj.fields["completions"] = *item.Spec.Completions
		}
		if item.Spec.ActiveDeadlineSeconds != nil {
			obj.fields["active_deadline"] = *item.Spec.ActiveDeadlineSeconds
		}
		if item.Spec.BackoffLimit != nil {
			obj.fields["backoff_limit"] = *item.Spec.BackoffLimit
		}

		if y, err := yaml.Marshal(item); err != nil {
			l.Debugf("failed to get job yaml %s, namespace %s, name %s, ignored", err.Error(), item.Namespace, item.Name)
		} else {
			obj.fields["yaml"] = string(y)
		}

		obj.tags.append(j.extraTags)

		obj.fields.addMapWithJSON("annotations", item.Annotations)
		obj.fields.addLabel(item.Labels)
		obj.fields.mergeToMessage(obj.tags)
		obj.fields.delete("annotations")
		obj.fields.delete("yaml")

		res = append(res, obj)
	}

	return res, nil
}

func (j *job) count() (map[string]int, error) {
	if err := j.pullItems(); err != nil {
		return nil, err
	}

	m := make(map[string]int)
	for _, item := range j.items {
		m[defaultNamespace(item.Namespace)]++
	}

	if len(m) == 0 {
		m["default"] = 0
	}

	return m, nil
}

type jobMetric struct {
	tags     tagsType
	fields   fieldsType
	election bool
}

func (j *jobMetric) LineProto() (*point.Point, error) {
	return point.NewPoint("kube_job", j.tags, j.fields, point.MOptElectionV2(j.election))
}

//nolint:lll
func (*jobMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kube_job",
		Desc: "Kubernetes Job 指标数据",
		Type: "metric",
		Tags: map[string]interface{}{
			"job":       inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace": inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
		},
		Fields: map[string]interface{}{
			// "active":               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of actively running pods."},
			"count":                &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Number of jobs"},
			"failed":               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of pods which reached phase Failed."},
			"succeeded":            &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of pods which reached phase Succeeded."},
			"completion_succeeded": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The job has completed its execution."},
			"completion_failed":    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The job has failed its execution."},
		},
	}
}

type jobObject struct {
	tags     tagsType
	fields   fieldsType
	election bool
}

func (j *jobObject) LineProto() (*point.Point, error) {
	return point.NewPoint("kubernetes_jobs", j.tags, j.fields, point.OOptElectionV2(j.election))
}

//nolint:lll
func (*jobObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kubernetes_jobs",
		Desc: "Kubernetes Job 对象数据",
		Type: "object",
		Tags: map[string]interface{}{
			"name":      inputs.NewTagInfo("UID"),
			"job_name":  inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace": inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
		},
		Fields: map[string]interface{}{
			"age":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "age (seconds)"},
			"active":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of actively running pods."},
			"succeeded":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of pods which reached phase Succeeded."},
			"failed":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of pods which reached phase Failed."},
			"completions":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Specifies the desired number of successfully finished pods the job should be run with."},
			"parallelism":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Specifies the maximum desired number of pods the job should run at any given time."},
			"backoff_limit":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Specifies the number of retries before marking this job failed."},
			"active_deadline": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "Specifies the duration in seconds relative to the startTime that the job may be active before the system tries to terminate it"},
			"message":         &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "object details"},
		},
	}
}

//nolint:gochecknoinits
func init() {
	registerK8sResourceMetric(func(c k8sClientX, m map[string]string) k8sResourceMetricInterface {
		return newJob(c, m)
	})
	registerK8sResourceObject(func(c k8sClientX, m map[string]string) k8sResourceObjectInterface {
		return newJob(c, m)
	})
	registerMeasurement(&jobMetric{})
	registerMeasurement(&jobObject{})
}
