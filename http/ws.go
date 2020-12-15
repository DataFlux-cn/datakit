package http

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	tgi "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/telegraf_inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/wsmsg"
)

var (
	cli   *wscli
	wsurl *url.URL
)

type wscli struct {
	c    *websocket.Conn
	id   string
	exit chan interface{}
}

func StartWS() {
	l.Infof("ws start")
	wsurl = datakit.Cfg.MainCfg.DataWay.BuildWSURL(datakit.Cfg.MainCfg)

	cli = &wscli{
		id:   datakit.Cfg.MainCfg.UUID,
		exit: make(chan interface{}),
	}

	cli.tryConnect(wsurl.String())

	go func() {
		cli.waitMsg()
	}()

	go func() {
		cli.sendHeartbeat()
	}()
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("start ws exit")
			return
		case <-cli.exit:
			cli.reset()
		}
	}

}

func (wc *wscli) tryConnect(wsurl string) {

	for {
		c, resp, err := websocket.DefaultDialer.Dial(wsurl, nil)
		if err != nil {
			l.Errorf("websocket.DefaultDialer.Dial(): %s", err.Error())
			time.Sleep(time.Second * 3)
			continue
		}
		_ = resp

		wc.c = c
		l.Infof("ws ready")
		break
	}
}

func (wc *wscli) reset() {
	l.Infof("ws reset run")
	wc.c.Close()
	wc.tryConnect(wsurl.String())
	wc.exit = make(chan interface{})

	go func() {
		cli.waitMsg()
	}()
	go func() {
		cli.sendHeartbeat()
	}()

}

func (wc *wscli) waitMsg() {

	var f rtpanic.RecoverCallback
	f = func(trace []byte, err error) {
		for {

			defer rtpanic.Recover(f, nil)

			if trace != nil {
				l.Warn("recover ok: %s err:", string(trace), err.Error())

			}

			_, resp, err := wc.c.ReadMessage()
			if err != nil {
				l.Error(err)
				wc.exit <- make(chan interface{})
				return
			}

			wm, err := wsmsg.ParseWrapMsg(resp)
			if err != nil {
				l.Error("msg.ParseWrapMsg(): %s", err.Error())
				continue
			}
			l.Infof("dk hand message %s", wm)

			if err := wc.handle(wm); err != nil {
				wc.exit <- make(chan interface{})
				return
			}
		}
	}

	f(nil, nil)
}

func (wc *wscli) sendHeartbeat() {
	m := wsmsg.MsgDatakitHeartbeat{UUID: wc.id}

	var f rtpanic.RecoverCallback
	f = func(trace []byte, _ error) {

		defer rtpanic.Recover(f, nil)
		if trace != nil {
			l.Warn("recover ok: %s", string(trace))
		}
		heartbeatTime := datakit.Cfg.MainCfg.DataWay.Heartbeat
		if heartbeatTime == "" {
			heartbeatTime = "30s"
		}
		heart, err := time.ParseDuration(heartbeatTime)

		if err != nil {
			l.Error(err)
		}
		tick := time.NewTicker(heart)
		defer tick.Stop()

		for {
			wm, err := wsmsg.BuildMsg(m)
			if err != nil {
				l.Error(err)
			}

			err = wc.sendText(wm)
			if err != nil {
				wc.exit <- make(chan interface{})
				return
			}

			select {
			case <-tick.C:
			case <-datakit.Exit.Wait():
				l.Info("ws heartbeat exit")
				return
			}
		}
	}

	f(nil, nil)
}

func (wc *wscli) sendText(wm *wsmsg.WrapMsg) error {
	wm.Dest = []string{datakit.Cfg.MainCfg.UUID}
	j, err := json.Marshal(wm)
	if err != nil {
		l.Error(err)
		return err
	}

	if err := wc.c.WriteMessage(websocket.TextMessage, j); err != nil {
		l.Errorf("WriteMessage(): %s", err.Error())

		return err
	}

	return nil
}

func (wc *wscli) handle(wm *wsmsg.WrapMsg) error {
	switch wm.Type {
	case wsmsg.MTypeOnline:
		wc.OnlineInfo(wm)
	case wsmsg.MTypeGetInput:
		wc.GetInputsConfig(wm)
	case wsmsg.MTypeGetEnableInput:
		wc.GetEnableInputsConfig(wm)
	case wsmsg.MTypeDisableInput:
		wc.DisableInput(wm)
	case wsmsg.MTypeSetInput:
		wc.SetInput(wm)
	case wsmsg.MTypeTestInput:
		wc.TestInput(wm)
	case wsmsg.MTypeEnableInput:
		wc.EnableInputs(wm)

	case wsmsg.MTypeReload:
		wc.Reload(wm)

	//case wsmsg.MTypeHeartbeat:
	default:
		wc.SetMessage(wm, "error", fmt.Errorf("unknow type %s ", wm.Type).Error())

	}
	return wc.sendText(wm)
}

func (wc *wscli) EnableInputs(wm *wsmsg.WrapMsg) {
	var names wsmsg.MsgGetInputConfig
	err := names.Handle(wm)
	if err != nil {
		wc.SetMessage(wm, "bad_request", fmt.Sprintf("parse config err:%s",err.Error()))
		return
	}
	isExist := false
	if err := filepath.Walk(datakit.ConfdDir, func(fp string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}

		if !strings.HasSuffix(f.Name(), ".off") {
			return nil
		}

		for _,name := range names.Names {
			fileName := strings.TrimSuffix(filepath.Base(fp), path.Ext(fp))
			if fileName == fmt.Sprintf("%s.conf", name) {
				isExist = true
				err = os.Rename(fp,filepath.Join(filepath.Dir(fp),fileName))
				return err
			}
		}
		return nil
	}); err != nil {
		wc.SetMessage(wm,"error",fmt.Sprintf("os walk err:%s",err.Error()))
		return
	}
	if isExist {
		wc.SetMessage(wm,"ok","")
	}else {
		wc.SetMessage(wm,"error","input not exist disabled config")
	}

}



func (wc *wscli) TestInput(wm *wsmsg.WrapMsg) {
	var configs wsmsg.MsgSetInputConfig
	err := configs.Handle(wm)
	if err != nil {
		wc.SetMessage(wm, "error", err.Error())
		return
	}
	var returnMap = map[string]string{}
	for k, v := range configs.Configs {
		data, err := base64.StdEncoding.DecodeString(v["toml"])
		if err != nil {
			wc.SetMessage(wm, "error", err.Error())
			return
		}
		if creator, ok := inputs.Inputs[k]; ok {
			tbl, err := toml.Parse(data)
			if err != nil {
				l.Error(err)
			}
			for _, node := range tbl.Fields {
				stbl, _ := node.(*ast.Table)
				for _, d := range stbl.Fields {
					inputList, err := config.TryUnmarshal(d, k, creator)
					if err != nil {
						wc.SetMessage(wm, "error", err.Error())
						return
					}
					if len(inputList) > 0 {
						result, err := inputList[0].Test()
						if err != nil {
							wc.SetMessage(wm, "error", err.Error())
							return
						}
						returnMap[k] = base64.StdEncoding.EncodeToString(result.Result)
					}
				}
			}
			continue
		}

		if _, ok := tgi.TelegrafInputs[k]; ok {
			result, err := inputs.TestTelegrafInput(data)
			if err != nil {
				wc.SetMessage(wm, "error", err.Error())
				return
			}

			returnMap[k] = base64.StdEncoding.EncodeToString(result.Result)
			continue
		}

		wc.SetMessage(wm, "error", fmt.Sprintf("input %s not available", k))
		return

	}
	wc.SetMessage(wm, "ok", returnMap)
}

func (wc *wscli) Reload(wm *wsmsg.WrapMsg) {
	err := ReloadDatakit()
	if err != nil {
		l.Errorf("reload err:%s", err.Error())
		wc.SetMessage(wm, "error", err.Error())
	}
	go func() {
		RestartHttpServer()
		l.Info("reload HTTP server ok")
	}()
	wc.SetMessage(wm, "ok", "")

}

func checkConfig(listInput []string) bool {
	var Set = map[string]bool{}
	for _,inp := range listInput {
		Set[inp] = true
	}
	if len(Set) != len(listInput) {
		return false
	}
	return true
}


func parseConf(conf string, name string) (listMd5 []string,catelog string, err error) {
	data, err := base64.StdEncoding.DecodeString(conf)
	if err != nil {
		return
	}
	tbl, err := toml.Parse(data)
	if err != nil {
		return
	}

	for _, node := range tbl.Fields {
		stbl, _ := node.(*ast.Table)
		for _, sstbl := range stbl.Fields {
			for _, tb := range sstbl.([]*ast.Table) {
				if name != tb.Name {
					err = fmt.Errorf("input: %s parse config err",name)
					return
				}
				if creator, ok := inputs.Inputs[name]; ok {
					inp := creator()
					err = toml.UnmarshalTable(tb, inp)
					listMd5 = append(listMd5, inputs.SetInputsMD5(name,inp))
					catelog = inp.Catalog()
					continue
				}
				if creator, ok := tgi.TelegrafInputs[name]; ok {
					catelog = creator.Catalog
					cr := reflect.ValueOf(creator).Elem().Interface()
					err = toml.UnmarshalTable(tb, cr.(tgi.TelegrafInput).Input)
					listMd5 = append(listMd5, inputs.SetInputsMD5(name,cr.(tgi.TelegrafInput)))
					continue
				}
				err = fmt.Errorf("input:%s is not available", name)
				return
			}
		}
	}
	return
}


func (wc *wscli) parseTomlToFile(tomlStr, name string) error {
	listInput,catalog, err := parseConf(tomlStr, name)
	if err != nil {
		return err
	}
	if !checkConfig(listInput){
		return fmt.Errorf("cannot set same config")
	}
	inputPath := filepath.Join(datakit.ConfdDir, catalog, fmt.Sprintf("%s.conf",name))
	n,cfg := inputs.InputEnabled(name)
	if n > 0 {
		for _,fp := range cfg {
			fileName := strings.TrimSuffix(filepath.Base(fp), path.Ext(fp))
			if fileName == name {
				inputPath = fp
				break
			}
		}
	}
	if err = wc.WriteFile(tomlStr,inputPath); err != nil {
		return err
	}
	return nil
}

func (wc *wscli) SetInput(wm *wsmsg.WrapMsg) {
	var configs wsmsg.MsgSetInputConfig
	err := configs.Handle(wm)
	if err != nil {
		wc.SetMessage(wm, "bad_request", fmt.Sprintf("parse config err:%s",err.Error()))
		return
	}
	for k, v := range configs.Configs {
		if err := wc.parseTomlToFile(v["toml"],k); err != nil {
			wc.SetMessage(wm, "error", err.Error())
			return
		}

	}
	wc.SetMessage(wm, "ok", "")
}

func (wc *wscli) WriteFile(tomlStr,cfgPath string) error{
	data, err := base64.StdEncoding.DecodeString(tomlStr)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(cfgPath, data, 0660)
	return err
}

func (wc *wscli) DisableInput(wm *wsmsg.WrapMsg) {
	var names wsmsg.MsgGetInputConfig
	err := names.Handle(wm)
	if err != nil {
		wc.SetMessage(wm, "bad_request", fmt.Sprintf("parse config err:%s",err.Error()))
		return
	}
	for _,name := range names.Names {
		n,cfg := inputs.InputEnabled(name)
		if n > 0 {
			for _,fp := range cfg {
				os.Rename(fp,fmt.Sprintf("%s.off",fp))
			}
		}else {
			wc.SetMessage(wm, "error", fmt.Sprintf("input:%s not eanble",name))
			return
		}
	}
	wc.SetMessage(wm, "ok", "")
}

func (wc *wscli) GetEnableInputsConfig(wm *wsmsg.WrapMsg) {
	var names wsmsg.MsgGetInputConfig
	err := names.Handle(wm)
	if err != nil || len(names.Names)==0{
		wc.SetMessage(wm, "bad_request", fmt.Sprintf("params error"))
		return
	}
	var Enable []map[string]map[string]string
	for _, v := range names.Names {
		n, cfg := inputs.InputEnabled(v)
		if n > 0 {
			for _, p := range cfg {
				cfgData, err := ioutil.ReadFile(p)
				if err != nil {
					errorMessage := fmt.Sprintf("get enable config read file error path:%s", p)
					wc.SetMessage(wm, "error", errorMessage)
					return
				}
				//_, fileName := filepath.Split(p)
				fileName := strings.TrimSuffix(filepath.Base(p), path.Ext(p))
				Enable = append(Enable, map[string]map[string]string{fileName: {"toml": base64.StdEncoding.EncodeToString(cfgData)}})
			}
		} else {
			wc.SetMessage(wm, "error", fmt.Sprintf("input %s not enable", v))
			return
		}
	}
	wc.SetMessage(wm, "ok", Enable)
}

func (wc *wscli) SetMessage(wm *wsmsg.WrapMsg, code string, Message interface{}) {
	wm.Code = code
	if Message != "" {
		wm.B64Data = ToBase64(Message)
	} else {
		wm.B64Data = ""
	}

}

func (wc *wscli) GetInputsConfig(wm *wsmsg.WrapMsg) {
	var names wsmsg.MsgGetInputConfig
	err := names.Handle(wm)
	if err != nil {
		errMessage := fmt.Sprintf("GetInputsConfig %s params error", wm)
		l.Error(errMessage)
		wc.SetMessage(wm, "error", errMessage)
		return
	}
	var data []map[string]map[string]string

	for _, v := range names.Names {
		sample, err := inputs.GetSample(v)
		if err != nil {
			errMessage := fmt.Sprintf("get config error %s", err)
			l.Error(errMessage)
			wc.SetMessage(wm, "error", errMessage)
			return
		}
		data = append(data, map[string]map[string]string{v: {"toml": base64.StdEncoding.EncodeToString([]byte(sample))}})
	}
	wc.SetMessage(wm, "ok", data)
}

func (wc *wscli) OnlineInfo(wm *wsmsg.WrapMsg) {
	m := wsmsg.MsgDatakitOnline{
		UUID:            wc.id,
		Name:            datakit.Cfg.MainCfg.Name,
		Version:         git.Version,
		OS:              runtime.GOOS,
		Arch:            runtime.GOARCH,
		Heartbeat:       datakit.Cfg.MainCfg.DataWay.Heartbeat,
		InputInfo:       map[string]interface{}{},
	}
	m.InputInfo["availableInputs"] = GetAvailableInputs()
	m.InputInfo["enabledInputs"] = GetEnableInputs()
	state,err := io.GetStats()
	if err != nil {
		l.Errorf("get state err:%s",err.Error())
		state = []*io.InputsStat{}
	}
	m.InputInfo["state"] = state

	wc.SetMessage(wm, "ok", m)
}

func ToBase64(wm interface{}) string {
	body, err := json.Marshal(wm)
	if err != nil {
		l.Errorf("%s toBase64 err:%s", wm, err)
	}
	return base64.StdEncoding.EncodeToString(body)
}

func GetAvailableInputs() []string {
	var AvailableInputs []string
	for k, _ := range inputs.Inputs {
		AvailableInputs = append(AvailableInputs, k)
	}
	for k, _ := range tgi.TelegrafInputs {
		AvailableInputs = append(AvailableInputs, k)
	}
	return AvailableInputs
}

func GetEnableInputs()(Enable []string)  {
	for k, _ := range inputs.Inputs {
		n, _ := inputs.InputEnabled(k)
		if n > 0 {
			Enable = append(Enable,k)
		}
	}

	for k, _ := range tgi.TelegrafInputs {
		n, _ := inputs.InputEnabled(k)
		if n > 0 {
			Enable = append(Enable,k)
		}
	}
	return Enable
}
