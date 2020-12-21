// +build windows

package telegraf_inputs

import (
	"github.com/influxdata/telegraf/plugins/inputs/win_perf_counters"
	"github.com/influxdata/telegraf/plugins/inputs/win_services"
)

var (
	telegrafInputsWin = map[string]*TelegrafInput{
		"win_services":      {name: "win_services", Catalog: "windows", Input: &win_services.WinServices{}},
		"win_perf_counters": {name: "win_perf_counters", Catalog: "windows", Input: &win_perf_counters.Win_PerfCounters{}},
		`dotnetclr`:         {name: "dotnetclr", Catalog: "windows", Sample: samples["dotnetclr"], Input: &win_perf_counters.Win_PerfCounters{}},
		`aspdotnet`:         {name: "aspdotnet", Catalog: "windows", Sample: samples["aspdotnet"], Input: &win_perf_counters.Win_PerfCounters{}},
		`msexchange`:        {name: "msexchange", Catalog: "windows", Sample: samples["msexchange"], Input: &win_perf_counters.Win_PerfCounters{}},
		`iis`:               {name: "iis", Catalog: "windows", Sample: samples["iis"], Input: &win_perf_counters.Win_PerfCounters{}},
		`active_directory`:  {name: "active_directory", Catalog: "windows", Sample: samples["active_directory"], Input: &win_perf_counters.Win_PerfCounters{}},
	}
)

func init() {
	for k, v := range telegrafInputsWin {
		TelegrafInputs[k] = v
	}
}
