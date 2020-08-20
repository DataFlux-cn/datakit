package inputs

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/toml/ast"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

type Input interface {
	Catalog() string
	Run()
	SampleConfig() string

	// add more...
}

type HttpRegInput interface {
	Input
	RegHttpHandler()
}

type Creator func() Input

var (
	Inputs     = map[string]Creator{}
	inputInfos = map[string][]*inputInfo{}

	l *logger.Logger = logger.DefaultSLogger("inputs")

	panicInputs = map[string]int{}
	mtx         = sync.RWMutex{}
)

func Add(name string, creator Creator) {
	if _, ok := Inputs[name]; ok {
		panic(fmt.Sprintf("inputs %s exist(from datakit)", name))
	}

	if _, ok := TelegrafInputs[name]; ok {
		panic(fmt.Sprintf("inputs %s exist(from telegraf)", name))
	}

	l.Infof("add input %s", name)
	Inputs[name] = creator
}

type inputInfo struct {
	input Input
	ti    *TelegrafInput
	cfg   string
}

func (ii *inputInfo) Run() {
	if ii.input == nil {
		return
	}

	switch ii.input.(type) {
	case Input:
		ii.input.Run()
	default:
		l.Errorf("invalid input type, cfg: %s", ii.cfg)
	}
}

func AddInput(name string, input Input, table *ast.Table, fp string) error {

	mtx.Lock()
	defer mtx.Unlock()

	var dur time.Duration
	var err error
	if node, ok := table.Fields["interval"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				dur, err = time.ParseDuration(str.Value)
				if err != nil {
					l.Errorf("parse duration(%s) from %s failed: %s", str.Value, name, err.Error())
					return err
				}
			}
		}
	}

	l.Debugf("try set MaxLifeCheckInterval to %v from %s...", dur, name)
	if datakit.MaxLifeCheckInterval+5*time.Second < dur { // use the max interval from all inputs
		datakit.MaxLifeCheckInterval = dur
		l.Debugf("set MaxLifeCheckInterval to %v from %s", dur, name)
	}

	inputInfos[name] = append(inputInfos[name], &inputInfo{input: input, cfg: fp})

	return nil
}

func InputInstaces(name string) int {
	mtx.RLock()
	defer mtx.RUnlock()

	if arr, ok := inputInfos[name]; ok {
		return len(arr)
	}
	return 0
}

func ResetInputs() {

	mtx.Lock()
	defer mtx.Unlock()
	inputInfos = map[string][]*inputInfo{}
}

func AddSelf(i Input) {

	mtx.Lock()
	defer mtx.Unlock()

	inputInfos["self"] = append(inputInfos["self"], &inputInfo{input: i, cfg: "no config for `self' input"})
}

func AddTelegrafInput(name, fp string) {

	mtx.Lock()
	defer mtx.Unlock()

	l.Debugf("add telegraf input %s from %s", name, fp)
	inputInfos[name] = append(inputInfos[name],
		&inputInfo{input: nil, /* not used */
			ti:  nil, /*not used*/
			cfg: fp})
}

func StartTelegraf() error {

	if !HaveTelegrafInputs() {
		l.Info("no telegraf inputs enabled")
		return nil
	}

	datakit.WG.Add(1)
	go func() {
		defer datakit.WG.Done()
		_ = doStartTelegraf()

		l.Info("telegraf process exit ok")
	}()

	return nil
}

func RunInputs() error {

	l = logger.SLogger("inputs")
	mtx.RLock()
	defer mtx.RUnlock()

	for name, arr := range inputInfos {
		for _, ii := range arr {
			if ii.input == nil {
				l.Debugf("skip non-datakit-input %s", name)
				continue
			}
			switch inp := ii.input.(type) {
			case HttpRegInput:
				inp.RegHttpHandler()
			}

			l.Infof("starting input %s ...", name)
			datakit.WG.Add(1)
			go func(name string, ii *inputInfo) {
				defer datakit.WG.Done()
				protectRunningInput(name, ii)
				l.Infof("input %s exited", name)
			}(name, ii)
		}
	}
	return nil
}

var (
	MaxCrash = 6
)

func protectRunningInput(name string, ii *inputInfo) {
	var f rtpanic.RecoverCallback
	crashTime := []string{}

	f = func(trace []byte, err error) {

		defer rtpanic.Recover(f, nil)

		if trace != nil {
			l.Warnf("input %s panic err: %v", name, err)
			l.Warnf("input %s panic trace:\n%s", name, string(trace))

			crashTime = append(crashTime, fmt.Sprintf("%v", time.Now()))
			addPanic(name)

			if len(crashTime) >= MaxCrash {
				l.Warnf("input %s crash %d times(at %+#v), exit now.",
					name, len(crashTime), strings.Join(crashTime, ","))
				return
			}
		}

		ii.Run()
	}

	f(nil, nil)
}

func InputEnabled(name string) (int, []string) {
	mtx.RLock()
	defer mtx.RUnlock()
	arr, ok := inputInfos[name]
	if !ok {
		return 0, nil
	}

	cfgs := []string{}
	for _, i := range arr {
		cfgs = append(cfgs, i.cfg)
	}

	return len(arr), cfgs
}

func GetPanicCnt(name string) int {
	mtx.RLock()
	defer mtx.RUnlock()

	return panicInputs[name]
}

func addPanic(name string) {
	mtx.Lock()
	defer mtx.Unlock()

	panicInputs[name]++
}
