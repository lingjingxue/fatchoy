// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

import (
	"qchen.fun/fatchoy/debug"
)

type ImmediateExecutor struct {
}

var _ie = NewImmediateExecutor()

func NewImmediateExecutor() Executor {
	return &ImmediateExecutor{}
}

func (e *ImmediateExecutor) Start() {
	// do nothing here
}

func (e *ImmediateExecutor) Shutdown() {
	// do nothing here
}

func (e *ImmediateExecutor) Execute(r Runnable) error {
	defer debug.CatchPanic()
	return r.Run()
}

func (e *ImmediateExecutor) Instance() Executor {
	return _ie
}
