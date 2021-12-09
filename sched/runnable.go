// Copyright © 2020 qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

// Runner是一个可执行对象
type Runnable interface {
	Run()
}

type Task struct {
	F func()
}

func (r *Task) Run() {
	r.F()
}

func NewTask(f func()) *Task {
	return &Task{
		F: f,
	}
}
