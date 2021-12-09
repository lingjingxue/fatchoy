// Copyright © 2020 simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

import (
	"fmt"
	"strings"
	"sync/atomic"
)

const (
	TimeoutStateInit      = 0
	TimeoutStateCancelled = 1
	TimeoutStateExpired   = 2
)

// 定时器节点
type HashedWheelTimeout struct {
	// used to chain timeouts in HashedWheelTimerBucket via a double-linked-list.
	next, prev *HashedWheelTimeout

	// The bucket to which the timeout was added
	bucket *HashedWheelBucket
	Timer  *HashedWheelTimer //

	state int32 // cancelled or expired

	// remainingRounds will be calculated and set before added to the correct HashedWheelBucket
	remainingRounds int32

	Task     Runnable // 到期触发任务
	Id       int64     // 唯一ID
	Deadline int64     // 到期时间
}

func NewHashedWheelTimeout(timer *HashedWheelTimer, id, deadline int64, task Runnable) *HashedWheelTimeout {
	return &HashedWheelTimeout{
		state:    TimeoutStateInit,
		Id:       id,
		Timer:    timer,
		Deadline: deadline,
		Task:     task,
	}
}

func (m *HashedWheelTimeout) State() int32 {
	return atomic.LoadInt32(&m.state)
}

func (m *HashedWheelTimeout) IsExpired() bool {
	return m.State() == TimeoutStateExpired
}

func (m *HashedWheelTimeout) IsCanceled() bool {
	return m.State() == TimeoutStateCancelled
}

func (m *HashedWheelTimeout) Cancel() bool {
	if !atomic.CompareAndSwapInt32(&m.state, TimeoutStateInit, TimeoutStateCancelled) {
		return false
	}
	// If a task should be canceled we put this to another queue which will be processed on each tick.
	// So this means that we will have a GC latency of max. 1 tick duration which is good enough.
	m.Timer.cancelledTimeouts.Enqueue(m)
	return true
}

func (m *HashedWheelTimeout) Expire(bus chan<- Runnable) {
	if !atomic.CompareAndSwapInt32(&m.state, TimeoutStateInit, TimeoutStateExpired) {
		return
	}
	if m.Task != nil {
		bus <- m.Task
	}
}

func (m *HashedWheelTimeout) remove() {
	if m.bucket != nil {
		m.bucket.Remove(m)
	} else {
		m.Timer.decrementPending()
	}
}

func (m *HashedWheelTimeout) String() string {
	var remaining = m.Deadline - currentMs()
	var buf strings.Builder
	buf.WriteString("HashedWheelTimeout(deadline: ")
	if remaining > 0 {
		fmt.Fprintf(&buf, "%d ms later", remaining)
	} else if remaining < 0 {
		fmt.Fprintf(&buf, "%d ms ago", remaining)
	} else {
		buf.WriteString("now")
	}
	if m.IsCanceled() {
		buf.WriteString(", cancelled")
	}
	buf.WriteString(")")
	return buf.String()
}
