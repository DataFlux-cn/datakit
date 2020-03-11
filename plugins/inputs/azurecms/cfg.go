package azurecms

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2019-06-01/insights"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
)

const (
	sampleConfig = `#[[instances]]
# client_id = ''
# client_secret = ''
# tenant_id = ''
# subscription_id = ''
# #end_point = 'https://management.chinacloudapi.cn'

#[[instances.resource]]
#resource_id = ''

#[[instances.resource.metrics]]
#metric_name = 'Percentage CPU'
# #interval = '1m'
`
)

type (
	azureMetric struct {
		MetricName string
		Interval   internal.Duration
	}

	azureResource struct {
		ResourceID string
		Metrics    []*azureMetric
	}

	azureInstance struct {
		ClientID       string
		ClientSecret   string
		TenantID       string
		SubscriptionID string
		EndPoint       string //https://management.chinacloudapi.cn

		Resource []*azureResource
	}

	metricMeta struct {
		supportTimeGrain      []int64
		supportedAggregations []string
		unit                  string
	}

	queryListInfo struct {
		meta *metricMeta

		resourceID   string
		timespan     string
		intervalTime time.Duration
		interval     string
		metricname   string
		aggregation  string
		top          int32
		orderby      string
		filter       string
		resultType   insights.ResultType
		//apiVersion  string // "2018-01-01"
		metricnamespace string // `Microsoft.Compute/virtualMachines`

		lastFetchTime time.Time
	}
)

func (a *azureInstance) genQueryInfo() []*queryListInfo {

	var infos []*queryListInfo

	for _, res := range a.Resource {
		for _, ms := range res.Metrics {
			info := &queryListInfo{
				resourceID:   res.ResourceID,
				metricname:   ms.MetricName,
				intervalTime: ms.Interval.Duration,
				interval:     convertInterval(ms.Interval.Duration),
			}
			if info.intervalTime < time.Minute {
				info.intervalTime = time.Minute
			}
			infos = append(infos, info)
		}
	}

	return infos
}

func convertInterval(interval time.Duration) string {

	if interval == time.Minute {
		return "PT1M"
	} else if interval == 5*time.Minute {
		return "PT5M"
	} else if interval == 15*time.Minute {
		return "PT15M"
	} else if interval == 30*time.Minute {
		return "PT30M"
	} else if interval == time.Hour {
		return "PT1H"
	} else if interval == 6*time.Hour {
		return "PT6H"
	} else if interval == 12*time.Hour {
		return "PT12H"
	} else if interval == 24*time.Hour {
		return "P1D"
	}
	return "PT1M"
}

func unixTimeStrISO8601(t time.Time) string {
	_, zoff := t.Zone()
	nt := t.Add(-(time.Duration(zoff) * time.Second))
	s := nt.Format(`2006-01-02T15:04:05Z`)
	return s
}
