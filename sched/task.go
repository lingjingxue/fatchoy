// Copyright © 2020 qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

import (
	"sync/atomic"
)

const (
	StateInit      = 0
	StateScheduled = 1 // task is scheduled for execution
	StateExecuted  = 2 // a non-repeating task has already executed (or is currently executing) and has not been cancelled.
	StateCancelled = 3 // task has been cancelled (with a call to TimerTask.Cancel).
)

// Runnable代表一个可执行对象
type Runnable interface {
	Run() error
}

// 待执行的任务
type Task struct {
	state  int32
	action func() error
}

func NewTask(action func() error) *Task {
	return &Task{
		action: action,
	}
}

func (r *Task) State() int32 {
	return atomic.LoadInt32(&r.state)
}

func (r *Task) SetState(state int32) {
	atomic.StoreInt32(&r.state, state)
}

func (r *Task) Cancel() bool {
	return atomic.CompareAndSwapInt32(&r.state, StateScheduled, StateCancelled)
}

func (r *Task) Run() error {
	if r.action != nil {
		return r.action()
	}
	return nil
}
