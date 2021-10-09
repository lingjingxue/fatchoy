// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"gopkg.in/qchencc/fatchoy/debug"
	"gopkg.in/qchencc/fatchoy/log"
	"gopkg.in/qchencc/fatchoy/x/stats"
)

var (
	ErrExecutorClosed = errors.New("executor is closed already")
	ErrExecutorBusy   = errors.New("executor queue is full")
)

// 执行器
type Executor struct {
	closing   int32              //
	ctx       context.Context    //
	cancel    context.CancelFunc //
	workerCnt int                // 并发数量
	bus       chan Runner        // 待执行runner队列
	stats     *stats.Stats       // 执行统计
}

func NewExecutor(parentCtx context.Context, concurrency, queueSize int) *Executor {
	e := &Executor{}
	e.Init(parentCtx, concurrency, queueSize)
	return e
}

func (e *Executor) Init(parentCtx context.Context, concurrency, queueSize int) {
	if queueSize <= 0 {
		queueSize = 1
	}
	if concurrency == 0 {
		queueSize = 0
	}
	e.workerCnt = concurrency
	e.bus = make(chan Runner, queueSize)
	ctx, cancel := context.WithCancel(parentCtx)
	e.ctx = ctx
	e.cancel = cancel
}

func (e *Executor) Start() {
	if e.workerCnt > 0 {
		//e.wg.Add(e.concurrency)
		for i := 1; i <= e.workerCnt; i++ {
			go e.serve(i)
		}
	}
}

func (e *Executor) Stats() *stats.Stats {
	return e.stats.Clone()
}

func (e *Executor) Stop() {
	e.cancel()
}

func (e *Executor) StopAndWait() {
	e.cancel()
	timer := time.NewTimer(time.Second * 5)
	select {
	case <-timer.C:
		return
	case <-e.ctx.Done():
		return
	}
}

//
func (e *Executor) Shutdown() {
	log.Debugf("start shutting down executor")
	if !atomic.CompareAndSwapInt32(&e.closing, 0, 1) {
		return
	}
	e.StopAndWait()
	close(e.bus)
	e.bus = nil
	e.stats = nil
}

func (e *Executor) Execute(r Runner) error {
	if atomic.LoadInt32(&e.closing) == 1 {
		return ErrExecutorClosed
	}
	// 同步执行
	if e.workerCnt == 0 {
		return e.run(r)
	}
	// 异步并发执行
	select {
	case e.bus <- r:
	default:
		return ErrExecutorBusy
	}
	return nil
}

func (e *Executor) run(r Runner) (err error) {
	defer func() {
		if v := recover(); v != nil {
			debug.Backtrace(v, os.Stderr)
			err = fmt.Errorf("%v", v)
		}
	}()

	err = r.Run()
	return
}

func (e *Executor) serve(idx int) {
	log.Debugf("executor worker #%d start serving", idx)
	for {
		select {
		case r, ok := <-e.bus:
			if !ok {
				return
			}
			if err := e.run(r); err != nil {
				log.Warnf("execute %T: %v", r, err)
			}

		case <-e.ctx.Done():
			return
		}
	}
}
