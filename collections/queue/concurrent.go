// Copyright © 2020 simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package queue

import (
	"sync"
)

// Unbounded FIFO concurrent queue
type ConcurrentUnboundedQueue struct {
	queue UnboundedQueue
	guard sync.RWMutex
	wait  chan struct{}
}

func NewConcurrentUnboundedQueue() *ConcurrentUnboundedQueue {
	return &ConcurrentUnboundedQueue{
		wait: make(chan struct{}, 1),
	}
}

func (q *ConcurrentUnboundedQueue) Signal() <-chan struct{} {
	return q.wait
}

func (q *ConcurrentUnboundedQueue) notify() {
	select {
	case q.wait <- struct{}{}:
	default:
		return
	}
}

func (q *ConcurrentUnboundedQueue) IsEmpty() bool {
	return q.Len() == 0
}

func (q *ConcurrentUnboundedQueue) Len() int {
	q.guard.RLock()
	var n = q.queue.Len()
	q.guard.RUnlock()
	return n
}

// Enqueue enqueues an element
func (q *ConcurrentUnboundedQueue) Enqueue(item interface{}) {
	q.guard.Lock()
	q.queue.Push(item)
	q.guard.Unlock()
	q.notify()
}

// Dequeue dequeues an element. Returns false if queue is locked or empty.
func (q *ConcurrentUnboundedQueue) Dequeue() (v interface{}, ok bool) {
	q.guard.Lock()
	v, ok = q.queue.Pop()
	q.guard.Unlock()
	return
}

// Peek returns front element's value and keeps the element at the queue
func (q *ConcurrentUnboundedQueue) Peek() (v interface{}, ok bool) {
	q.guard.Lock()
	v, ok = q.queue.Front()
	q.guard.Unlock()
	return
}
