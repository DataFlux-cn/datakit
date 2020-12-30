package cshark

import (
	"testing"
	"fmt"
	"time"
	// "gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var msg = `
{
    "device": ["lo"],
    "sync": true,
    "stream": {
        "duration": "10s",
        "protocol": "http",
        "count": 10,
        "port": ["8080"],
        "srcIP": ["127.0.0.1"],
        "dstIP": ["127.0.0.1"]
    }
}
`
func TestRun(t *testing.T) {
	t.Run("case-push-data", func(t *testing.T) {
		datakit.InstallDir = "."
		datakit.DataDir = "."
		datakit.OutputFile = "metrics.txt"

		s := &Shark{}
		s.Interval = "3s"
		s.TsharkPath = "/usr/bin/tshark"

		go s.Run()

		time.Sleep(time.Second*10)

		if err := SendCmdOpt(msg); err != nil {
			fmt.Println("err", err)
		}

		time.Sleep(10*time.Second)

		t.Log("ok")
	})
}