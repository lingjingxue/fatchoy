// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

//go:build !ignore

package sched

import (
	"context"
	"sync"
	"testing"
)

type testRunner2 struct {
	guard  sync.Mutex
	expect int32
	count  int32
	done   chan struct{}
}

func newTestRunner2(n int32) *testRunner2 {
	return &testRunner2{
		expect: n,
		done:   make(chan struct{}),
	}
}

func (r *testRunner2) Run() error {
	r.guard.Lock()
	r.count++
	if r.count >= r.expect {
		select {
		case r.done <- struct{}{}:
		default:
		}
	}
	r.guard.Unlock()
	return nil
}

func TestExecutorSingle(t *testing.T) {
	var exe = NewExecutor(context.Background(), 0, 1000)
	exe.Start()
	defer exe.Shutdown()

	var r = newTestRunner2(10)
	for i := 0; i < 10; i++ {
		exe.Execute(r)
	}
	t.Logf("runner execute count %d", r.count)
}

func TestExecutorConcurrent(t *testing.T) {
	var exe = NewExecutor(context.Background(), 12, 1000)
	exe.Start()
	defer exe.Shutdown()

	var r = newTestRunner2(100)
	for i := 0; i < 100; i++ {
		exe.Execute(r)
	}
	<-r.done
	t.Logf("runner execute count %d", r.count)
}
