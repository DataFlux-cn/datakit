// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package statsd

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

func (ipt *Input) gather() {
	l.Debugf("try locking...")
	ipt.Lock()
	defer ipt.Unlock()
	now := time.Now()

	for _, m := range ipt.distributions {
		fields := map[string]interface{}{
			defaultFieldName: m.value,
		}
		l.Debugf("[distributions] add %s, fields: %+#v, tags: %+#v", m.name, fields, m.tags)
		ipt.acc.addFields(m.name, fields, m.tags, now)
	}
	ipt.distributions = make([]cacheddistributions, 0)

	for _, m := range ipt.timings {
		// Defining a template to parse field names for timers allows us to split
		// out multiple fields per timer. In this case we prefix each stat with the
		// field name and store these all in a single measurement.
		fields := make(map[string]interface{})
		for fieldName, stats := range m.fields {
			var prefix string
			if fieldName != defaultFieldName {
				prefix = fieldName + "_"
			}
			fields[prefix+"mean"] = stats.Mean()
			fields[prefix+"stddev"] = stats.Stddev()
			fields[prefix+"sum"] = stats.Sum()
			fields[prefix+"upper"] = stats.Upper()
			fields[prefix+"lower"] = stats.Lower()
			fields[prefix+"count"] = stats.Count()
			for _, percentile := range ipt.Percentiles {
				name := fmt.Sprintf("%s%v_percentile", prefix, percentile)
				fields[name] = stats.Percentile(percentile)
			}
		}

		l.Debugf("[timings] add %s, fields: %+#v, tags: %+#v", m.name, fields, m.tags)
		ipt.acc.addFields(m.name, fields, m.tags, now)
	}
	if ipt.DeleteTimings {
		ipt.timings = make(map[string]cachedtimings)
	}

	for _, m := range ipt.gauges {
		l.Debugf("[gauges] add %s, fields: %+#v, tags: %+#v", m.name, m.fields, m.tags)
		ipt.acc.addFields(m.name, m.fields, m.tags, now)
	}
	if ipt.DeleteGauges {
		ipt.gauges = make(map[string]cachedgauge)
	}

	for _, m := range ipt.counters {
		l.Debugf("[counters] add %s, fields: %+#v, tags: %+#v", m.name, m.fields, m.tags)
		ipt.acc.addFields(m.name, m.fields, m.tags, now)
	}
	if ipt.DeleteCounters {
		ipt.counters = make(map[string]cachedcounter)
	}

	for _, m := range ipt.sets {
		fields := make(map[string]interface{})
		for field, set := range m.fields {
			fields[field] = int64(len(set))
		}
		l.Debugf("[sets] add %s, fields: %+#v, tags: %+#v", m.name, m.fields, m.tags)
		ipt.acc.addFields(m.name, fields, m.tags, now)
	}

	if ipt.DeleteSets {
		ipt.sets = make(map[string]cachedset)
	}

	l.Debugf("feed %d points...", len(ipt.acc.measurements))
	if len(ipt.acc.measurements) > 0 {
		if err := ipt.Feeder.Feed(ipt.acc.feedMetricName, point.Metric, ipt.acc.measurements, &dkio.Option{CollectCost: time.Since(now)}); err != nil {
			l.Error(err)
		} else {
			ipt.acc.measurements = ipt.acc.measurements[:0]
		}
	}

	ipt.expireCachedMetrics()
}
