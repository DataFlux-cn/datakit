package cmds

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/term"

	markdown "github.com/MichaelMure/go-term-markdown"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
)

func CMDMonitor(intervalStr, addrStr string, verbose bool) {
	addr := "http://localhost:9529/stats"
	if addrStr != "" {
		addr = "http://" + addrStr + "/stats"
	}

	interval := 3 * time.Second
	if intervalStr != "" {
		if du, err := time.ParseDuration(intervalStr); err == nil {
			if du >= time.Second {
				interval = du // only accept interval >= 1s
			}
		}
	}

	tick := time.NewTicker(interval)
	for {
		select {
		case <-tick.C:
			fmt.Print("\033[H\033[2J") // clean screen

			x, err := doCMDMonitor(addr, verbose)
			if err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println(string(x))
				fmt.Printf("(Refresh at %s)Press ctrl+c to exit.\n", interval)
			}
		}
	}
}

func doCMDMonitor(url string, verbose bool) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("%s", string(body))
	}

	ds := dkhttp.DatakitStats{
		DisableMonofont: true,
	}
	if err := json.Unmarshal(body, &ds); err != nil {
		return nil, err
	}

	l.Debugf("stats.ReloadInfo: %s", ds.ReloadInfo)

	mdtxt, err := ds.Markdown("", verbose)
	if err != nil {
		return nil, err
	}

	width := 100
	if term.IsTerminal(0) {
		w, _, err := term.GetSize(0)
		if err == nil {
			width = w
		}
	}

	if err != nil {
		return nil, err
	} else {
		if len(mdtxt) == 0 {
			return nil, fmt.Errorf("no monitor info available")
		} else {
			result := markdown.Render(string(mdtxt), width, 2)
			return result, nil
		}
	}
}
