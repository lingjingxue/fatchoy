// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

import (
	"errors"
	"sync"

	"qchen.fun/fatchoy"
	"qchen.fun/fatchoy/debug"
	"qchen.fun/fatchoy/log"
)

var ErrExecutorNotRunning = errors.New("executor not running")

type ThreadPoolExecutor struct {
	done    chan struct{}
	wg      sync.WaitGroup
	state   fatchoy.State //
	queue   chan Runnable // work queue
	nworker int           //
}

func NewThreadPoolExecutor(nworker, capacity int) Executor {
	if nworker <= 0 {
		nworker = 1
	}
	return &ThreadPoolExecutor{
		nworker: nworker,
		done:    make(chan struct{}),
		queue:   make(chan Runnable, capacity),
	}
}

func NewAsyncExecutor(capacity int) Executor {
	return NewThreadPoolExecutor(1, capacity)
}

func (e *ThreadPoolExecutor) Execute(r Runnable) error {
	e.start()
	if e.state.Get() != fatchoy.StateRunning {
		return ErrExecutorNotRunning
	}
	e.queue <- r // may block
	return nil
}

func (e *ThreadPoolExecutor) Shutdown() {
	if !e.state.CAS(fatchoy.StateRunning, fatchoy.StateShutdown) {
		return
	}
	close(e.done)
	e.wg.Wait()
	close(e.queue)
	e.state.Set(fatchoy.StateTerminated)
}

func (e *ThreadPoolExecutor) start() {
	var state = e.state.Get()
	switch state {
	case fatchoy.StateInit:
		if e.state.CAS(fatchoy.StateInit, fatchoy.StateStarted) {
			var ready = make(chan struct{}, e.nworker)
			for i := 0; i < e.nworker; i++ {
				e.wg.Add(1)
				go e.worker(i + 1)
			}
			for i := 0; i < e.nworker; i++ {
				<-ready
			}
			e.state.Set(fatchoy.StateRunning)
		}

	case fatchoy.StateRunning:
		return

	default:
		log.Panicf("invalid executor state %v", state)
	}
}

func (e *ThreadPoolExecutor) run(r Runnable) {
	defer debug.CatchPanic()
	if err := r.Run(); err != nil {
		log.Errorf("executor run: %v", err)
	}
}

func (e *ThreadPoolExecutor) worker(i int) {
	defer e.wg.Done()
	for {
		select {
		case r := <-e.queue:
			e.run(r)

		case <-e.done:
			return
		}
	}
}
