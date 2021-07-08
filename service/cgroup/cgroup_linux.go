package cgroup

import (
	"os"
	"runtime"
	"time"

	"github.com/containerd/cgroups"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/shirou/gopsutil/cpu"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
)

var (
	period = uint64(1000000) // 1 second
)

func start() {
	// config.Cfg.Cgroup 是百分比
	high := config.Cfg.Cgroup.CPUMax * float64(runtime.NumCPU()) / 100
	low := config.Cfg.Cgroup.CPUMin * float64(runtime.NumCPU()) / 100

	quotaHigh := int64(float64(period) * high)
	quotaLow := int64(float64(period) * low)

	pid := os.Getpid()

	l.Infof("with %d CPU, set CPU limimt %.2f%%", runtime.NumCPU(), float64(quotaLow)/float64(period)*100.0)

	control, err := cgroups.New(cgroups.V1, cgroups.StaticPath("/datakit"),
		&specs.LinuxResources{
			CPU: &specs.LinuxCPU{
				Period: &period,
				Quota:  &quotaLow,
			}})
	if err != nil {
		l.Errorf("failed of new cgroup: %s", err)
		return
	}
	defer control.Delete()

	if err := control.Add(cgroups.Process{Pid: pid}); err != nil {
		l.Errorf("faild of add cgroup: %s", err)
		return
	}

	l.Infof("add PID %d to cgroup", pid)

	level := "low"
	waitNum := 0
	for {
		percpu, err := getCPUPercent(time.Second * 3)
		if err != nil {
			l.Debug(err)
			continue
		}

		var q int64

		if 100 < percpu+high {
			if level == "low" {
				continue
			}
			q = quotaLow
			level = "low"
		} else {
			if level == "high" {
				continue
			}
			if waitNum < 3 {
				waitNum++
				continue
			}
			q = quotaHigh
			level = "high"
			waitNum = 0
		}

		err = control.Update(&specs.LinuxResources{
			CPU: &specs.LinuxCPU{
				Period: &period,
				Quota:  &q,
			}})
		if err != nil {
			l.Debugf("failed of update cgroup: %s", err)
			continue
		}
		l.Debugf("switch to quota %.2f%%", float64(q)/float64(period)*100.0)
	}
}

func getCPUPercent(interval time.Duration) (float64, error) {
	percent, err := cpu.Percent(interval, false)
	if err != nil {
		return 0, err
	}
	if len(percent) < 0 {
		return 0, nil
	}
	return percent[0] / 100, nil
}