package hostobject

const (
	InputName = "hostobject"
	InputCat  = "host"

	SampleConfig = `
[inputs.hostobject]

# ##(optional) collect interval, default is 5 miniutes
interval = '5m'

# ##(optional) 
#pipeline = ''

# ##(optional) custom tags
#[inputs.hostobject.tags]
#  key1 = "value1"
#  key2 = "value2"
`

	pipelineSample = ``
)
