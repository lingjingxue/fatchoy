// Copyright © 2020 qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

// Runner是一个可执行对象
type Runnable interface {
	Run() error
}

type Task struct {
	F func() error
}

func (r *Task) Run() error {
	return r.F()
}

func NewTask(f func() error) *Task {
	return &Task{
		F: f,
	}
}
