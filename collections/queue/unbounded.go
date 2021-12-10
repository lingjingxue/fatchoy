// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package queue

// Keeping below as var so it is possible to run the slice size bench tests with no coding changes.
var (
	// firstSliceSize holds the size of the first slice.
	firstSliceSize = 1

	// maxFirstSliceSize holds the maximum size of the first slice.
	maxFirstSliceSize = 16

	// maxInternalSliceSize holds the maximum size of each internal slice.
	maxInternalSliceSize = 128
)

// queueNode represents a queue node.
// Each node holds a slice of user managed values.
type queueArrayNode struct {
	val  []interface{}   // v holds the list of user added values in this node.
	next *queueArrayNode // n points to the next node in the linked list.
}

// newQueueNode returns an initialized node.
func newQueueArrayNode(capacity int) *queueArrayNode {
	return &queueArrayNode{
		val: make([]interface{}, 0, capacity),
	}
}

// UnboundedQueue represents an unbounded, dynamically growing FIFO queue.
// The zero value for queue is an empty queue ready to use.
// see https://github.com/golang/go/issues/27935
type UnboundedQueue struct {
	// In an empty queue, head and tail points to the same node.
	head *queueArrayNode
	tail *queueArrayNode

	hp  int // index pointing to the current first element in the queue
	len int // Len holds the current queue values length.

	lastSliceSize int // lastSliceSize holds the size of the last created internal slice.
}

func NewUnbounded() *UnboundedQueue {
	return new(UnboundedQueue).Init()
}

// Init initializes or clears queue q.
func (q *UnboundedQueue) Init() *UnboundedQueue {
	q.head = nil
	q.tail = nil
	q.hp = 0
	q.len = 0
	return q
}

func (q *UnboundedQueue) Len() int {
	return q.len
}

// Front returns the first element of queue q or nil if the queue is empty.
// The second, bool result indicates whether a valid value was returned;
//   if the queue is empty, false will be returned.
// The complexity is O(1).
func (q *UnboundedQueue) Front() (interface{}, bool) {
	if q.head == nil {
		return nil, false
	}
	return q.head.val[q.hp], true
}

// Push adds a value to the queue.
// The complexity is O(1).
func (q *UnboundedQueue) Push(v interface{}) {
	if q.head == nil {
		h := newQueueArrayNode(firstSliceSize)
		q.head = h
		q.tail = h
		q.lastSliceSize = maxFirstSliceSize
	} else if len(q.tail.val) >= q.lastSliceSize {
		n := newQueueArrayNode(maxInternalSliceSize)
		q.tail.next = n
		q.tail = n
		q.lastSliceSize = maxInternalSliceSize
	}

	q.tail.val = append(q.tail.val, v)
	q.len++
}

// Pop retrieves and removes the current element from the queue.
// The second, bool result indicates whether a valid value was returned;
// 	if the queue is empty, false will be returned.
// The complexity is O(1).
func (q *UnboundedQueue) Pop() (interface{}, bool) {
	if q.head == nil {
		return nil, false
	}

	v := q.head.val[q.hp]
	q.head.val[q.hp] = nil // Avoid memory leaks
	q.len--
	q.hp++
	if q.hp >= len(q.head.val) {
		n := q.head.next
		q.head.next = nil // Avoid memory leaks
		q.head = n
		q.hp = 0
	}
	return v, true
}
