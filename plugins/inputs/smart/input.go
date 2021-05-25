package smart

import (
	"bufio"
	"fmt"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmd"
	ipath "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/path"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	defSmartCmd     = "smartctl"
	defSmartCtlPath = "/usr/bin/smartctl"
	defNvmeCmd      = "nvme"
	defNvmePath     = "/usr/bin/nvme"
	defInterval     = datakit.Duration{Duration: 10 * time.Second}
	defTimeout      = datakit.Duration{Duration: 3 * time.Second}
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

  ## Skip checking disks in this power mode. Defaults to "standby" to not wake up disks that have stopped rotating.
  ## See --nocheck in the man pages for smartctl.
  ## smartctl version 5.41 and 5.42 have faulty detection of power mode and might require changing this value to "never" depending on your disks.
  # no_check = "standby"

  ## Gather all returned S.M.A.R.T. attribute metrics and the detailed
  ## information from each drive into the 'smart_attribute' measurement.
  # attributes = false

  ## Optionally specify devices to exclude from reporting if disks auto-discovery is performed.
  # excludes = [ "/dev/pass6" ]

  ## Optionally specify devices and device type, if unset a scan (smartctl --scan and smartctl --scan -d nvme) for S.M.A.R.T. devices will be done
  ## and all found will be included except for the excluded in excludes.
  # devices = [ "/dev/ada0 -d atacam", "/dev/nvme0"]
`
	l = logger.SLogger(inputName)
)

type nvmeDevice struct {
	name         string
	vendorID     string
	model        string
	serialNumber string
}

type Input struct {
	SmartCtlPath     string           `toml:"smartctl_path"`
	NvmePath         string           `toml:"nvme_path"`
	Interval         datakit.Duration `toml:"interval"`
	Timeout          datakit.Duration `toml:"timeout"`
	EnableExtensions []string         `toml:"enable_extensions"`
	UseSudo          bool             `toml:"use_sudo"`
	NoCheck          string           `toml:"no_check"`
	Attributes       bool             `toml:"attributes"`
	Excludes         []string         `toml:"excludes"`
	Devices          []string         `toml:"devices"`
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

func (s *Input) Run() {
	l.Info("smartctl input started")

	var err error
	if s.SmartCtlPath == "" || !ipath.IsFileExists(s.SmartCtlPath) {
		if s.SmartCtlPath, err = exec.LookPath(defSmartCmd); err != nil {
			l.Errorf("Can not find executable sensor command, install 'smartmontools' first.")

			return
		}
		l.Info("Command fallback to %q due to invalide path provided in 'smart' input", s.SmartCtlPath)
	}
	if s.NvmePath == "" || !ipath.IsFileExists(s.NvmePath) {
		if s.NvmePath, err = exec.LookPath(defNvmeCmd); err != nil {
			l.Errorf("Can not find executable sensor command, install 'nvme-cli' first.")

			return
		}
		l.Info("Command fallback to %q due to invalide path provided in 'smart' input", s.NvmePath)
	}

	tick := time.NewTicker(s.Interval.Duration)
	for {
		select {
		case <-tick.C:
			if err := s.gather(); err != nil {
				l.Error(err.Error())
				io.FeedLastError(inputName, err.Error())
				continue
			}
		case <-datakit.Exit.Wait():
			l.Info("smart input exits")

			return
		}
	}
}

// Gather takes in an accumulator and adds the metrics that the SMART tools gather.
func (s *Input) gather() error {
	var err error
	var scannedNVMeDevices []string
	var scannedNonNVMeDevices []string

	devicesFromConfig := s.Devices
	isNVMe := len(s.NvmePath) != 0
	isVendorExtension := len(s.EnableExtensions) != 0

	if len(s.Devices) != 0 {
		s.getAttributes(devicesFromConfig)
		// if nvme-cli is present, vendor specific attributes can be gathered
		if isVendorExtension && isNVMe {
			scannedNVMeDevices, _, err = s.scanAllDevices(true)
			if err != nil {
				return err
			}
			nvmeDevices := distinguishNVMeDevices(devicesFromConfig, scannedNVMeDevices)
			s.getVendorNVMeAttributes(nvmeDevices)
		}

		return nil
	}

	scannedNVMeDevices, scannedNonNVMeDevices, err = s.scanAllDevices(false)
	if err != nil {
		return err
	}
	var devicesFromScan []string
	devicesFromScan = append(devicesFromScan, scannedNVMeDevices...)
	devicesFromScan = append(devicesFromScan, scannedNonNVMeDevices...)

	s.getAttributes(devicesFromScan)
	if isVendorExtension && isNVMe {
		s.getVendorNVMeAttributes(scannedNVMeDevices)
	}

	return nil
}

func excludedDevice(excludes []string, deviceLine string) bool {
	device := strings.Split(deviceLine, " ")
	if len(device) != 0 {
		for _, exclude := range excludes {
			if device[0] == exclude {
				return true
			}
		}
	}

	return false
}

// Scan for S.M.A.R.T. devices from smartctl
func (s *Input) scanDevices(ignoreExcludes bool, scanArgs ...string) ([]string, error) {
	output, err := cmd.RunWithTimeout(s.Timeout.Duration, s.UseSudo, s.SmartCtlPath, scanArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to run command '%s %s': %s - %s", s.SmartCtlPath, scanArgs, err, string(output))
	}

	var devices []string
	for _, line := range strings.Split(string(output), "\n") {
		dev := strings.Split(line, " ")
		if len(dev) <= 1 {
			continue
		}
		if !ignoreExcludes {
			if !excludedDevice(s.Excludes, strings.TrimSpace(dev[0])) {
				devices = append(devices, strings.TrimSpace(dev[0]))
			}
		} else {
			devices = append(devices, strings.TrimSpace(dev[0]))
		}
	}

	return devices, nil
}

func (s *Input) scanAllDevices(ignoreExcludes bool) ([]string, []string, error) {
	// this will return all devices (including NVMe devices) for smartctl version >= 7.0
	// for older versions this will return non NVMe devices
	devices, err := s.scanDevices(ignoreExcludes, "--scan")
	if err != nil {
		return nil, nil, err
	}

	// this will return only NVMe devices
	nvmeDevices, err := s.scanDevices(ignoreExcludes, "--scan", "--device=nvme")
	if err != nil {
		return nil, nil, err
	}

	// to handle all versions of smartctl this will return only non NVMe devices
	nonNVMeDevices := difference(devices, nvmeDevices)

	return nvmeDevices, nonNVMeDevices, nil
}

// Get info and attributes for each S.M.A.R.T. device
func (s *Input) getAttributes(devices []string) {
	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(len(devices))
	for _, device := range devices {
		go func() {
			if cache, err := gatherDisk(s.Timeout.Duration, s.UseSudo, s.Attributes, s.SmartCtlPath, s.NoCheck, device); err != nil {
				io.FeedLastError(inputName, err.Error())
			} else {
				if err := inputs.FeedMeasurement(inputName, datakit.Metric, cache, &io.Option{CollectCost: time.Now().Sub(start)}); err != nil {
					io.FeedLastError(inputName, err.Error())
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func (s *Input) getVendorNVMeAttributes(devices []string) {
	start := time.Now()
	nvmeDevices := getDeviceInfoForNVMeDisks(devices, s.NvmePath, s.Timeout.Duration, s.UseSudo)

	var wg sync.WaitGroup
	for _, device := range nvmeDevices {
		if contains(s.EnableExtensions, "auto-on") {
			switch device.vendorID {
			case intelVID:
				wg.Add(1)
				go func() {
					if cache, err := gatherIntelNVMeDisk(s.Timeout.Duration, s.UseSudo, s.NvmePath, device); err != nil {
						io.FeedLastError(inputName, err.Error())
					} else {
						if err := inputs.FeedMeasurement(inputName, datakit.Metric, cache, &io.Option{CollectCost: time.Now().Sub(start)}); err != nil {
							io.FeedLastError(inputName, err.Error())
						}
					}
					wg.Done()
				}()
			}
		} else if contains(s.EnableExtensions, "Intel") && device.vendorID == intelVID {
			wg.Add(1)
			go func() {
				if cache, err := gatherIntelNVMeDisk(s.Timeout.Duration, s.UseSudo, s.NvmePath, device); err != nil {
					io.FeedLastError(inputName, err.Error())
				} else {
					if err := inputs.FeedMeasurement(inputName, datakit.Metric, cache, &io.Option{CollectCost: time.Now().Sub(start)}); err != nil {
						io.FeedLastError(inputName, err.Error())
					}
				}
				wg.Done()
			}()
		}
	}
	wg.Wait()
}

func distinguishNVMeDevices(userDevices []string, availableNVMeDevices []string) []string {
	var nvmeDevices []string
	for _, userDevice := range userDevices {
		for _, NVMeDevice := range availableNVMeDevices {
			// double check. E.g. in case when nvme0 is equal nvme0n1, will check if "nvme0" part is present.
			if strings.Contains(NVMeDevice, userDevice) || strings.Contains(userDevice, NVMeDevice) {
				nvmeDevices = append(nvmeDevices, userDevice)
			}
		}
	}

	return nvmeDevices
}

func getDeviceInfoForNVMeDisks(devices []string, nvme string, timeout time.Duration, useSudo bool) []nvmeDevice {
	var nvmeDevices []nvmeDevice
	for _, device := range devices {
		vid, sn, mn, err := gatherNVMeDeviceInfo(nvme, device, timeout, useSudo)
		if err != nil {
			io.FeedLastError(inputName, fmt.Sprintf("cannot find device info for %s device", device))
			continue
		}
		newDevice := nvmeDevice{
			name:         device,
			vendorID:     vid,
			model:        mn,
			serialNumber: sn,
		}
		nvmeDevices = append(nvmeDevices, newDevice)
	}

	return nvmeDevices
}

func gatherNVMeDeviceInfo(nvme, device string, timeout time.Duration, useSudo bool) (string, string, string, error) {
	args := append([]string{"id-ctrl"}, strings.Split(device, " ")...)
	output, err := cmd.RunWithTimeout(timeout, useSudo, nvme, args...)
	if err != nil {
		return "", "", "", err
	}

	return findNVMeDeviceInfo(string(output))
}

func findNVMeDeviceInfo(output string) (string, string, string, error) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	var vid, sn, mn string

	for scanner.Scan() {
		line := scanner.Text()

		if matches := nvmeIDCtrlExpressionPattern.FindStringSubmatch(line); len(matches) > 2 {
			matches[1] = strings.TrimSpace(matches[1])
			matches[2] = strings.TrimSpace(matches[2])
			if matches[1] == "vid" {
				if _, err := fmt.Sscanf(matches[2], "%s", &vid); err != nil {
					return "", "", "", err
				}
			}
			if matches[1] == "sn" {
				sn = matches[2]
			}
			if matches[1] == "mn" {
				mn = matches[2]
			}
		}
	}

	return vid, sn, mn, nil
}

func gatherIntelNVMeDisk(timeout time.Duration, useSudo bool, nvme string, device nvmeDevice) ([]inputs.Measurement, error) {
	args := append([]string{"intel", "smart-log-add"}, strings.Split(device.name, " ")...)
	output, err := cmd.RunWithTimeout(timeout, useSudo, nvme, args...)
	if _, err = cmd.ExitStatus(err); err != nil {
		return nil, fmt.Errorf("failed to run command '%s %s': %s - %s", nvme, strings.Join(args, " "), err, string(output))
	}

	var (
		cache   []inputs.Measurement
		scanner = bufio.NewScanner(strings.NewReader(string(output)))
	)
	for scanner.Scan() {
		line := scanner.Text()
		tags := map[string]string{}
		fields := make(map[string]interface{})

		tags["device"] = path.Base(device.name)
		tags["model"] = device.model
		tags["serial_no"] = device.serialNumber

		if matches := intelExpressionPattern.FindStringSubmatch(line); len(matches) > 3 {
			matches[1] = strings.TrimSpace(matches[1])
			matches[3] = strings.TrimSpace(matches[3])
			if attr, ok := intelAttributes[matches[1]]; ok {
				tags["name"] = attr.Name
				if attr.ID != "" {
					tags["id"] = attr.ID
				}

				parse := parseCommaSeparatedIntWithCache
				if attr.Parse != nil {
					parse = attr.Parse
				}

				if err := parse(&cache, fields, tags, matches[3]); err != nil {
					continue
				}
			}
		}
	}

	return cache, nil
}

func gatherDisk(timeout time.Duration, sudo, collectAttributes bool, smartctl, nocheck, device string) ([]inputs.Measurement, error) {
	// smartctl 5.41 & 5.42 have are broken regarding handling of --nocheck/-n
	args := append([]string{"--info", "--health", "--attributes", "--tolerance=verypermissive", "-n", nocheck, "--format=brief"}, strings.Split(device, " ")...)
	output, err := cmd.RunWithTimeout(timeout, sudo, smartctl, args...)
	// Ignore all exit statuses except if it is a command line parse error
	exitStatus, err := cmd.ExitStatus(err)
	if err != nil {
		return nil, err
	}

	deviceTags := map[string]string{}
	deviceNode := strings.Split(device, " ")[0]
	deviceTags["device"] = path.Base(deviceNode)
	deviceFields := make(map[string]interface{})
	deviceFields["exit_status"] = exitStatus

	var (
		cache   []inputs.Measurement
		scanner = bufio.NewScanner(strings.NewReader(string(output)))
	)
	for scanner.Scan() {
		line := scanner.Text()

		model := modelInfo.FindStringSubmatch(line)
		if len(model) > 2 {
			deviceTags["model"] = model[2]
		}

		serial := serialInfo.FindStringSubmatch(line)
		if len(serial) > 1 {
			deviceTags["serial_no"] = serial[1]
		}

		wwn := wwnInfo.FindStringSubmatch(line)
		if len(wwn) > 1 {
			deviceTags["wwn"] = strings.Replace(wwn[1], " ", "", -1)
		}

		capacity := userCapacityInfo.FindStringSubmatch(line)
		if len(capacity) > 1 {
			deviceTags["capacity"] = strings.Replace(capacity[1], ",", "", -1)
		}

		enabled := smartEnabledInfo.FindStringSubmatch(line)
		if len(enabled) > 1 {
			deviceTags["enabled"] = enabled[1]
		}

		health := smartOverallHealth.FindStringSubmatch(line)
		if len(health) > 2 {
			deviceFields["health_ok"] = health[2] == "PASSED" || health[2] == "OK"
		}

		tags := map[string]string{}
		fields := make(map[string]interface{})

		if collectAttributes {
			for _, key := range [...]string{"device", "model", "serial_no", "wwn", "capacity", "enabled"} {
				if value, ok := deviceTags[key]; ok {
					tags[key] = value
				}
			}
		}

		attr := attribute.FindStringSubmatch(line)
		if len(attr) > 1 {
			// attribute has been found, add it only if collectAttributes is true
			if collectAttributes {
				tags["id"] = attr[1]
				tags["name"] = attr[2]
				tags["flags"] = attr[3]

				fields["exit_status"] = exitStatus
				if i, err := strconv.ParseInt(attr[4], 10, 64); err == nil {
					fields["value"] = i
				}
				if i, err := strconv.ParseInt(attr[5], 10, 64); err == nil {
					fields["worst"] = i
				}
				if i, err := strconv.ParseInt(attr[6], 10, 64); err == nil {
					fields["threshold"] = i
				}

				tags["fail"] = attr[7]
				if val, err := parseRawValue(attr[8]); err == nil {
					fields["raw_value"] = val
				}

				cache = append(cache, &smartMeasurement{name: "smart_attribute", tags: tags, fields: fields, ts: time.Now()})
			}

			// If the attribute matches on the one in deviceFieldIds
			// save the raw value to a field.
			if field, ok := deviceFieldIds[attr[1]]; ok {
				if val, err := parseRawValue(attr[8]); err == nil {
					deviceFields[field] = val
				}
			}
		} else {
			// what was found is not a vendor attribute
			if matches := sasNvmeAttr.FindStringSubmatch(line); len(matches) > 2 {
				if attr, ok := sasNvmeAttributes[matches[1]]; ok {
					tags["name"] = attr.Name
					if attr.ID != "" {
						tags["id"] = attr.ID
					}

					parse := parseCommaSeparatedInt
					if attr.Parse != nil {
						parse = attr.Parse
					}

					if err := parse(fields, deviceFields, matches[2]); err != nil {
						continue
					}
					// if the field is classified as an attribute, only add it if collectAttributes is true
					if collectAttributes {
						cache = append(cache, &smartMeasurement{name: "smart_attribute", tags: tags, fields: fields, ts: time.Now()})
					}
				}
			}
		}
	}
	cache = append(cache, &smartMeasurement{name: "smart_device", tags: deviceTags, fields: deviceFields, ts: time.Now()})

	return cache, nil
}

func contains(args []string, element string) bool {
	for _, arg := range args {
		if arg == element {
			return true
		}
	}

	return false
}

func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}

	return diff
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			SmartCtlPath: defSmartCtlPath,
			NvmePath:     defNvmePath,
			Interval:     defInterval,
			Timeout:      defTimeout,
		}
	})
}
