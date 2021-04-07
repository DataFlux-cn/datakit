package tailf

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "tailf"

	sampleCfg = `
[[inputs.tailf]]
    # required, glob logfiles
    logfiles = ["/path/to/your/file.log"]

    # glob filteer
    ignore = [""]

    # your logging source, if it's empty, use 'default'
    source = ""

    # add service tag, if it's empty, use $source.
    service = ""

    # grok pipeline script path
    pipeline = ""

    # optional status: 
    #   "emerg","alert","critical","error","warning","info","debug","OK"
    ignore_status = []

    # read file from beginning
    # if from_begin was false, off auto discovery file
    from_beginning = false

    # optional encodings:
    #    "utf-8", "utf-16le", "utf-16le", "gbk", "gb18030" or ""
    character_encoding = ""

    # The pattern should be a regexp. Note the use of '''this regexp'''
    # regexp link: https://golang.org/pkg/regexp/syntax/#hdr-Syntax
    match = '''^\S'''

    [inputs.tailf.tags]
    # tags1 = "value1"
`
)

type Tailf struct {
	LogFiles           []string          `toml:"logfiles"`
	Ignore             []string          `toml:"ignore"`
	Source             string            `toml:"source"`
	Service            string            `toml:"service"`
	Pipeline           string            `toml:"pipeline"`
	DeprecatedPipeline string            `toml:"pipeline_path"`
	IgnoreStatus       []string          `toml:"ignore_status"`
	FromBeginning      bool              `toml:"from_beginning"`
	CharacterEncoding  string            `toml:"character_encoding"`
	Match              string            `toml:"match"`
	Tags               map[string]string `toml:"tags"`

	log *logger.Logger
}
