package smart

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	defSmartCtlPath = "/usr/bin/smartctl"
	defNvmePath     = "/usr/bin/nvme"
	inputName       = "smart"
	SampleConfig    = `
[[inputs.smart]]
	## The path to the smartctl executable
  # path_smartctl = "/usr/bin/smartctl"

  ## The path to the nvme-cli executable
  # path_nvme = "/usr/bin/nvme"

	## Gathering interval
	# interval = "10s"

  ## Timeout for the cli command to complete.
  # timeout = "30s"

  ## Optionally specify if vendor specific attributes should be propagated for NVMe disk case
  ## ["auto-on"] - automatically find and enable additional vendor specific disk info
  ## ["vendor1", "vendor2", ...] - e.g. "Intel" enable additional Intel specific disk info
  # enable_extensions = ["auto-on"]

  ## On most platforms used cli utilities requires root access.
  ## Setting 'use_sudo' to true will make use of sudo to run smartctl or nvme-cli.
  ## Sudo must be configured to allow the telegraf user to run smartctl or nvme-cli
  ## without a password.
  # use_sudo = false

  ## Skip checking disks in this power mode. Defaults to
  ## "standby" to not wake up disks that have stopped rotating.
  ## See --nocheck in the man pages for smartctl.
  ## smartctl version 5.41 and 5.42 have faulty detection of
  ## power mode and might require changing this value to
  ## "never" depending on your disks.
  # nocheck = "standby"

  ## Gather all returned S.M.A.R.T. attribute metrics and the detailed
  ## information from each drive into the 'smart_attribute' measurement.
  # attributes = false

  ## Optionally specify devices to exclude from reporting if disks auto-discovery is performed.
  # excludes = [ "/dev/pass6" ]

  ## Optionally specify devices and device type, if unset
  ## a scan (smartctl --scan and smartctl --scan -d nvme) for S.M.A.R.T. devices will be done
  ## and all found will be included except for the excluded in excludes.
  # devices = [ "/dev/ada0 -d atacam", "/dev/nvme0"]
`
	l = logger.SLogger(inputName)
)

type Input struct {
	SmartCtlPath string
	NvmePath     string
	Interval     datakit.Duration
	Timeout      datakit.Duration
}

func (*Input) Catalog() string {
	return inputName
}

func (*Input) SampleConfig() string {
	return SampleConfig
}

func (*Input) AvailabelArch() []string {
	return datakit.AllArch
}

func (i *Input) Run() {

}

func (i *Input) gather() {

}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
