// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskio

const sampleCfg = `
[[inputs.diskio]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'

  ## By default, gather stats for all devices including
  ## disk partitions.
  ## Setting interfaces using regular expressions will collect these expected devices.
  # devices = ['''^sda\d*''', '''^sdb\d*''', '''vd.*''']

  ## If the disk serial number is not required, please uncomment the following line.
  # skip_serial_number = true

  ## On systems which support it, device metadata can be added in the form of
  ## tags.
  ## Currently only Linux is supported via udev properties. You can view
  ## available properties for a device by running:
  ## 'udevadm info -q property -n /dev/sda'
  ## Note: Most, but not all, udev properties can be accessed this way. Properties
  ## that are currently inaccessible include DEVTYPE, DEVNAME, and DEVPATH.
  # device_tags = ["ID_FS_TYPE", "ID_FS_USAGE"]

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

[inputs.diskio.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"`
