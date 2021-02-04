package tailf

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "tailf"

	sampleCfg = `
[[inputs.tailf]]
    # glob logfiles
    # required
    logfiles = ["/usr/local/cloudcare/dataflux/datakit/*.txt"]

    # glob filteer
    ignore = [""]

    # required
    source = "abc"

    # add service tag, if it's empty, use $source.
    service = ""

    # grok pipeline script path
    pipeline = ""

    # read file from beginning
    # if from_begin was false, off auto discovery file
    from_beginning = false

    ## characters are replaced using the unicode replacement character
    ## When set to the empty string the data is not decoded to text.
    ## ex: character_encoding = "utf-8"
    ##     character_encoding = "utf-16le"
    ##     character_encoding = "utf-16le"
    ##     character_encoding = "gbk"
    ##     character_encoding = "gb18030"
    ##     character_encoding = ""
    # character_encoding = ""

    ## The pattern should be a regexp
    ## Note the use of '''XXX'''
    # match = '''^\d{4}-\d{2}-\d{2}.*'''

    # [inputs.tailf.tags]
    # tags1 = "value1"
`
)

const (
	findNewFileInterval    = time.Second * 10
	checkFileExistInterval = time.Minute * 10
)

// var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return NewTailf(inputName, "log", sampleCfg, nil)
	})
}
