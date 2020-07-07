package inputs

import (
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/aliyunactiontrail"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/aliyuncdn"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/aliyuncms"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/aliyuncost"

	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/aliyunddos"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/aliyunlog"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/aliyunprice"

	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/aliyunrdsslowlog"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/aliyunsecurity"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/awsbill"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/baiduIndex"

	//_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/containerd"
	//_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/coredns"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/dataclean"
	//_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/etcd"
	//_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/gitlab"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/harborMonitor"
	//_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/hostobject"
	//_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/httpstat"
	//_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/jira"
	//_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/lighttpd"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/druid"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/mock"

	//_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/mongodboplog"
	//_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/mysqlmonitor"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/prometheus"
	//_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/replication"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/self"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/squid"

	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/ssh"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/statsd"

	//_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/tailf"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/tencentcms"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/timezone"

	//_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"

	//_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/traefik"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/ucmon"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/yarn"

	//_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/zabbix"

	// 32bit disabled, only 64 bit available
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/binlog"

	// external inputs wrap
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/oraclemonitor"
	// with dll/so dependencies, and also 32bit disabled

	// only windows
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/wmi"
)
