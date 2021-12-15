// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"sync/atomic"
)

const (
	StateInit       = 0
	StateStarted    = 1
	StateRunning    = 2
	StateShutdown   = 3
	StateTerminated = 4
)

// service state
type State int32

func (s *State) Get() int32 {
	return atomic.LoadInt32((*int32)(s))
}

func (s *State) Set(n int32) {
	atomic.StoreInt32((*int32)(s), n)
}

func (s *State) CAS(old, new int32) bool {
	return atomic.CompareAndSwapInt32((*int32)(s), old, new)
}

func (s State) IsRunning() bool {
	return s.Get() == StateRunning
}

func (s State) IsShuttingDown() bool {
	return s.Get() == StateShutdown
}

func (s State) IsTerminated() bool {
	return s.Get() == StateTerminated
}
