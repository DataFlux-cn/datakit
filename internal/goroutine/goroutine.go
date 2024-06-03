// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package goroutine implement grouptine pool
package goroutine

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
)

// stat cache the group statistic info.
var (
	stat = make(map[string]*StatInfo)
	mu   sync.Mutex

	log = logger.DefaultSLogger("goroutine")
)

// A Group is a collection of goroutines working on subtasks that are part of
// the same overall task.
type Group struct {
	chs []func(ctx context.Context) error

	name string
	err  error
	ctx  context.Context

	panicCb  func([]byte) bool // callback when panic
	beforeCb func()            // job before callback

	panicTimeout time.Duration // time duration between panicCb call
	ch           chan func(ctx context.Context) error
	cancel       func()
	wg           sync.WaitGroup

	errOnce    sync.Once
	workerOnce sync.Once
	panicTimes int8 // max panic times
}

// Option provides the setup of a group.
type Option struct {
	Name         string
	PanicCb      func([]byte) bool
	PanicTimes   int8
	PanicTimeout time.Duration
}

// NewGroup create a custom group.
func NewGroup(option Option) *Group {
	log = logger.SLogger("goroutine")

	name := "default"
	if len(option.Name) > 0 {
		name = option.Name
	}
	g := &Group{
		name:         name,
		panicCb:      option.PanicCb,
		panicTimes:   option.PanicTimes,
		panicTimeout: option.PanicTimeout,
	}

	if g.panicCb == nil {
		g.panicCb = func(crashStack []byte) bool {
			log.Errorf("recover panic: %s", string(crashStack))
			goroutineCrashedVec.WithLabelValues(name).Inc()
			return true
		}
	}

	goroutineGroups.Inc()

	return g
}

// WithContext create a Group.
// given function from Go will receive this context,.
func WithContext(ctx context.Context) *Group {
	return &Group{ctx: ctx}
}

// WithCancel create a new Group and an associated Context derived from ctx.
//
// given function from Go will receive context derived from this ctx,
// The derived Context is canceled the first time a function passed to Go
// returns a non-nil error or the first time Wait returns, whichever occurs
// first.
func WithCancel(ctx context.Context) *Group {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{ctx: ctx, cancel: cancel}
}

func (g *Group) do(f func(ctx context.Context) error) {
	if g.beforeCb != nil {
		g.beforeCb()
	}

	ctx := g.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	panicTimes := g.panicTimes - 1

	var (
		err   error
		run   func()
		start = time.Now()
	)

	run = func() {
		defer func() {
			if r := recover(); r != nil {
				goroutineCrashedVec.WithLabelValues(g.name).Inc()

				isPanicRetry := true
				buf := make([]byte, 4096) //nolint:gomnd
				buf = buf[:runtime.Stack(buf, false)]

				if e, ok := r.(error); ok {
					buf = append([]byte(fmt.Sprintf("%s\n", e.Error())), buf...)
				}

				if g.panicCb != nil {
					isPanicRetry = g.panicCb(buf)
				}

				if isPanicRetry && panicTimes > 0 {
					panicTimes--

					if g.panicTimeout > 0 {
						time.Sleep(g.panicTimeout)
					}

					goroutineRecoverVec.WithLabelValues(g.name).Inc()

					run()

					return
				} else {
					goroutineCounterVec.WithLabelValues(g.name).Dec()
					goroutineCostVec.WithLabelValues(g.name).Observe(float64(time.Since(start)) / float64(time.Second))
					goroutineStoppedVec.WithLabelValues(g.name).Inc()
				}

				err = fmt.Errorf("goroutine: panic recovered: %s", r)
			} else {
				goroutineCounterVec.WithLabelValues(g.name).Dec()
				goroutineCostVec.WithLabelValues(g.name).Observe(float64(time.Since(start)) / float64(time.Second))
				goroutineStoppedVec.WithLabelValues(g.name).Inc()
			}

			if err != nil {
				g.errOnce.Do(func() {
					g.err = err

					if g.cancel != nil {
						g.cancel()
					}
				})
			}

			g.wg.Done()
		}()

		err = f(ctx)
	}

	run()
}

// GOMAXPROCS set max goroutine to work.
func (g *Group) GOMAXPROCS(n int) {
	if n <= 0 {
		panic("goroutine: GOMAXPROCS must great than 0")
	}

	g.workerOnce.Do(func() {
		g.ch = make(chan func(context.Context) error, n)
		for i := 0; i < n; i++ {
			go func() {
				for f := range g.ch {
					g.do(f)
				}
			}()
		}
	})
}

// Go calls the given function in a new goroutine.
//
// The first call to return a non-nil error cancels the group; its error will be
// returned by Wait.
func (g *Group) Go(f func(ctx context.Context) error) {
	g.wg.Add(1)

	goroutineCounterVec.WithLabelValues(g.name).Inc()

	if g.ch != nil {
		select {
		case g.ch <- f:
		default:
			g.chs = append(g.chs, f)
		}

		return
	}

	go g.do(f)
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
func (g *Group) Wait() error {
	if g.ch != nil {
		for _, f := range g.chs {
			g.ch <- f
		}
	}

	g.wg.Wait()
	if g.ch != nil {
		close(g.ch) // let all receiver exit
	}

	if g.cancel != nil {
		g.cancel()
	}

	return g.err
}

func (g *Group) Name() string {
	return g.name
}
