// Copyright © 2020 simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

import (
	"context"
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type timerContext struct {
	interval     int
	fireCount    int
	startTime    time.Time
	lastFireTime time.Time
}

func newTimerContext(interval int) *timerContext {
	return &timerContext{
		interval:  interval,
		startTime: time.Now(),
	}
}

func (r *timerContext) Run() error {
	r.lastFireTime = time.Now()
	r.fireCount++
	return nil
}

func testTimerCancel(t *testing.T, sched Timer) {
	const interval = 1000 // 1s
	var timerCtx = newTimerContext(interval)

	var timerId = sched.RunAfter(interval, timerCtx)
	time.Sleep(time.Millisecond) // wait for worker
	if n := sched.Size(); n != 1 {
		t.Fatalf("timer size unexpected %d", n)
	}
	sched.Cancel(timerId)
	if n := sched.Size(); n != 0 {
		t.Fatalf("timer size unexpected %d", n)
	}
	if timerCtx.fireCount > 0 {
		t.Fatalf("timeout %d unexpectly triggered", timerId)
	}
}

func testTimerRunAfter(t *testing.T, sched Timer, interval int) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	var timerCtx = newTimerContext(interval)

	sched.RunAfter(interval, timerCtx)

	for timerCtx.fireCount == 0 {
		select {
		case task := <-sched.Chan():
			task.Run()
			duration := timerCtx.lastFireTime.Sub(timerCtx.startTime)
			t.Logf("timer fired after %v at %s", duration, timerCtx.lastFireTime.Format(time.RFC3339))
			if duration < time.Duration(interval)*time.Millisecond {
				t.Fatalf("fired too early %v != %v", duration, interval)
			}
			timerCtx.startTime = time.Now()

		case <-ctx.Done():
			t.Fatalf("test deadline exceeded")
			return
		}
	}
}

func testTimerRunEvery(t *testing.T, sched Timer, interval int) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	var timerCtx = newTimerContext(interval)
	sched.RunEvery(interval, timerCtx)

	t.Logf("timer fired start at %s", timerCtx.startTime.Format(time.RFC3339))

	for timerCtx.fireCount < 20 {
		select {
		case task := <-sched.Chan():
			task.Run()
			var duration = timerCtx.lastFireTime.Sub(timerCtx.startTime)
			t.Logf("timer fired after %v at %s", duration, timerCtx.lastFireTime.Format(time.RFC3339))
			if d := duration.Milliseconds(); d < int64(interval) {
				t.Errorf("timeout too early %d < %d", d, interval)
			}
			timerCtx.startTime = timerCtx.lastFireTime

		case <-ctx.Done():
			t.Fatalf("test deadline exceeded")
			return
		}
	}
}

func TestTimerQueue_RunAfter(t *testing.T) {
	var timer = NewDefaultTimerQueue()
	timer.Start()
	defer timer.Shutdown()

	testTimerCancel(t, timer)
	for i := 100; i <= 1000; i+= 100 {
		testTimerRunAfter(t, timer, i)
	}
}

func TestTimerQueue_RunEvery(t *testing.T) {
	var timer = NewDefaultTimerQueue()
	timer.Start()
	defer timer.Shutdown()

	testTimerRunEvery(t, timer, 300)
}

func TestHHWheel_RunAfter(t *testing.T) {
	var timer = NewDefaultHHWheelTimer()
	timer.Start()
	defer timer.Shutdown()

	testTimerCancel(t, timer)
	for i := 100; i <= 1000; i+= 100 {
		testTimerRunAfter(t, timer, i)
	}
}

func TestHHWheel_RunEvery(t *testing.T) {
	var timer = NewDefaultHHWheelTimer()
	timer.Start()
	defer timer.Shutdown()

	testTimerRunEvery(t, timer, 300)
}
