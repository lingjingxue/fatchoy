// Copyright © 2020 simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package collections

import (
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

const (
	TickDuration = 100 * time.Millisecond
	WheelSize    = 512
	WheelMask    = WheelSize - 1
)

// A timer optimized for approximated I/O timeout scheduling.
//
// Implementation Details
//   [Netty HashedWheelTimer](https://github.com/netty/netty/blob/4.1/common/src/main/java/io/netty/util/HashedWheelTimer.java)
//   [Hashed and Hierarchical Timing Wheels](http://www.cs.columbia.edu/~nahum/w6998/papers/sosp87-timing-wheels.pdf)
type HashedWheelTimer struct {
	done               chan struct{}
	wg                 sync.WaitGroup
	state              int32
	wheel              []HashedWheelBucket      // wheel buckets
	C                  <-chan TimerTask         // 到期的定时器
	timeouts           UnboundedConcurrentQueue // all timeout
	cancelledTimeouts  UnboundedConcurrentQueue //
	pendingTimeouts    int32                    // maximum number of pending timeouts
	maxPendingTimeouts int32                    //
	lastId             int64                    //
	startedAt          int64                    //
}

func NewHashedWheelTimer() *HashedWheelTimer {
	return &HashedWheelTimer{
		done:  make(chan struct{}),
		wheel: make([]HashedWheelBucket, WheelSize),
	}
}

func (t *HashedWheelTimer) Shutdown() {
	close(t.done)
	t.wg.Wait()
	t.wheel = nil
}

func (t *HashedWheelTimer) CreateTimeout(delay int64, task TimerTask) *HashedWheelTimeout {
	// Starts the background thread explicitly.
	t.start()

	if delay < 0 {
		delay = 0
	}
	var deadline = currentMs() + delay
	if delay > 0 && deadline < 0 {
		deadline = math.MaxInt64
	}

	var id = atomic.AddInt64(&t.lastId, 1)
	var timeout = NewHashedWheelTimeout(t, id, deadline, task)
	t.timeouts.Enqueue(timeout)
	return timeout
}

func (t *HashedWheelTimer) decrementPending() int32 {
	return atomic.AddInt32(&t.pendingTimeouts, -1)
}

func (t *HashedWheelTimer) start() {
	var state = atomic.LoadInt32(&t.state)
	switch state {
	case WorkerInit:
		if atomic.CompareAndSwapInt32(&t.state, WorkerInit, WorkerStarted) {
			var ready = make(chan struct{}, 1)
			t.wg.Add(1)
			go t.worker(ready)
			<-ready
		}
	case WorkerStarted:
		return

	default:
		log.Panicf("invalid worker state %v", state)
	}
}

func (t *HashedWheelTimer) worker(ready chan struct{}) {
	defer func() {
		atomic.StoreInt32(&t.state, WorkerShutdown)
		t.wg.Done()
	}()

	var ticker = time.NewTicker(TickDuration)
	defer ticker.Stop()

	var bus = make(chan TimerTask, 1000)
	t.startedAt = currentMs()
	t.C = bus
	ready <- struct{}{}

	var ticks = 0
	for {
		select {
		case now := <-ticker.C:
			t.tick(bus, timeMs(now), ticks)

		case <-t.done:
			t.finalize()
			return
		}
	}
}

// Fill the unprocessedTimeouts so we can return them from stop() method
func (t *HashedWheelTimer) finalize() {
	var unprocessedTimeouts = make(map[int64]*HashedWheelTimeout)
	for _, bucket := range t.wheel {
		bucket.ClearTimeouts(unprocessedTimeouts)
	}
}

func (t *HashedWheelTimer) processCancelledTasks() {
	for {
		v, ok := t.cancelledTimeouts.Dequeue()
		if !ok || v == nil {
			break
		}
		var timeout = v.(*HashedWheelTimeout)
		timeout.remove()
	}
}

func (t *HashedWheelTimer) transferTimeoutToBuckets(nowTick int) {
	// transfer only max. 100000 timeouts per tick to prevent a thread
	// to stale when it just  adds new timeouts in a loop.
	for i := 0; i <= 1e5; i++ {
		v, ok := t.timeouts.Dequeue()
		if !ok || v == nil {
			break
		}
		var timeout = v.(*HashedWheelTimeout)
		if timeout.IsCanceled() {
			continue
		}
		var calculated = int(timeout.Deadline / int64(TickDuration))
		timeout.remainingRounds = int32((calculated - nowTick) / len(t.wheel))
		// Ensure we don't schedule for past.
		var ticks = calculated
		if ticks < nowTick {
			ticks = nowTick
		}
		var stopIndex = ticks & WheelMask
		t.wheel[stopIndex].AddTimeout(timeout)
	}
}

func (t *HashedWheelTimer) tick(bus chan<- TimerTask, deadline int64, ticks int) {
	var idx = ticks & WheelMask
	t.processCancelledTasks()
	var bucket = t.wheel[idx]
	t.transferTimeoutToBuckets(ticks)
	bucket.ExpireTimeouts(bus, deadline)
}
