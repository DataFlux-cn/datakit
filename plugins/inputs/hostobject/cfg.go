package hostobject

const (
	inputName = "hostobject"

	sampleConfig = `
#[inputs.hostobject]
# ##(optional) default use host name
#name = ''

# ##(optional) default is Servers
#class = 'Servers'

# ##(optional) collect interval, default is 3 miniutes
#interval = '3m'

# ##(optional) 
#pipeline = ''

# ##(optional) custom tags
#[inputs.hostobject.tags]
# key1 = 'val1'
`

	pipelineSample = ``
)
