// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskio

import (
	"regexp"

	"github.com/shirou/gopsutil/disk"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type DiskIO func(names ...string) (map[string]disk.IOCountersStat, error)

type DevicesFilter struct {
	filters []*regexp.Regexp
}

func (f *DevicesFilter) Compile(exprs []string) error {
	f.filters = make([]*regexp.Regexp, 0) // clear
	for _, expr := range exprs {
		if filter, err := regexp.Compile(expr); err == nil {
			f.filters = append(f.filters, filter)
		} else {
			return err
		}
	}
	return nil
}

func (f *DevicesFilter) Match(s string) bool {
	for _, filter := range f.filters {
		if filter.MatchString(s) {
			return true
		}
	}
	return false
}

func newFieldsInfoMS(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.DurationMS,
		Desc:     desc,
	}
}

func newFieldsInfoBytes(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.SizeByte,
		Desc:     desc,
	}
}

func newFieldsInfoBytesPerSec(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.BytesPerSec,
		Desc:     desc,
	}
}

func newFieldsInfoCount(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}
