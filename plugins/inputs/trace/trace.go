package trace

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	traceConfigSample = `
#[inputs.trace]
#	[inputs.trace.tags]
#		tag1 = "tag1"
#		tag2 = "tag2"
#		tag3 = "tag3"
`
	log *logger.Logger
)

var gTags map[string]string

type Trace struct {
	Tags map[string]string
}

func (_ *Trace) Catalog() string {
	return "trace"
}

func (_ *Trace) SampleConfig() string {
	return traceConfigSample
}

func (t *Trace) Run() {
	log = logger.SLogger("trace")
	log.Infof("trace input started...")
	gTags = t.Tags
}

func init() {
	inputs.Add("trace", func() inputs.Input {
		t := &Trace{}
		return t
	})
}
