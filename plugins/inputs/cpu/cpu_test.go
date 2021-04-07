package cpu

import (
	"testing"

	"github.com/shirou/gopsutil/cpu"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func TestCpuActiveTotalTime(t *testing.T) {
	cputime := cpu.TimesStat{
		CPU:       "cpu-total",
		User:      17105.0, // delta -5.0
		System:    8550.5,  // delta -2.7
		Idle:      83678.4, // delta -56.4
		Nice:      226.9,
		Iowait:    4758.5,
		Irq:       0.0,
		Softirq:   137.8,
		Steal:     0.0,
		Guest:     0.0,
		GuestNice: 0.0,
	}

	active, total := CpuActiveTotalTime(cputime)
	if total != cputime.Total() {
		t.Errorf("error: The CPU total time should be %.2f, but now it's %.2f", cputime.Total(), total)
	}
	if active != cputime.Total()-cputime.Idle {
		t.Errorf("error: The CPU active time should be %.2f, but now it's %.2f", cputime.Total()-cputime.Idle, active)
	}
}

func TestCPUStatStructToMap(t *testing.T) {
	mU := make(map[string]interface{})

	sU := UsageStat{}

	if ok := CPUStatStructToMap(mU, sU, ""); !ok {
		t.Errorf("Failed to convert cpu usage stat struct to map. There may be private variables in struct")
	}

	mUe := map[string]interface{}{
		"user":       0,
		"system":     0,
		"idle":       0,
		"nice":       0,
		"iowait":     0,
		"irq":        0,
		"softirq":    0,
		"steal":      0,
		"guest":      0,
		"guest_nice": 0,
	}

	for k := range mU {
		if _, ok := mUe[k]; !ok {
			t.Errorf("error: CPU usage stat unexpected field: %s", k)
		}
	}
	if len(mU) != len(mUe) {
		t.Errorf("error: CPU usage stat have %d fields, but expect %d", len(mU), len(mUe))
	}
}

func TestCalculateUsage(t *testing.T) {
	lastT := cpu.TimesStat{
		CPU:       "cpu-total",
		User:      17105.0, // delta -5.0
		System:    8550.5,  // delta -2.7
		Idle:      83678.4, // delta -56.4
		Nice:      226.9,
		Iowait:    4758.5,
		Irq:       0.0,
		Softirq:   137.8,
		Steal:     0.0,
		Guest:     0.0,
		GuestNice: 0.0,
	}
	nowT := cpu.TimesStat{
		CPU:       "cpu-total",
		User:      17110.0,
		System:    8553.2,
		Idle:      83734.8,
		Nice:      226.9,
		Iowait:    4758.5,
		Irq:       0.0,
		Softirq:   137.8,
		Steal:     0.0,
		Guest:     0.0,
		GuestNice: 0.0,
	}
	nowTCpu0 := cpu.TimesStat{
		CPU:       "cpu0",
		User:      17110.0,
		System:    8553.2,
		Idle:      83734.8,
		Nice:      226.9,
		Iowait:    4758.5,
		Irq:       0.0,
		Softirq:   137.8,
		Steal:     0.0,
		Guest:     0.0,
		GuestNice: 0.0,
	}

	_, nowTotal := CpuActiveTotalTime(nowT)
	_, lastTotal := CpuActiveTotalTime(lastT)
	totalDelta := nowTotal - lastTotal

	if _, err := CalculateUsage(nowTCpu0, lastT, totalDelta); err == nil {
		t.Error("error: Use data of different CPUs to calculate CPU utilization should be disabled.")
	}
	uSt, _ := CalculateUsage(nowT, lastT, totalDelta)
	assertEqualFloat64(t, 100*(nowT.User-lastT.User-(nowT.Guest-lastT.Guest))/totalDelta, uSt.User, "user")
	assertEqualFloat64(t, 100*(nowT.System-lastT.System)/totalDelta, uSt.System, "system")
	assertEqualFloat64(t, 100*(nowT.Idle-lastT.Idle)/totalDelta, uSt.Idle, "idle")
	assertEqualFloat64(t, 100*(nowT.Nice-lastT.Nice-(nowT.GuestNice-lastT.GuestNice))/totalDelta, uSt.Nice, "nice")
	assertEqualFloat64(t, 100*(nowT.Iowait-lastT.Iowait)/totalDelta, uSt.Iowait, "iowait")
	assertEqualFloat64(t, 100*(nowT.Irq-lastT.Irq)/totalDelta, uSt.Irq, "irq")
	assertEqualFloat64(t, 100*(nowT.Softirq-lastT.Softirq)/totalDelta, uSt.Softirq, "softirq")
	assertEqualFloat64(t, 100*(nowT.Steal-lastT.Steal)/totalDelta, uSt.Steal, "steal")
	assertEqualFloat64(t, 100*(nowT.Guest-lastT.Guest)/totalDelta, uSt.Guest, "guest")
	assertEqualFloat64(t, 100*(nowT.GuestNice-lastT.GuestNice)/totalDelta, uSt.GuestNice, "guest_nice")
}

func TestCollect(t *testing.T) {
	lastT := cpu.TimesStat{
		CPU:       "cpu-total",
		User:      17105.0, // delta -5.0
		System:    8550.5,  // delta -2.7
		Idle:      83678.4, // delta -56.4
		Nice:      226.9,
		Iowait:    4758.5,
		Irq:       0.0,
		Softirq:   137.8,
		Steal:     0.0,
		Guest:     0.0,
		GuestNice: 0.0,
	}
	nowT := cpu.TimesStat{
		CPU:       "cpu-total",
		User:      17110.0,
		System:    8553.2,
		Idle:      83734.8,
		Nice:      226.9,
		Iowait:    4758.5,
		Irq:       0.0,
		Softirq:   137.8,
		Steal:     0.0,
		Guest:     0.0,
		GuestNice: 0.0,
	}

	timeStats := [][]cpu.TimesStat{
		[]cpu.TimesStat{lastT},
		[]cpu.TimesStat{nowT},
	}
	i := &Input{logger: logger.SLogger("cpu"), ps: &CPUInfoTest{timeStat: timeStats}}
	if err := i.Collect(); err != nil {
		t.Error(err)
	} else if len(i.collectCache) != 0 {
		t.Error("")
	}
	if err := i.Collect(); err != nil {
		t.Error(err)
	} else if len(i.collectCache) != 1 {
		t.Error("")
	}

	// tags
	tags := i.collectCache[0].(*cpuMeasurement).tags
	if tags["cpu"] != "cpu-total" {
		t.Errorf("cpu:%s expected: cpu-total", tags["cpu"])
	}
	if tags["host"] != datakit.Cfg.MainCfg.Hostname {
		t.Errorf("host name:%s expected: %s", tags["cpu"], datakit.Cfg.MainCfg.Hostname)
	}

	// fields
	fields := i.collectCache[0].(*cpuMeasurement).fields

	_, nowTotal := CpuActiveTotalTime(nowT)
	_, lastTotal := CpuActiveTotalTime(lastT)
	totalDelta := nowTotal - lastTotal
	// active, _ := CpuActiveTotalTime(nowT)
	// lastActive, _ := CpuActiveTotalTime(lastT)

	assertEqualFloat64(t, 100*(nowT.User-lastT.User-(nowT.Guest-lastT.Guest))/totalDelta, fields["user"].(float64), "user")
	assertEqualFloat64(t, 100*(nowT.System-lastT.System)/totalDelta, fields["system"].(float64), "system")
	assertEqualFloat64(t, 100*(nowT.Idle-lastT.Idle)/totalDelta, fields["idle"].(float64), "idle")
	assertEqualFloat64(t, 100*(nowT.Nice-lastT.Nice-(nowT.GuestNice-lastT.GuestNice))/totalDelta, fields["nice"].(float64), "nice")
	assertEqualFloat64(t, 100*(nowT.Iowait-lastT.Iowait)/totalDelta, fields["iowait"].(float64), "iowait")
	assertEqualFloat64(t, 100*(nowT.Irq-lastT.Irq)/totalDelta, fields["irq"].(float64), "irq")
	assertEqualFloat64(t, 100*(nowT.Softirq-lastT.Softirq)/totalDelta, fields["softirq"].(float64), "softirq")
	assertEqualFloat64(t, 100*(nowT.Steal-lastT.Steal)/totalDelta, fields["steal"].(float64), "steal")
	assertEqualFloat64(t, 100*(nowT.Guest-lastT.Guest)/totalDelta, fields["guest"].(float64), "guest")
	assertEqualFloat64(t, 100*(nowT.GuestNice-lastT.GuestNice)/totalDelta, fields["guest_nice"].(float64), "guest_nice")
	// assertEqualFloat64(t, 100*(active-lastActive)/totalDelta, fields["active"].(float64), "active")
}

func TestHumpToUnderline(t *testing.T) {
	if HumpToUnderline("SaRRdDD") != "sa_r_rd_dd" {
		t.Errorf("error: SaRRdDD --> %s", HumpToUnderline("SaRRdDD"))
	}
	if HumpToUnderline("UserAgent") != "user_agent" {
		t.Errorf("error: UserAgent --> %s", HumpToUnderline("UserAgent"))
	}
	if HumpToUnderline("userAgent") != "user_agent" {
		t.Errorf("error: userAgent --> %s", HumpToUnderline("userAgent"))
	}
	if HumpToUnderline("aaa") != "aaa" {
		t.Errorf("error: aaa --> %s", HumpToUnderline("aaa"))
	}
	if HumpToUnderline("AAA") != "aaa" {
		t.Errorf("error: AAA --> %s", HumpToUnderline("AAA"))
	}
}

func assertEqualFloat64(t *testing.T, expected, actual float64, mName string) {
	if expected != actual {
		t.Errorf("error: "+mName+" expected: %f \t actual %f", expected, actual)
	}
}
