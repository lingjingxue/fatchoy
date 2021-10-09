// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

type RunFunc func() error

// Runner是一个可执行对象
type Runner interface {
	Run() error
}

type Runnable struct {
	F RunFunc
}

func (r *Runnable) Run() error {
	return r.F()
}

func NewRunner(f RunFunc) Runner {
	return &Runnable{
		F: f,
	}
}
