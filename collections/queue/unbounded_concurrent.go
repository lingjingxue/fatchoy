// Copyright © 2020 simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package queue

import (
	"sync"
)

type UnboundedConcurrentQueue struct {
	queue UnboundedQueue
	guard sync.RWMutex
}

func NewUnboundedConcurrentQueue() *UnboundedConcurrentQueue {
	return &UnboundedConcurrentQueue{}
}

func (q *UnboundedConcurrentQueue) Len() int {
	q.guard.RLock()
	var n = q.queue.Len()
	q.guard.RUnlock()
	return n
}

func (q *UnboundedConcurrentQueue) Enqueue(item interface{}) {
	q.guard.Lock()
	q.queue.Push(item)
	q.guard.Unlock()
}

func (q *UnboundedConcurrentQueue) Dequeue() (v interface{}, ok bool) {
	q.guard.Lock()
	v, ok = q.queue.Pop()
	q.guard.Unlock()
	return
}
