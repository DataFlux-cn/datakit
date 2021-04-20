package diskio

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/shirou/gopsutil/disk"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName    = "diskio"
	metricName   = "diskio"
	collectCycle = time.Second * 10
	diskioLogger = logger.DefaultSLogger(inputName)
	varRegex     = regexp.MustCompile(`\$(?:\w+|\{\w+\})`)
	sampleConfig = `
[[inputs.diskio]]
  ## By default, gather stats for all devices including
  ## disk partitions.
  ## Setting interfaces using regular expressions will collect these expected devices.
  # devices = ['''^sda\d*''', '''^sdb\d*''', '''vd.*''']
  #
  ## If the disk serial number is not required, please uncomment the following line.
  # skip_serial_number = true
  #
  ## On systems which support it, device metadata can be added in the form of
  ## tags.
  ## Currently only Linux is supported via udev properties. You can view
  ## available properties for a device by running:
  ## 'udevadm info -q property -n /dev/sda'
  ## Note: Most, but not all, udev properties can be accessed this way. Properties
  ## that are currently inaccessible include DEVTYPE, DEVNAME, and DEVPATH.
  # device_tags = ["ID_FS_TYPE", "ID_FS_USAGE"]
  # 
  ## Using the same metadata source as device_tags, 
  ## you can also customize the name of the device through a template. 
  ## The "name_templates" parameter is a list of templates to try to apply equipment. 
  ## The template can contain variables of the form "$PROPERTY" or "${PROPERTY}". 
  ## The first template that does not contain any variables that do not exist 
  ## for the device is used as the device name label. 
  ## A typical use case for LVM volumes is to obtain VG/LV names, 
  ## not DM-0 names which are almost meaningless. 
  ## In addition, "device" is reserved specifically to indicate the device name. 
  # name_templates = ["$ID_FS_LABEL","$DM_VG_NAME/$DM_LV_NAME", "$device:$ID_FS_TYPE"]
  # 
  [inputs.diskio.tags]
    # tag1 = "a"
  `
)

type diskioMeasurement measurement

func (m *diskioMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

// https://www.kernel.org/doc/Documentation/ABI/testing/procfs-diskstats

func (m *diskioMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "diskio",
		Fields: map[string]interface{}{
			"reads":            newFieldsInfoCount("reads completed successfully"),
			"writes":           newFieldsInfoCount("writes completed"),
			"read_bytes":       newFieldsInfoBytes("read bytes"),
			"write_bytes":      newFieldsInfoBytes("write bytes"),
			"read_time":        newFieldsInfoMS("time spent reading"),
			"write_time":       newFieldsInfoMS("time spent writing"),
			"io_time":          newFieldsInfoMS("time spent doing I/Os"),
			"weighted_io_time": newFieldsInfoMS("weighted time spent doing I/Os"),
			"iops_in_progress": newFieldsInfoCount("I/Os currently in progress"),
			"merged_reads":     newFieldsInfoCount("reads merged"),
			"merged_writes":    newFieldsInfoCount("writes merged"),
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "主机名"},
			"name": &inputs.TagInfo{Desc: "磁盘设备名"},
		},
	}
}

type Input struct {
	Devices          []string
	DeviceTags       []string
	NameTemplates    []string
	SkipSerialNumber bool
	Tags             map[string]string

	collectCache []inputs.Measurement

	diskIO DiskIO

	infoCache    map[string]diskInfoCache
	deviceFilter *DevicesFilter
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) Catalog() string {
	return "host"
}

func (i *Input) SampleConfig() string {
	return sampleConfig
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&diskioMeasurement{},
	}
}

func (i *Input) Collect() error {
	// 设置 disk device 过滤器
	i.deviceFilter = &DevicesFilter{}
	err := i.deviceFilter.Compile(i.Devices)
	if err != nil {
		return err
	}

	// disk io stat
	diskio, err := i.diskIO([]string{}...)
	if err != nil {
		return fmt.Errorf("error getting disk io info: %s", err.Error())
	}

	ts := time.Now()
	for _, io := range diskio {
		match := false

		// 匹配 disk name
		if len(i.deviceFilter.filters) < 1 || i.deviceFilter.Match(io.Name) {
			match = true
		}

		tags := map[string]string{}
		// 用户自定义tags
		for k, v := range i.Tags {
			tags[k] = v
		}
		var devLinks []string

		tags["name"], devLinks = i.diskName(io.Name)

		if !match {
			for _, devLink := range devLinks {
				if i.deviceFilter.Match(devLink) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		for t, v := range i.diskTags(io.Name) {
			tags[t] = v
		}

		if !i.SkipSerialNumber {
			if len(io.SerialNumber) != 0 {
				tags["serial"] = io.SerialNumber
			} else {
				tags["serial"] = "unknown"
			}
		}

		fields := map[string]interface{}{
			"reads":            io.ReadCount,
			"writes":           io.WriteCount,
			"read_bytes":       io.ReadBytes,
			"write_bytes":      io.WriteBytes,
			"read_time":        io.ReadTime,
			"write_time":       io.WriteTime,
			"io_time":          io.IoTime,
			"weighted_io_time": io.WeightedIO,
			"iops_in_progress": io.IopsInProgress,
			"merged_reads":     io.MergedReadCount,
			"merged_writes":    io.MergedWriteCount,
		}
		i.collectCache = append(i.collectCache, &diskioMeasurement{name: metricName, tags: tags, fields: fields, ts: ts})
	}

	return nil
}

func (i *Input) Run() {
	diskioLogger.Infof("diskio input started")
	tick := time.NewTicker(collectCycle)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			start := time.Now()
			i.collectCache = make([]inputs.Measurement, 0)
			if err := i.Collect(); err == nil {
				inputs.FeedMeasurement(metricName, io.Metric, i.collectCache,
					&io.Option{CollectCost: time.Since(start)})
			} else {
				diskioLogger.Error(err)
			}
		case <-datakit.Exit.Wait():
			diskioLogger.Infof("diskio input exit")
			return
		}
	}
}

func (i *Input) diskName(devName string) (string, []string) {
	di, err := i.diskInfo(devName)

	devLinks := strings.Split(di["DEVLINKS"], " ")
	for i, devLink := range devLinks {
		devLinks[i] = strings.TrimPrefix(devLink, "/dev/")
	}

	if err != nil {
		diskioLogger.Warnf("Error gathering disk info: %s", err)
		return devName, devLinks
	}

	// diskInfo empty
	if len(i.NameTemplates) == 0 || len(di) == 0 {
		return devName, devLinks
	}

	// render name templates
	for _, nt := range i.NameTemplates {
		miss := false
		name := varRegex.ReplaceAllStringFunc(nt, func(sub string) string {
			sub = sub[1:]
			if sub[0] == '{' {
				sub = sub[1 : len(sub)-1]
			}
			if v, ok := di[sub]; ok {
				return v
			}
			if sub == "device" {
				return devName
			}
			miss = true
			return ""
		})
		if !miss { // must match all variables
			return name, devLinks
		}
	}
	return devName, devLinks
}

func (i *Input) diskTags(devName string) map[string]string {
	if len(i.DeviceTags) == 0 {
		return nil
	}

	di, err := i.diskInfo(devName)
	if err != nil {
		diskioLogger.Warnf("Error gathering disk info: %s", err)
		return nil
	}

	tags := map[string]string{}
	for _, dt := range i.DeviceTags {
		if v, ok := di[dt]; ok {
			tags[dt] = v
		}
	}

	return tags
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{diskIO: disk.IOCounters}
	})
}
