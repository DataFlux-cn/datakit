package hostobject

const (
	InputName = "hostobject"
	InputCat  = "host"

	SampleConfig = `
[inputs.hostobject]

#pipeline = '' # optional

## Datakit does not collect network virtual interfaces under the linux system.
## Setting enable_net_virtual_interfaces to true will collect network virtual interfaces stats for linux.
# enable_net_virtual_interfaces = true

## Ignore mount points by filesystem type. Default ingore following FS types
# ignore_fs = ["tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "autofs", "squashfs", "aufs"]


[inputs.hostobject.tags] # (optional) custom tags
#  cloud_provider = "aliyun" # aliyun/tencent/aws
#  tag2 = "value2"
`
	pipelineSample = ``
)
