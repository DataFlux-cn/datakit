// +build !solaris

package tailf

import (
	"bytes"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hpcloud/tail"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
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

    # read file from beginning
    # if from_begin was false, off auto discovery file
    from_beginning = false

    # required
    source = ""

    # [inputs.tailf.tags]
    # tags1 = "value1"
`

	updateFileListInterval   = time.Second * 3
	checkFileIsExistInterval = time.Minute * 20
	metricsFeedInterval      = time.Second * 5
	metricsFeedCount         = 10
)

var l = logger.DefaultSLogger(inputName)

type Tailf struct {
	LogFiles          []string          `toml:"logfiles"`
	Ignore            []string          `toml:"ignore"`
	FromBeginning     bool              `toml:"from_beginning"`
	Source            string            `toml:"source"`
	CharacterEncoding string            `toml:"character_encoding"`
	Tags              map[string]string `toml:"tags"`

	tailerConf tail.Config

	runningFileList sync.Map
	wg              sync.WaitGroup
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Tailf{}
	})
}

func (*Tailf) Catalog() string {
	return "log"
}

func (*Tailf) SampleConfig() string {
	return sampleCfg
}

func (t *Tailf) Run() {
	l = logger.SLogger(inputName)

	if t.loadcfg() {
		return
	}

	l.Infof("tailf input started.")

	ticker := time.NewTicker(updateFileListInterval)
	defer ticker.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Infof("waiting for all tailers to exit")
			t.wg.Wait()
			l.Info("exit")
			return

		case <-ticker.C:
			fileList := getFileList(t.LogFiles, t.Ignore)
			for _, f := range fileList {
				if _, ok := t.runningFileList.Load(f); !ok {
					t.runningFileList.Store(f, nil)
					l.Debugf("start tail, %s", f)

					t.wg.Add(1)
					go t.startTail(f)
				} else {
					l.Debugf("file %s already tailing now", f)
				}
			}

			if t.FromBeginning {
				// off auto discovery file
				// ticker was unreachable
				ticker.Stop()
			}
		}
	}
}

func (t *Tailf) loadcfg() bool {
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		if t.Source == "" {
			l.Errorf("tailf source was empty")
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	var seek *tail.SeekInfo
	if !t.FromBeginning {
		seek = &tail.SeekInfo{
			Whence: 2, // seek is 2
			Offset: 0,
		}
	}

	t.tailerConf = tail.Config{
		ReOpen:    true,
		Follow:    true,
		Location:  seek,
		MustExist: true,
		Poll:      false, // default watch method is "inotify"
		Pipe:      false,
		Logger:    tail.DiscardingLogger,
	}
	t.runningFileList = sync.Map{}
	t.wg = sync.WaitGroup{}
	return false
}

func (t *Tailf) startTail(file string) {
	defer t.wg.Done()

	err := t.getLines(file)
	// file is not exist or datakit exit
	if err == nil {
		t.runningFileList.Delete(file)
		l.Debugf("file %s is ending", file)
	}
}

func (t *Tailf) getLines(file string) error {
	tailer, err := tail.TailFile(file, t.tailerConf)
	if err != nil {
		l.Error("build tailer, %s", err)
		return err
	}
	defer tailer.Cleanup()

	tags := make(map[string]string)
	for k, v := range t.Tags {
		tags[k] = v
	}
	tags["filename"] = file

	feedTicker := time.NewTicker(metricsFeedInterval)
	defer feedTicker.Stop()

	checkTicker := time.NewTicker(checkFileIsExistInterval)
	defer checkTicker.Stop()

	var (
		cacheBuffer bytes.Buffer
		textLine    bytes.Buffer
	)
	count := 0
	for {
		select {
		case <-datakit.Exit.Wait():
			return nil

		case line := <-tailer.Lines:
			if line.Err != nil {
				l.Error("tailer lines, %s", err)
			}

			text := strings.TrimRight(line.Text, "\r")
			if textLine.Len() == 0 {
				textLine.WriteString(text)
				continue
			}

			// 如果行首含有空格、换行符、制表符、换行符，则将其添加到上一条 log 的末尾
			if strings.HasPrefix(text, " ") ||
				strings.HasPrefix(text, "\r") ||
				strings.HasPrefix(text, "\t") ||
				strings.HasPrefix(text, "\n") ||
				len(text) == 0 {

				textLine.WriteString("\n")
				textLine.WriteString(text)
				continue
			}

			fields := map[string]interface{}{"__content": textLine.String()}
			textLine.Reset()
			textLine.WriteString(text)

			data, err := io.MakeMetric(t.Source, tags, fields, time.Now())
			if err != nil {
				l.Error(err)
				continue
			}

			cacheBuffer.Write(data)
			cacheBuffer.WriteString("\n")
			count++

			if count >= metricsFeedCount {
				if err := io.NamedFeed(cacheBuffer.Bytes(), io.Logging, inputName); err != nil {
					l.Error(err)
				}
				count = 0
				// not use buffer.Reset()
				cacheBuffer = bytes.Buffer{}
			}

		case <-feedTicker.C:
			if textLine.Len() != 0 {
				fields := map[string]interface{}{"__content": textLine.String()}
				textLine.Reset()
				data, err := io.MakeMetric(t.Source, tags, fields, time.Now())
				if err != nil {
					l.Error(err)
					continue
				}
				cacheBuffer.Write(data)
				count++
			}
			if count > 0 {
				if err := io.NamedFeed(cacheBuffer.Bytes(), io.Logging, inputName); err != nil {
					l.Error(err)
				}
				count = 0
				cacheBuffer = bytes.Buffer{}
			}

		case <-checkTicker.C:
			_, statErr := os.Lstat(file)
			if os.IsNotExist(statErr) {
				l.Warnf("check file %s is not exist", file)
				return nil
			}
		}
	}
}
