package puppetagent

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	yaml "gopkg.in/yaml.v2"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "puppetagent"

	defaultMeasurement = "puppetagent"

	sampleCfg = `
# [inputs.puppetagent]
# 	# puppetagent location of lastrunfile
# 	location = "/opt/puppetlabs/puppet/cache/state/last_run_summary.yaml"
# 	
# 	# [inputs.puppetagent.tags]
# 	# tags1 = "tags1"
`
)

var (
	l *logger.Logger

	testAssert bool
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &PuppetAgent{}
	})
}

type PuppetAgent struct {
	Location string            `toml:"location"`
	Tags     map[string]string `toml:"tags"`
	watcher  *fsnotify.Watcher
}

func (_ *PuppetAgent) SampleConfig() string {
	return sampleCfg
}

func (_ *PuppetAgent) Catalog() string {
	return inputName
}

func (pa *PuppetAgent) Run() {
	l = logger.SLogger(inputName)

	if pa.initcfg() {
		return
	}

	defer pa.watcher.Close()
	pa.do()
}

func (pa *PuppetAgent) initcfg() bool {
	var err error

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		if _, err = os.Stat(pa.Location); err != nil {
			goto _NEXT
		}
		pa.watcher, err = fsnotify.NewWatcher()
		if err != nil {
			goto _NEXT
		}
		err = pa.watcher.Add(pa.Location)
		if err != nil {
			goto _NEXT
		}

		break
	_NEXT:
		l.Error(err)
		time.Sleep(time.Second)
	}

	pa.Tags["location"] = pa.Location

	return false
}

func (pa *PuppetAgent) do() {
	l.Infof("puppetagent input started...")

	for {
		select {

		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

		case event, ok := <-pa.watcher.Events:
			if !ok {
				l.Warn("notfound watcher event")
				continue
			}

			if event.Op&fsnotify.Write == fsnotify.Write {

				data, err := buildPoint(pa.Location, pa.Tags)
				if err != nil {
					l.Error(err)
					continue
				}
				if testAssert {
					fmt.Printf("data: %s\n", string(data))
					continue
				}
				if err := io.Feed(data, io.Metric); err != nil {
					l.Error(err)
					continue
				}
				l.Debugf("feed %d bytes to io ok", len(data))
			}

			if event.Op&fsnotify.Remove == fsnotify.Remove {
				_ = pa.watcher.Remove(pa.Location)
				if err := pa.watcher.Add(pa.Location); err != nil {
					l.Errorf(err.Error())
					time.Sleep(time.Second)
				}
			}

		case err, ok := <-pa.watcher.Errors:
			if !ok {
				l.Warn(err.Error())
			}
		}
	}
}

type State struct {
	Version   version
	Events    event
	Resources resource
	Changes   change
	Timer     timer
}

type version struct {
	ConfigString string `yaml:"config"`
	Puppet       string `yaml:"puppet"`
}

type resource struct {
	Changed          int64 `yaml:"changed"`
	CorrectiveChange int64 `yaml:"corrective_change"`
	Failed           int64 `yaml:"failed"`
	FailedToRestart  int64 `yaml:"failed_to_restart"`
	OutOfSync        int64 `yaml:"out_of_sync"`
	Restarted        int64 `yaml:"restarted"`
	Scheduled        int64 `yaml:"scheduled"`
	Skipped          int64 `yaml:"skipped"`
	Total            int64 `yaml:"total"`
}

type change struct {
	Total int64 `yaml:"total"`
}

type event struct {
	Failure int64 `yaml:"failure"`
	Total   int64 `yaml:"total"`
	Success int64 `yaml:"success"`
}

type timer struct {
	FactGeneration float64 `yaml:"fact_generation"`
	Plugin_sync    float64 `yaml:"plugin_sync"`
	Total          float64 `yaml:"total"`
	LastRun        int64   `yaml:"last_run"`
}

func buildPoint(fn string, tags map[string]string) ([]byte, error) {
	fh, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}

	if len(fn) == 0 {
		return nil, fmt.Errorf("location file is empty")
	}

	var puppetState State

	err = yaml.Unmarshal(fh, &puppetState)
	if err != nil {
		return nil, err
	}

	e := reflect.ValueOf(&puppetState).Elem()

	fields := make(map[string]interface{})

	for tLevelFNum := 0; tLevelFNum < e.NumField(); tLevelFNum++ {
		name := e.Type().Field(tLevelFNum).Name
		nameNumField := e.FieldByName(name).NumField()

		for sLevelFNum := 0; sLevelFNum < nameNumField; sLevelFNum++ {
			sName := e.FieldByName(name).Type().Field(sLevelFNum).Name
			sValue := e.FieldByName(name).Field(sLevelFNum).Interface()

			lname := strings.ToLower(name)
			lsName := strings.ToLower(sName)
			fields[fmt.Sprintf("%s_%s", lname, lsName)] = sValue
		}
	}

	return io.MakeMetric(defaultMeasurement, tags, fields)
}
