// Copyright © 2020 simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

import (
	"log"
)

// Bucket that stores HashedWheelTimeouts.
// These are stored in a linked-list like data structure to allow easy removal of HashedWheelTimeouts in the middle.
// Also, the HashedWheelTimeout act as nodes them-self and so no extra object creation is needed.
type HashedWheelBucket struct {
	head, tail *HashedWheelTimeout // linked list
}

// add `timeout` to this bucket
func (b *HashedWheelBucket) AddTimeout(timeout *HashedWheelTimeout) {
	if timeout.bucket != nil {
		panic("unexpected timeout linked to another bucket")
	}
	timeout.bucket = b
	if b.head == nil {
		b.head = timeout
		b.tail = timeout
	} else {
		b.tail.next = timeout
		timeout.prev = b.tail
		b.tail = timeout
	}
}

// Expire all HashedWheelTimeouts for the given deadline.
func (b *HashedWheelBucket) ExpireTimeouts(bus chan<- Runnable, deadline int64) {
	var timeout = b.head
	for timeout != nil {
		var next = timeout.next
		if timeout.remainingRounds <= 0 {
			next = b.Remove(timeout)
			if timeout.Deadline <= deadline {
				timeout.Expire(bus)
			} else {
				// The timeout was placed into a wrong slot. This should never happen.
				log.Panicf("timeout.deadline (%d) > deadline (%d)", timeout.Deadline, deadline)
			}
		} else if timeout.IsCanceled() {
			next = b.Remove(timeout)
		} else {
			timeout.remainingRounds -= 1
		}
		timeout = next
	}
}

// remove `timeout` from linked list and return next linked one
func (b *HashedWheelBucket) Remove(timeout *HashedWheelTimeout) *HashedWheelTimeout {
	var next = timeout.next
	// remove timeout that was either processed or cancelled by updating the linked-list
	if timeout.prev != nil {
		timeout.prev.next = next
	}
	if timeout.next != nil {
		timeout.next.prev = timeout.prev
	}
	if timeout == b.head {
		// if timeout is also the tail we need to adjust the entry too
		if timeout == b.tail {
			b.head = nil
			b.tail = nil
		} else {
			b.head = next
		}
	} else if timeout == b.tail {
		// if the timeout is the tail modify the tail to be the prev node.
		b.tail = timeout.prev
	}
	// null out prev, next and bucket to allow for GC.
	timeout.prev = nil
	timeout.next = nil
	timeout.bucket = nil
	timeout.Timer.decrementPending()
	return next
}

// Clear this bucket and return all not expired / cancelled Timeouts.
func (b *HashedWheelBucket) ClearTimeouts(set map[int64]*HashedWheelTimeout) {
	for {
		var timeout = b.pollTimeout()
		if timeout == nil {
			return
		}
		if timeout.IsExpired() || timeout.IsCanceled() {
			continue
		}
		set[timeout.Id] = timeout
	}
}

// poll first timeout
func (b *HashedWheelBucket) pollTimeout() *HashedWheelTimeout {
	var head = b.head
	if head == nil {
		return nil
	}
	var next = head.next
	if next == nil {
		b.tail = nil
		b.head = nil
	} else {
		b.head = next
		next.prev = nil
	}
	// null out prev and next to allow for GC.
	head.next = nil
	head.prev = nil
	head.bucket = nil
	return head
}
