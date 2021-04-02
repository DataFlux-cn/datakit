package nginx

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"net"
	"net/url"
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
		if addr.Scheme == "http" {
			port = "80"
		} else if addr.Scheme == "https" {
			port = "443"
		} else {
			port = ""
		}
	}
	tags["nginx_server"] = host
	tags["nginx_port"] = port
	return tags
}

func newCountFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.UnknownUnit,
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

func newOtherFieldInfo(datatype, Type, unit, desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: datatype,
		Type:     Type,
		Unit:     unit,
		Desc:     desc,
	}
}
