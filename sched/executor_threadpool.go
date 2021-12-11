// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

import (
	"sync"
)

const (
	StateRunning    = 1
	StateShutdown   = 2
	StateStop       = 3
	StateTidying    = 4
	StateTerminated = 5
)

type ThreadPoolExecutor struct {
	done  chan struct{}
	wg    sync.WaitGroup
	queue chan Runnable // work queue
}

func NewThreadPoolExecutor(capacity int) Executor {
	return &ThreadPoolExecutor{
		done: make(chan struct{}),
		queue: make(chan Runnable, capacity),
	}
}

func (e *ThreadPoolExecutor) Execute(r Runnable) error {
	return nil
}
