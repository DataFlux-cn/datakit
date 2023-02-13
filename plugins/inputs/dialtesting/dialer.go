// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	_ "github.com/go-ping/ping"
	dt "github.com/GuanceCloud/cliutils/dialtesting"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
)

type dialer struct {
	task dt.Task

	ticker *time.Ticker

	initTime time.Time
	testCnt  int64
	class    string

	tags     map[string]string
	updateCh chan dt.Task

	category string

	failCnt int
	done    <-chan interface{}
}

func (d *dialer) updateTask(t dt.Task) error {
	select {
	case <-d.updateCh: // if closed?
		l.Warnf("task %s closed", d.task.ID())
		return fmt.Errorf("task exited")
	default:
		d.updateCh <- t
		return nil
	}
}

func (d *dialer) stop() {
	if err := d.task.Stop(); err != nil {
		l.Warnf("stop task %s failed: %s", d.task.ID(), err.Error())
	}
}

func newDialer(t dt.Task, ts map[string]string) *dialer {
	return &dialer{
		task:     t,
		updateCh: make(chan dt.Task),
		initTime: time.Now(),
		tags:     ts,
		class:    t.Class(),
	}
}

func (d *dialer) getSendFailCount() int32 {
	if d.category != "" {
		return dataway.GetSendStat(d.category)
	}
	return 0
}

func (d *dialer) run() error {
	d.ticker = d.task.Ticker()

	l.Debugf("dialer: %+#v", d)

	defer d.ticker.Stop()
	defer close(d.updateCh)

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Infof("dial testing %s exit", d.task.ID())
			return nil

		case <-d.done:
			l.Infof("dial testing %s exit", d.task.ID())
			return nil

		case <-d.ticker.C:
			failCount := d.getSendFailCount()
			if failCount > MaxSendFailCount {
				l.Warnf("dial testing %s send data failed %d times", d.task.ID(), failCount)
				return nil
			}

			l.Debugf(`dialer run %+#v, fail count: %d`, d, failCount)
			d.testCnt++

			switch d.task.Class() {
			case dt.ClassHeadless:
				return fmt.Errorf("headless task deprecated")
			default:
				_ = d.task.Run() //nolint:errcheck
			}

			// dialtesting start
			// 无论成功或失败，都要记录测试结果
			err := d.feedIO()
			if err != nil {
				l.Warnf("io feed failed, %s", err.Error())
			}

		case t := <-d.updateCh:
			d.doUpdateTask(t)

			if strings.ToLower(d.task.Status()) == dt.StatusStop {
				d.stop()
				l.Info("task %s stopped", d.task.ID())
				return nil
			}
		}
	}
}

func (d *dialer) feedIO() error {
	// 考虑到推送至不同的dataway地址
	u, err := url.Parse(d.task.PostURLStr())
	if err != nil {
		l.Warn("get invalid url, ignored")
		return err
	}

	u.Path += datakit.Logging // `/v1/write/logging`

	urlStr := u.String()
	switch d.task.Class() {
	case dt.ClassHTTP, dt.ClassTCP, dt.ClassICMP, dt.ClassWebsocket:
		d.category = urlStr
		return d.pointsFeed(urlStr)
	case dt.ClassHeadless:
		return fmt.Errorf("headless task deprecated")
	default:
		// TODO other class
	}

	return nil
}

func (d *dialer) doUpdateTask(t dt.Task) {
	if err := t.Init(); err != nil {
		l.Warn(err)
		return
	}

	if d.task.GetFrequency() != t.GetFrequency() {
		d.ticker = t.Ticker() // update ticker
	}

	d.task = t
}
