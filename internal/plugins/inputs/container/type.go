// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"encoding/json"
	"strings"
)

type tagsType map[string]string

func (tags tagsType) append(m map[string]string) {
	for k, v := range m {
		tags[k] = v
	}
}

func (tags tagsType) addValueIfNotEmpty(key, value string) {
	if value == "" {
		return
	}
	tags[key] = value
}

type fieldsType map[string]interface{}

func (fields fieldsType) addMapWithJSON(key string, value map[string]string) { //nolint:unparam
	if len(value) == 0 {
		// 如果该map为空，则对应值为空字符串，否则在json序列化为"null"
		fields[key] = ""
		return
	}

	b, err := json.Marshal(value)
	if err != nil {
		return
	}
	fields[key] = string(b)
}

func (fields fieldsType) addSlice(key string, value []string) {
	fields[key] = strings.Join(value, ",")
}

const maxMessageLength = 256 * 1024 // 256KB

func (fields fieldsType) mergeToMessage(tags map[string]string) {
	temp := make(map[string]interface{})
	for k, v := range tags {
		temp[k] = v
	}
	for k, v := range fields {
		temp[k] = v
	}
	b, err := json.Marshal(temp)
	if err != nil {
		return
	}
	// limit length
	if len(b) > maxMessageLength {
		b = b[:maxMessageLength]
	}
	fields["message"] = string(b)
}

func (fields fieldsType) delete(key string) { //nolint:unparam
	delete(fields, key)
}

func (fields fieldsType) addLabel(labels map[string]string) {
	// empty array
	labelsString := "[]"
	if len(labels) != 0 {
		var lb []string
		for k, v := range labels {
			lb = append(lb, k+":"+v)
		}

		b, err := json.Marshal(lb)
		if err == nil {
			labelsString = string(b)
		}
	}
	// http://gitlab.jiagouyun.com/cloudcare-tools/kodo/-/issues/61#note_11580
	fields["df_label"] = labelsString
	fields["df_label_permission"] = "read_only"
	fields["df_label_source"] = "datakit"
}
