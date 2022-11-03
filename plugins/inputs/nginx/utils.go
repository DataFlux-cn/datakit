// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package nginx

import (
	"net"
	"net/url"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func getTags(urlString string) map[string]string {
	tags := map[string]string{
		"server": "",
		"port":   "",
	}
	addr, err := url.Parse(urlString)
	if err != nil {
		return tags
	}

	h := addr.Host
	host, port, err := net.SplitHostPort(h)
	if err != nil {
		host = addr.Host
		switch addr.Scheme {
		case "http":
			port = "80"
		case "https":
			port = "443"
		default:
			port = ""
		}
	}
	tags["nginx_server"] = host
	tags["nginx_port"] = port
	if !strings.Contains(host, "127.0.0.1") && !strings.Contains(host, "localhost") {
		tags["host"] = host
	}
	tags["host"] = host
	return tags
}

func newCountFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}

func newByteFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     desc,
	}
}

func newOtherFieldInfo(datatype, ftype, unit, desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: datatype,
		Type:     ftype,
		Unit:     unit,
		Desc:     desc,
	}
}
