package cmds

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/c-bata/go-prompt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/man"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func Man() {

	// load input-names
	for k, _ := range inputs.Inputs {
		suggestions = append(suggestions, prompt.Suggest{Text: k, Description: ""})
	}

	// TODO: add suggestions for pipeline

	c, _ := newCompleter()

	p := prompt.New(
		runMan,
		c.Complete,
		prompt.OptionTitle("man: DataKit manual query"),
		prompt.OptionPrefix("man > "),
	)

	p.Run()
}

func ExportMan(to string) error {
	if err := os.MkdirAll(to, os.ModePerm); err != nil {
		return err
	}

	for k, _ := range inputs.Inputs {
		data, err := man.BuildMarkdownManual(k)
		if err != nil {
			return err
		}

		if len(data) == 0 {
			continue
		}

		if err := ioutil.WriteFile(filepath.Join(to, k+".md"), data, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

func runMan(txt string) {
	s := strings.Join(strings.Fields(strings.TrimSpace(txt)), " ")
	if s == "" {
		return
	}

	switch s {
	case "Q", "q", "exit":
		fmt.Println("Bye!")
		os.Exit(0)
	default:
		x, err := man.BuildMarkdownManual(s)
		if err != nil {
			fmt.Printf("[E] %s\n", err.Error())
		} else {
			if len(x) == 0 {
				fmt.Printf("[E] intput %s got no manual", s)
			} else {
				result := markdown.Render(string(x), 80, 6)
				fmt.Println(string(result))
			}
		}
	}
}
