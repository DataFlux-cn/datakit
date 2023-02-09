// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tomcat

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (m *measurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOptElection())
}

type TomcatGlobalRequestProcessorM struct{ measurement }

type TomcatJspMonitorM struct{ measurement }

type TomcatThreadPoolM struct{ measurement }

type TomcatServletM struct{ measurement }

type TomcatCacheM struct{ measurement }

//nolint:lll
func (m *TomcatGlobalRequestProcessorM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "tomcat_global_request_processor",
		Fields: map[string]interface{}{
			"requestCount":   newFielInfoCount("Number of requests processed."),
			"bytesReceived":  newFielInfoCount("Amount of data received, in bytes."),
			"bytesSent":      newFielInfoCount("Amount of data sent, in bytes."),
			"processingTime": newFielInfoInt("Total time to process the requests."),
			"errorCount":     newFielInfoCount("Number of errors."),
		},
		Tags: map[string]interface{}{
			"name":              inputs.NewTagInfo("Protocol handler name."),
			"jolokia_agent_url": inputs.NewTagInfo("Jolokia agent url."),
			"host":              inputs.NewTagInfo("System hostname."),
		},
	}
}

//nolint:lll
func (m *TomcatJspMonitorM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "tomcat_jsp_monitor",
		Fields: map[string]interface{}{
			"jspCount":       newFielInfoCount("The number of JSPs that have been loaded into a webapp."),
			"jspReloadCount": newFielInfoCount("The number of JSPs that have been reloaded."),
			"jspUnloadCount": newFielInfoCount("The number of JSPs that have been unloaded."),
		},
		Tags: map[string]interface{}{
			"J2EEApplication":   inputs.NewTagInfo("J2EE Application."),
			"J2EEServer":        inputs.NewTagInfo("J2EE Servers."),
			"WebModule":         inputs.NewTagInfo("Web Module."),
			"jolokia_agent_url": inputs.NewTagInfo("Jolokia agent url."),
			"host":              inputs.NewTagInfo("System hostname."),
		},
	}
}

//nolint:lll
func (m *TomcatThreadPoolM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "tomcat_thread_pool",
		Fields: map[string]interface{}{
			"maxThreads":         newFielInfoCount("MaxThreads."),
			"currentThreadCount": newFielInfoCount("CurrentThreadCount."),
			"currentThreadsBusy": newFielInfoCount("CurrentThreadsBusy."),
		},
		Tags: map[string]interface{}{
			"name":              inputs.NewTagInfo("Protocol handler name."),
			"jolokia_agent_url": inputs.NewTagInfo("Jolokia agent url."),
			"host":              inputs.NewTagInfo("System hostname."),
		},
	}
}

//nolint:lll
func (m *TomcatServletM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "tomcat_servlet",
		Fields: map[string]interface{}{
			"processingTime": newFielInfoInt("Total execution time of the servlet's service method."),
			"errorCount":     newFielInfoCount("Error count."),
			"requestCount":   newFielInfoCount("Number of requests processed by this wrapper."),
		},
		Tags: map[string]interface{}{
			"J2EEApplication":   inputs.NewTagInfo("J2EE Application."),
			"J2EEServer":        inputs.NewTagInfo("J2EE Server."),
			"WebModule":         inputs.NewTagInfo("Web Module."),
			"host":              inputs.NewTagInfo("System hostname."),
			"jolokia_agent_url": inputs.NewTagInfo("Jolokia agent url."),
			"name":              inputs.NewTagInfo("Name"),
		},
	}
}

//nolint:lll
func (m *TomcatCacheM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Desc: "",
		Name: "tomcat_cache",
		Fields: map[string]interface{}{
			"hitCount":    newFielInfoCount("The number of requests for resources that were served from the cache."),
			"lookupCount": newFielInfoCount("The number of requests for resources."),
		},
		Tags: map[string]interface{}{
			"tomcat_context":    inputs.NewTagInfo("Tomcat context."),
			"tomcat_host":       inputs.NewTagInfo("Tomcat host."),
			"host":              inputs.NewTagInfo("System hostname."),
			"jolokia_agent_url": inputs.NewTagInfo("Jolokia agent url."),
		},
	}
}

func newFielInfoInt(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.UnknownUnit,
		Desc:     desc,
	}
}

func newFielInfoCount(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}
