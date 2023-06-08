// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

package winevent

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"strings"
	"syscall"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"golang.org/x/sys/windows"
)

var statusList = []string{"info", "critical", "error", "warning", "info"}

func (*Input) SampleConfig() string {
	return sample
}

func (*Input) Catalog() string {
	return "windows"
}

// RunPipeline TODO.
func (*Input) RunPipeline() {
}

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLabelWindows}
}

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&Measurement{},
	}
}

func (ipt *Input) Run() {
	l = logger.SLogger("win event log")
	var err error

	ipt.subscription, err = ipt.evtSubscribe("", ipt.Query)
	if err != nil {
		io.FeedLastError(inputName, err.Error())
		return
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("win event exit")
			return

		case <-ipt.semStop.Wait():
			l.Info("win event return")
			return

		default:
			time.Sleep(time.Millisecond * 1)
			start := time.Now()
			events, err := ipt.fetchEvents(ipt.subscription)
			if err != nil {
				if errors.Is(err, ErrorNoMoreItems) {
					continue
				}
				l.Error(err.Error())
				io.FeedLastError(inputName, err.Error())
				return
			}
			for _, event := range events {
				ipt.handleEvent(event)
			}
			if len(ipt.collectCache) > 0 {
				err := inputs.FeedMeasurement(inputName, datakit.Logging,
					ipt.collectCache, &io.Option{CollectCost: time.Since(start)})
				if err != nil {
					l.Error(err.Error())
					io.FeedLastError(inputName, err.Error())
				}
				ipt.collectCache = ipt.collectCache[:0]
			}
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) handleEvent(event Event) {
	ts, err := time.Parse("2006-01-02T15:04:05.000000000Z", event.TimeCreated.SystemTime)
	if err != nil {
		l.Error(err.Error())
		ts = time.Now()
	}
	tags := map[string]string{}
	for k, v := range ipt.Tags {
		tags[k] = v
	}

	msg, err := json.Marshal(event)
	if err != nil {
		l.Error(err.Error())
		return
	}
	fields := map[string]interface{}{
		"event_source":    event.Source.Name,
		"event_id":        event.EventID,
		"version":         event.Version,
		"task":            event.TaskText,
		"keyword":         event.Keywords,
		"event_record_id": event.EventRecordID,
		"process_id":      int(event.Execution.ProcessID),
		"channel":         event.Channel,
		"computer":        event.Computer,
		"message":         event.Message,
		"level":           event.LevelText,
		"total_message":   string(msg),
		"status":          ipt.getEventStatus(event.Level),
	}
	metric := &Measurement{
		tags:   tags,
		fields: fields,
		ts:     ts,
		name:   "windows_event",
	}
	ipt.collectCache = append(ipt.collectCache, metric)
}

func (ipt *Input) getEventStatus(level int) string {
	if level >= 0 && level < len(statusList) {
		return statusList[level]
	}

	return "info"
}

func (ipt *Input) evtSubscribe(logName, xquery string) (EvtHandle, error) {
	var logNamePtr, xqueryPtr *uint16

	sigEvent, err := windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(sigEvent) // nolint:errcheck

	logNamePtr, err = syscall.UTF16PtrFromString(logName)
	if err != nil {
		return 0, err
	}

	xqueryPtr, err = syscall.UTF16PtrFromString(xquery)
	if err != nil {
		return 0, err
	}

	subsHandle, err := _EvtSubscribe(0, uintptr(sigEvent), logNamePtr, xqueryPtr,
		0, 0, 0, EvtSubscribeToFutureEvents)
	if err != nil {
		return 0, err
	}

	return subsHandle, nil
}

func (ipt *Input) fetchEventHandles(subsHandle EvtHandle) ([]EvtHandle, error) {
	var eventsNumber uint32
	var evtReturned uint32

	eventsNumber = 5

	eventHandles := make([]EvtHandle, eventsNumber)

	err := _EvtNext(subsHandle, eventsNumber, &eventHandles[0], 0, 0, &evtReturned)
	if err != nil {
		if errors.Is(err, ErrorInvalidOperation) && evtReturned == 0 {
			return nil, ErrorNoMoreItems
		}
		return nil, err
	}

	return eventHandles[:evtReturned], nil
}

func (ipt *Input) fetchEvents(subsHandle EvtHandle) ([]Event, error) {
	var events []Event

	eventHandles, err := ipt.fetchEventHandles(subsHandle)
	if err != nil {
		return nil, err
	}

	for _, eventHandle := range eventHandles {
		if eventHandle != 0 {
			event, err := ipt.renderEvent(eventHandle)
			if err == nil {
				// w.Log.Debugf("Got event: %v", event)
				events = append(events, event)
			}
		}
	}

	for i := 0; i < len(eventHandles); i++ {
		err := _EvtClose(eventHandles[i])
		if err != nil {
			return events, err
		}
	}
	return events, nil
}

func (ipt *Input) renderEvent(eventHandle EvtHandle) (Event, error) {
	var bufferUsed, propertyCount uint32

	event := Event{}
	err := _EvtRender(0, eventHandle, EvtRenderEventXML, uint32(len(ipt.buf)), &ipt.buf[0], &bufferUsed, &propertyCount)
	if err != nil {
		return event, err
	}

	eventXML, err := DecodeUTF16(ipt.buf[:bufferUsed])
	if err != nil {
		return event, err
	}
	err = xml.Unmarshal(eventXML, &event)
	if err != nil {
		// We can return event without most text values,
		// that way we will not loose information
		// This can happen when processing Forwarded Events
		return event, nil //nolint:nilerr
	}

	publisherHandle, err := openPublisherMetadata(0, event.Source.Name, 0)
	if err != nil {
		return event, nil //nolint:nilerr
	}
	defer _EvtClose(publisherHandle) // nolint:errcheck

	// Populating text values
	keywords, err := formatEventString(EvtFormatMessageKeyword, eventHandle, publisherHandle)
	if err == nil {
		event.Keywords = keywords
	} else {
		l.Warn(err)
	}
	message, err := formatEventString(EvtFormatMessageEvent, eventHandle, publisherHandle)
	if err == nil {
		scanner := bufio.NewScanner(strings.NewReader(message))
		scanner.Scan()
		message = scanner.Text()
		event.Message = message
	} else {
		l.Warn(err)
	}
	level, err := formatEventString(EvtFormatMessageLevel, eventHandle, publisherHandle)
	if err == nil {
		event.LevelText = level
	} else {
		l.Warn(err)
	}
	task, err := formatEventString(EvtFormatMessageTask, eventHandle, publisherHandle)
	if err == nil {
		event.TaskText = task
	} else {
		l.Warn(err)
	}
	opcode, err := formatEventString(EvtFormatMessageOpcode, eventHandle, publisherHandle)
	if err == nil {
		event.OpcodeText = opcode
	} else {
		l.Warn(err)
	}
	return event, nil
}

func formatEventString(
	messageFlag EvtFormatMessageFlag,
	eventHandle EvtHandle,
	publisherHandle EvtHandle,
) (string, error) {
	var bufferUsed uint32
	err := _EvtFormatMessage(publisherHandle, eventHandle, 0, 0, 0, messageFlag,
		0, nil, &bufferUsed)
	if err != nil && !errors.Is(err, ErrorInsufficientBuffer) {
		return "", err
	}

	bufferUsed *= 2
	buffer := make([]byte, bufferUsed)
	bufferUsed = 0

	err = _EvtFormatMessage(publisherHandle, eventHandle, 0, 0, 0, messageFlag,
		uint32(len(buffer)/2), &buffer[0], &bufferUsed)
	bufferUsed *= 2
	if err != nil {
		return "", err
	}

	result, err := DecodeUTF16(buffer[:bufferUsed])
	if err != nil {
		return "", err
	}

	var out string
	if messageFlag == EvtFormatMessageKeyword {
		// Keywords are returned as array of a zero-terminated strings
		splitZero := func(c rune) bool { return c == '\x00' }
		eventKeywords := strings.FieldsFunc(string(result), splitZero)
		// So convert them to comma-separated string
		out = strings.Join(eventKeywords, ",")
	} else {
		result := bytes.Trim(result, "\x00")
		out = string(result)
	}
	return out, nil
}

// openPublisherMetadata opens a handle to the publisher's metadata. Close must
// be called on returned EvtHandle when finished with the handle.
func openPublisherMetadata(
	session EvtHandle,
	publisherName string,
	lang uint32,
) (EvtHandle, error) {
	p, err := syscall.UTF16PtrFromString(publisherName)
	if err != nil {
		return 0, err
	}

	h, err := _EvtOpenPublisherMetadata(session, p, nil, lang, 0)
	if err != nil {
		return 0, err
	}

	return h, nil
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			buf:   make([]byte, 1<<14),
			Query: query,

			semStop: cliutils.NewSem(),
		}
	})
}
