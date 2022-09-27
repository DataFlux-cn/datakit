// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package dialtesting implement API dial testing.
// nolint:gosec
package dialtesting

import (
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	dt "gitlab.jiagouyun.com/cloudcare-tools/cliutils/dialtesting"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var ( // type assertions
	_ inputs.ReadEnv = (*Input)(nil)
	_ inputs.InputV2 = (*Input)(nil)
	g                = datakit.G("inputs_dialtesting")
)

var (
	AuthorizationType = `DIAL_TESTING`
	SignHeaders       = []string{
		`Content-MD5`,
		`Content-Type`,
		`Date`,
	}

	inputName = "dialtesting"
	l         = logger.DefaultSLogger(inputName)

	MaxFails               = 100
	MaxSendFailCount int32 = 16
)

const (
	maxCrashCnt = 6
	RegionInfo  = "region"
)

var apiTasksNum int

type Input struct {
	Region           string            `toml:"region,omitempty"`
	RegionID         string            `toml:"region_id"`
	Server           string            `toml:"server,omitempty"`
	AK               string            `toml:"ak"`
	SK               string            `toml:"sk"`
	PullInterval     string            `toml:"pull_interval,omitempty"`
	TimeOut          *datakit.Duration `toml:"time_out,omitempty"` // 单位为秒
	Workers          int               `toml:"workers,omitempty"`
	MaxSendFailCount int32             `toml:"max_send_fail_count,omitempty"` // max send fail count
	Tags             map[string]string

	semStop *cliutils.Sem // start stop signal
	cli     *http.Client  // class string

	curTasks map[string]*dialer
	pos      int64 // current largest-task-update-time
}

const sample = `
[[inputs.dialtesting]]
  # 中心任务存储的服务地址，即df_dialtesting center service。
  # 此处同时可配置成本地json 文件全路径 "file:///your/dir/json-file-name", 为task任务的json字符串。
  server = "https://dflux-dial.guance.com"

  # require，节点惟一标识ID
  region_id = "default"

  # 若server配为中心任务服务地址时，需要配置相应的ak或者sk
  ak = ""
  sk = ""

  pull_interval = "1m"

  time_out = "1m"
  workers = 6

  # 发送数据失败最大次数，根据任务的post_url进行累计，超过最大次数后，发送至该地址的拨测任务将退出
  max_send_fail_count = 16

  [inputs.dialtesting.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...`

func (*Input) SampleConfig() string {
	return sample
}

func (*Input) Catalog() string {
	return "network"
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&httpMeasurement{},
		&tcpMeasurement{},
		&icmpMeasurement{},
		&websocketMeasurement{},
	}
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (d *Input) Terminate() {
	if d.semStop != nil {
		d.semStop.Close()
	}
}

func (d *Input) Run() {
	l = logger.SLogger(inputName)

	// 根据Server配置，若为服务地址则定时拉取任务数据；
	// 若为本地json文件，则读取任务

	if d.Workers == 0 {
		d.Workers = 6
	}

	if d.MaxSendFailCount > 0 {
		MaxSendFailCount = d.MaxSendFailCount
	}

	reqURL, err := url.Parse(d.Server)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return
	}

	l.Debugf(`%+#v, %+#v`, d.cli, d.TimeOut)

	if d.TimeOut == nil {
		d.cli.Timeout = 60 * time.Second
	} else {
		d.cli.Timeout = d.TimeOut.Duration
	}

	switch reqURL.Scheme {
	case "http", "https":
		d.doServerTask() // task server

	case "file":
		d.doLocalTask(reqURL.Path)

	case "":
		d.doLocalTask(reqURL.String())

	default:
		l.Warnf(`no invalid scheme: %s`, reqURL.Scheme)
	}
}

func (d *Input) doServerTask() {
	var f rtpanic.RecoverCallback
	crashTimes := 0

	f = func(stack []byte, err error) {
		defer rtpanic.Recover(f, nil)

		if stack != nil {
			crashTimes++
			l.Warnf("[%dth]input paniced: %v", crashTimes, err)
			l.Warnf("[%dth]paniced trace: \n%s", crashTimes, string(stack))
			if crashTimes > 6 {
				return
			}
		}

		du, err := time.ParseDuration(d.PullInterval)
		if err != nil {
			l.Warnf("invalid frequency: %s, use default", d.PullInterval)
			du = time.Minute
		}
		if du > 24*time.Hour || du < time.Second*10 {
			l.Warnf("invalid frequency: %s, use default", d.PullInterval)
			du = time.Minute
		}

		tick := time.NewTicker(du)
		defer tick.Stop()

		for {
			select {
			// TODO: 调接口发送每个任务的执行情况，便于中心对任务的管理
			case <-datakit.Exit.Wait():
				l.Info("exit")
				return

			case <-d.semStop.Wait():
				l.Info("exit")
				return

			case <-tick.C:
				l.Debug("try pull tasks...")
				j, err := d.pullTask()
				if err != nil {
					l.Warnf(`pullTask: %s, ignore`, err.Error())
				} else {
					l.Debug("try dispatch tasks...")
					if err := d.dispatchTasks(j); err != nil {
						l.Warnf("dispatchTasks: %s, ignored", err.Error())
					}
				}
			}
		}
	}

	f(nil, nil)
}

func (d *Input) doLocalTask(path string) {
	data, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return
	}

	j, err := d.getLocalJSONTasks(data)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return
	}

	if err := d.dispatchTasks(j); err != nil {
		l.Errorf("dispatchTasks: %s", err.Error())
	}

	<-datakit.Exit.Wait()
}

func (d *Input) newTaskRun(t dt.Task) (*dialer, error) {
	if err := t.Init(); err != nil {
		l.Errorf(`%s`, err.Error())
		return nil, err
	}

	switch t.Class() {
	case dt.ClassHTTP:
		apiTasksNum++
	case dt.ClassHeadless:
		return nil, fmt.Errorf("headless task deprecated")
	case dt.ClassDNS:
		// TODO
	case dt.ClassTCP:
		// TODO
	case dt.ClassWebsocket:
		// TODO
	case dt.ClassICMP:
		// TODO
	case dt.ClassOther:
		// TODO
	case RegionInfo:
		break
		// no need dealwith
	default:
		l.Errorf("unknown task type")
		return nil, fmt.Errorf("invalid task type")
	}

	l.Debugf("input tags: %+#v", d.Tags)

	dialer := newDialer(t, d.Tags)
	dialer.done = d.semStop.Wait()

	func(id string) {
		g.Go(func(ctx context.Context) error {
			protectedRun(dialer)
			l.Infof("input %s exited", id)
			return nil
		})
	}(t.ID())

	return dialer, nil
}

func protectedRun(d *dialer) {
	crashcnt := 0
	var f rtpanic.RecoverCallback

	l.Infof("task %s(%s) starting...", d.task.ID(), d.class)

	f = func(trace []byte, err error) {
		defer rtpanic.Recover(f, nil)
		if trace != nil {
			l.Warnf("task %s panic: %+#v, trace: %s", d.task.ID(), err, string(trace))

			crashcnt++
			if crashcnt > maxCrashCnt {
				l.Warnf("task %s crashed %d times, exit now", d.task.ID(), crashcnt)
				return
			}
		}

		if err := d.run(); err != nil {
			l.Warnf("run: %s, ignored", err)
		}
	}

	f(nil, nil)
}

type taskPullResp struct {
	Content map[string]interface{} `json:"content"`
}

func (d *Input) dispatchTasks(j []byte) error {
	var resp taskPullResp

	if err := json.Unmarshal(j, &resp); err != nil {
		l.Errorf("json.Unmarshal: %s", err.Error())
		return err
	}

	l.Infof(`dispatching %d tasks...`, len(resp.Content))

	for k, arr := range resp.Content {
		switch k {
		case RegionInfo:
			for k, v := range arr.(map[string]interface{}) {
				switch v_ := v.(type) {
				case bool:
					if v_ {
						d.Tags[k] = `true`
					} else {
						d.Tags[k] = `false`
					}

				case string:
					if v_ != "name" && v_ != "status" {
						d.Tags[k] = v_
					} else {
						l.Debugf("ignore tag %s:%s from region info", k, v_)
					}
				default:
					l.Warnf("ignore key `%s' of type %s", k, reflect.TypeOf(v).String())
				}
			}

		default:
			l.Debugf("pass %s", k)
		}
	}

	for k, x := range resp.Content {
		l.Debugf(`class: %s`, k)

		if k == RegionInfo {
			continue
		}

		arr, ok := x.([]interface{})

		if !ok {
			l.Warnf("invalid resp.Content, expect []interface{}, got %s", reflect.TypeOf(x).String())
			continue
		}

		if k == dt.ClassHeadless {
			l.Debugf("ignore %d headless tasks", len(arr))
			continue
		}

		for _, data := range arr {
			var t dt.Task

			switch k {
			case dt.ClassHTTP:
				t = &dt.HTTPTask{Option: map[string]string{"userAgent": fmt.Sprintf("DataKit/%s dialtesting", datakit.Version)}}
			case dt.ClassDNS:
				// TODO
				l.Warnf("DNS task deprecated, ignored")
				continue
			case dt.ClassTCP:
				t = &dt.TcpTask{}
			case dt.ClassWebsocket:
				t = &dt.WebsocketTask{}
			case dt.ClassICMP:
				t = &dt.IcmpTask{}
			case dt.ClassOther:
				// TODO
				l.Warnf("OTHER task deprecated, ignored")
				continue
			default:
				l.Errorf("unknown task type: %s", k)
			}

			if t == nil {
				l.Warn("empty task, ignored")
				continue
			}

			j, ok := data.(string)
			if !ok {
				l.Warnf("invalid task data, expect string, got %s", reflect.TypeOf(data).String())
				continue
			}

			if err := json.Unmarshal([]byte(j), &t); err != nil {
				l.Errorf(`json.Unmarshal: %s`, err.Error())
				return err
			}

			l.Debugf("unmarshal task: %+#v", t)

			// update dialer pos
			ts := t.UpdateTimeUs()
			if d.pos < ts {
				d.pos = ts
				l.Debugf("update position to %d", d.pos)
			}

			l.Debugf(`%+#v id: %s`, d.curTasks[t.ID()], t.ID())

			if dialer, ok := d.curTasks[t.ID()]; ok { // update task
				if dialer.failCnt >= MaxFails {
					l.Warnf(`failed %d times,ignore`, dialer.failCnt)
					delete(d.curTasks, t.ID())
					continue
				}

				if err := dialer.updateTask(t); err != nil {
					l.Warnf(`%s,ignore`, err.Error())
				}

				if strings.ToLower(t.Status()) == dt.StatusStop {
					delete(d.curTasks, t.ID())
				}
			} else { // create new task
				if strings.ToLower(t.Status()) == dt.StatusStop {
					l.Warnf(`%s status is stop, exit ignore`, t.ID())
					continue
				}

				l.Debugf(`create new task %+#v`, t)
				dialer, err := d.newTaskRun(t)
				if err != nil {
					l.Errorf(`%s, ignore`, err.Error())
				} else {
					d.curTasks[t.ID()] = dialer
				}
			}
		}
	}

	l.Debugf("current tasks: %+#v", d.curTasks)
	return nil
}

func (d *Input) getLocalJSONTasks(data []byte) ([]byte, error) {
	// 转化结构，json结构转成与kodo服务一样的格式
	var resp map[string][]interface{}
	if err := json.Unmarshal(data, &resp); err != nil {
		l.Error(err)
		return nil, err
	}

	content := map[string]interface{}{}

	for k, v := range resp {
		vs := []string{}
		for _, v1 := range v {
			dt, err := json.Marshal(v1)
			if err != nil {
				l.Error(err)
				return nil, err
			}

			vs = append(vs, string(dt))
		}

		content[k] = vs
	}

	tasks := taskPullResp{
		Content: content,
	}
	rs, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		l.Error(err)
		return nil, err
	}

	return rs, nil
}

func (d *Input) pullTask() ([]byte, error) {
	reqURL, err := url.Parse(d.Server)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return nil, err
	}

	var res []byte
	for i := 0; i <= 3; i++ {
		var statusCode int
		res, statusCode, err = d.pullHTTPTask(reqURL, d.pos)
		if statusCode/100 != 5 { // 500 err 重试
			break
		}
	}

	l.Debugf("task body: %s", string(res))

	return res, err
}

func signReq(req *http.Request, ak, sk string) {
	so := &uhttp.SignOption{
		AuthorizationType: AuthorizationType,
		SignHeaders:       SignHeaders,
		SK:                sk,
	}

	reqSign, err := so.SignReq(req)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("DIAL_TESTING %s:%s", ak, reqSign))
}

func (d *Input) pullHTTPTask(reqURL *url.URL, sinceUs int64) ([]byte, int, error) {
	reqURL.Path = "/v1/task/pull"
	reqURL.RawQuery = fmt.Sprintf("region_id=%s&since=%d", d.RegionID, sinceUs)

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return nil, 5, err
	}

	bodymd5 := fmt.Sprintf("%x", md5.Sum([]byte(""))) //nolint:gosec
	req.Header.Set("Date", time.Now().Format(http.TimeFormat))
	req.Header.Set("Content-MD5", bodymd5)
	req.Header.Set("Connection", "close")
	signReq(req, d.AK, d.SK)

	resp, err := d.cli.Do(req)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return nil, 5, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return nil, 0, err
	}

	defer resp.Body.Close() //nolint:errcheck
	switch resp.StatusCode / 100 {
	case 2: // ok
		return body, resp.StatusCode / 100, nil
	default:
		l.Warn("request %s failed(%s): %s", d.Server, resp.Status, string(body))
		if strings.Contains(string(body), `kodo.RegionNotFoundOrDisabled`) {
			// stop all
			d.stopAlltask()
		}
		return nil, resp.StatusCode / 100, fmt.Errorf("pull task failed")
	}
}

// ReadEnv support envs:
// ENV_INPUT_DIALTESTING_AK: string
// ENV_INPUT_DIALTESTING_SK: string
// ENV_INPUT_DIALTESTING_REGION_ID: string
// ENV_INPUT_DIALTESTING_SERVER: string.
func (d *Input) ReadEnv(envs map[string]string) {
	if ak, ok := envs["ENV_INPUT_DIALTESTING_AK"]; ok {
		d.AK = ak
	}

	if sk, ok := envs["ENV_INPUT_DIALTESTING_SK"]; ok {
		d.SK = sk
	}

	if regionID, ok := envs["ENV_INPUT_DIALTESTING_REGION_ID"]; ok {
		d.RegionID = regionID
	}

	if server, ok := envs["ENV_INPUT_DIALTESTING_SERVER"]; ok {
		d.Server = server
	}
}

func (d *Input) stopAlltask() {
	for tid, dialer := range d.curTasks {
		dialer.stop()
		delete(d.curTasks, tid)
	}
}

func newDefaultInput() *Input {
	return &Input{
		Tags:     map[string]string{},
		curTasks: map[string]*dialer{},
		semStop:  cliutils.NewSem(),
		cli: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig:     &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
				TLSHandshakeTimeout: 30 * time.Second,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
			},
		},
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return newDefaultInput()
	})
}
