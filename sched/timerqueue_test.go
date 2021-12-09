// Copyright © 2020 simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

import (
	"math"
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type testTimerContext struct {
	interval     int
	fireCount    int
	startTime    time.Time
	lastFireTime time.Time
}

func newTestTimerContext(interval int) *testTimerContext {
	return &testTimerContext{
		interval:  interval,
		startTime: time.Now(),
	}
}

func (r *testTimerContext) Run() {
	r.lastFireTime = time.Now()
	r.fireCount++
}

func TestScheduler_RunAfter(t *testing.T) {
	var sched = NewTimerQueue()
	defer sched.Shutdown()

	var interval = 1200 // 1.2s
	var ctx = newTestTimerContext(interval)

	sched.RunAfter(interval, ctx)

	for ctx.fireCount == 0 {
		select {
		case task := <-sched.C:
			task.Run()
			duration := ctx.lastFireTime.Sub(ctx.startTime)
			t.Logf("timer fired after %v at %s", duration, ctx.lastFireTime.Format(time.RFC3339))
			if duration < time.Duration(interval)*time.Millisecond {
				t.Fatalf("invalid fire duration: %v != %v", duration, interval)
			}
			ctx.startTime = ctx.lastFireTime
		}
	}
}

func TestScheduler_RunEvery(t *testing.T) {
	var sched = NewTimerQueue()
	defer sched.Shutdown()

	var interval = 700 // 0.7s
	var ctx = newTestTimerContext(interval)
	sched.RunEvery(interval, ctx)

	for ctx.fireCount < 5 {
		select {
		case task := <-sched.C:
			task.Run()
			duration := ctx.lastFireTime.Sub(ctx.startTime)
			t.Logf("timer fired after %v at %s", duration, ctx.lastFireTime.Format(time.RFC3339))
			deviation := duration.Milliseconds() - int64(interval)
			if math.Abs(float64(deviation)) > float64(TimeUnit) {
				t.Fatalf("invalid fire duration: %v != %v", duration, interval)
			}
			ctx.startTime = ctx.lastFireTime
		}
	}
}
