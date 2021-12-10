// Copyright © 2020 simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package queue

import (
	"sync"
)

// Unbounded FIFO concurrent queue
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

// Enqueue enqueues an element
func (q *UnboundedConcurrentQueue) Enqueue(item interface{}) {
	q.guard.Lock()
	q.queue.Push(item)
	q.guard.Unlock()
}

// Dequeue dequeues an element. Returns false if queue is locked or empty.
func (q *UnboundedConcurrentQueue) Dequeue() (v interface{}, ok bool) {
	q.guard.Lock()
	v, ok = q.queue.Pop()
	q.guard.Unlock()
	return
}

// Peek returns front element's value and keeps the element at the queue
func (q *UnboundedConcurrentQueue) Peek() (v interface{}, ok bool) {
	q.guard.Lock()
	v, ok = q.queue.Front()
	q.guard.Unlock()
	return
}
