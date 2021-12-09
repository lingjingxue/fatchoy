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
)

type Executor interface {
	Execute(r Runnable) error
}
