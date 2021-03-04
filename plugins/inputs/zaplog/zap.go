package zaplog

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/tailf"
)

const (
	inputName = "zaplog"

	sampleCfg = `
[[inputs.tailf]]

    logfiles = [""]  # required
    source = "<your-source>" # required

    # glob filteer
    ignore = [""]

    # add service tag, if it's empty, use $source.
    service = "" # default same as $source

    # grok pipeline script path
    pipeline = "zaplog.p"

    # read file from beginning
    # if from_begin was false, off auto discovery file
    from_beginning = false

    # optional encodings:
    #    "utf-8", "utf-16le", "utf-16le", "gbk", "gb18030" or ""
    character_encoding = ""

    # The pattern should be a regexp. Note the use of '''this regexp'''
    match = '''^\S.*'''

    [inputs.tailf.tags]
    # tags1 = "value1"
`
	pipelineCfg = `
add_pattern("zap_date", "%{YEAR}-%{MONTHNUM}-%{MONTHDAY}T%{HOUR}:%{MINUTE}:%{SECOND}\\.%{INT}Z")
add_pattern("zap_level", "(DEBUG|INFO|WARN|ERROR|FATAL)")
add_pattern("zap_mod", "%{WORD}")
add_pattern("zap_source_file", "(/?[\\w_%!$@:.,-]?/?)(\\S+)?")
add_pattern("zap_msg", "%{GREEDYDATA}")

grok(_, '%{zap_date:time}%{SPACE}%{zap_level:level}%{SPACE}%{zap_mod:module}%{SPACE}%{zap_source_file:code}%{SPACE}%{zap_msg:msg}')
default_time(time)`
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		t := tailf.NewTailf(
			inputName,
			"log",
			sampleCfg,
			map[string]string{"zaplog": pipelineCfg},
		)
		return t
	})
}
