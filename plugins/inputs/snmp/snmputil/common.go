// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

// Package snmputil contains snmp utils.
package snmputil

import (
	"bytes"
	"fmt"
	"sort"
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

type MetricDatas struct {
	Data []*MetricData
}

func (md *MetricDatas) Add(name string, value float64, tags []string) {
	md.Data = append(md.Data, &MetricData{
		Name:  name,
		Value: value,
		Tags:  tags,
	})
}

type MetricData struct {
	Name     string
	Value    float64
	Tags     []string
	TagsHash string
}

//------------------------------------------------------------------------------

// CreateStringBatches batches strings into chunks with specific size.
func CreateStringBatches(elements []string, size int) ([][]string, error) {
	var batches [][]string

	if size <= 0 {
		return nil, fmt.Errorf("batch size must be positive. invalid size: %d", size)
	}

	for i := 0; i < len(elements); i += size {
		j := i + size
		if j > len(elements) {
			j = len(elements)
		}
		batch := elements[i:j]
		batches = append(batches, batch)
	}

	return batches, nil
}

//------------------------------------------------------------------------------
// Copy array

// CopyStrings makes a copy of a list of strings.
func CopyStrings(tags []string) []string {
	newTags := make([]string, len(tags))
	copy(newTags, tags)
	return newTags
}

func CopySymbolConfigs(scs []SymbolConfig) []SymbolConfig {
	newSCs := make([]SymbolConfig, len(scs))
	for k, v := range scs {
		newSCs[k] = v.Copy()
	}
	return newSCs
}

func CopyMetricIndexTransforms(mits []MetricIndexTransform) []MetricIndexTransform {
	newMITs := make([]MetricIndexTransform, len(mits))
	for k, v := range mits {
		newMITs[k] = v.Copy()
	}
	return newMITs
}

func CopyMetricsConfigs(mcs []MetricsConfig) []MetricsConfig {
	newMCs := make([]MetricsConfig, len(mcs))
	for k, v := range mcs {
		newMCs[k] = v.Copy()
	}
	return newMCs
}

func CopyMetricTagConfigs(mtcs []MetricTagConfig) []MetricTagConfig {
	newMTCs := make([]MetricTagConfig, len(mtcs))
	for k, v := range mtcs {
		newMTCs[k] = v.Copy()
	}
	return newMTCs
}

//------------------------------------------------------------------------------
// Copy map

func CopyMapStringString(in map[string]string) map[string]string {
	newMap := make(map[string]string, len(in))
	for k, v := range in {
		newMap[k] = v
	}
	return newMap
}

func CopyMapStringMetadataField(in map[string]MetadataField) map[string]MetadataField {
	newMap := make(map[string]MetadataField, len(in))
	for k, v := range in {
		newMap[k] = v.Copy()
	}
	return newMap
}

func CopyMapStringMetadataResourceConfig(in map[string]MetadataResourceConfig) map[string]MetadataResourceConfig {
	newMap := make(map[string]MetadataResourceConfig, len(in))
	for k, v := range in {
		newMap[k] = v.Copy()
	}
	return newMap
}

//------------------------------------------------------------------------------

// SortUniqInPlace sorts and remove duplicates from elements in place
// The returned slice is a subslice of elements.
func SortUniqInPlace(elements []string) []string {
	size := len(elements)
	if size < 2 {
		return elements
	}
	if size <= InsertionSortThreshold {
		InsertionSort(elements)
	} else {
		// this will trigger an alloc because sorts uses interface{} internaly
		// which confuses the escape analysis
		sort.Strings(elements)
	}
	return uniqSorted(elements)
}

// uniqSorted remove duplicate elements from the given slice
// the given slice needs to be sorted.
func uniqSorted(elements []string) []string {
	j := 0
	for i := 1; i < len(elements); i++ {
		if elements[j] == elements[i] {
			continue
		}
		j++
		elements[j] = elements[i]
	}
	return elements[:j+1]
}

// InsertionSortThreshold is the slice size after which we should consider
// using the stdlib sort method instead of the InsertionSort implemented below.
const InsertionSortThreshold = 40

// InsertionSort sorts in-place the given elements, not doing any allocation.
// It is very efficient for on slices but if memory allocation is not an issue,
// consider using the stdlib `sort.Sort` method on slices having a size > InsertionSortThreshold.
// See `pkg/util/sort_benchmarks_note.md` for more details.
func InsertionSort(elements []string) {
	for i := 1; i < len(elements); i++ {
		temp := elements[i]
		j := i
		for j > 0 && temp <= elements[j-1] {
			elements[j] = elements[j-1]
			j--
		}
		elements[j] = temp
	}
}

//------------------------------------------------------------------------------

// NormalizeNamespace applies policy according to hostname rule.
func NormalizeNamespace(namespace string) (string, error) {
	var buf bytes.Buffer

	// namespace longer than 100 characters are illegal
	if len(namespace) > 100 {
		return "", fmt.Errorf("namespace is too long, should contain less than 100 characters")
	}

	for _, r := range namespace {
		switch r {
		// has null rune just toss the whole thing
		case '\x00':
			return "", fmt.Errorf("namespace cannot contain null character")
		// drop these characters entirely
		case '\n', '\r', '\t':
			continue
		// replace characters that are generally used for xss with '-'
		case '>', '<':
			buf.WriteByte('-')
		default:
			buf.WriteRune(r)
		}
	}

	normalizedNamespace := buf.String()
	if normalizedNamespace == "" {
		return "", fmt.Errorf("namespace cannot be empty")
	}

	return normalizedNamespace, nil
}

//------------------------------------------------------------------------------

const packageName = "snmputil"

var (
	l       = logger.DefaultSLogger(packageName)
	runOnce sync.Once
)

func SetLog() {
	runOnce.Do(func() {
		l = logger.SLogger(packageName)
	})
}