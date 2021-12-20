// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

import (
	"errors"
)

var (
	ErrExecutorClosed = errors.New("executor is closed already")
	ErrExecutorBusy   = errors.New("executor queue is full")
 ErrExecutorNotRunning = errors.New("executor not running")
)

type Executor interface {
	Start()
	Execute(r Runnable) error
	Shutdown()
}
