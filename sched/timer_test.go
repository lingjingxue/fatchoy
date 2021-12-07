// Copyright © 2020 qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

import (
	"context"
	"math"
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type testRunner struct {
	interval     int
	fireCount    int
	startTime    time.Time
	lastFireTime time.Time
}

func newTestRunner(interval int) *testRunner {
	return &testRunner{
		interval:  interval,
		startTime: time.Now(),
	}
}

func (r *testRunner) Run(ctx context.Context) error {
	r.lastFireTime = time.Now()
	r.fireCount++
	return nil
}

func TestScheduler_RunAfter(t *testing.T) {
	var sched = NewHeapTimer()
	sched.Start()
	defer sched.Shutdown()

	var ctx = context.Background()
	var interval = 1200 // 1.2s
	var runner = newTestRunner(interval)

	sched.RunAfter(interval, runner)

	for runner.fireCount == 0 {
		select {
		case r := <-sched.C:
			r.R.Run(ctx)
			t.Logf("timer fired at %v", runner.lastFireTime)
			duration := runner.lastFireTime.Sub(runner.startTime)
			if duration < time.Duration(interval)*time.Millisecond {
				t.Fatalf("invalid fire duration: %v != %v", duration, interval)
			}
			runner.startTime = runner.lastFireTime
		}
	}
}

func TestScheduler_RunEvery(t *testing.T) {
	var sched = NewHeapTimer()
	sched.Start()
	defer sched.Shutdown()

	var interval = 700 // 0.7s
	var runner = newTestRunner(interval)
	sched.RunEvery(interval, runner)

	var ctx = context.Background()

	for runner.fireCount < 5 {
		select {
		case r := <-sched.C:
			r.R.Run(ctx)
			t.Logf("timer fired at %v", runner.lastFireTime)
			duration := runner.lastFireTime.Sub(runner.startTime).Milliseconds()
			deviation := duration - int64(interval)
			if math.Abs(float64(deviation)) > TimerPrecision {
				t.Fatalf("invalid fire duration: %v != %v", duration, interval)
			}
			runner.startTime = runner.lastFireTime
		}
	}
}
