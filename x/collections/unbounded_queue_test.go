// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package collections

import (
	"strconv"
	"testing"
)

func TestNewQueueShouldReturnInitiazedInstanceOfQueue(t *testing.T) {
	q := NewUnboundedQueue()

	if q == nil {
		t.Error("Expected: new instance of queue; Got: nil")
	}
}

func TestQueueWithZeroValueShouldReturnReadyToUseQueue(t *testing.T) {
	var q UnboundedQueue
	q.Push(1)
	q.Push(2)

	v, ok := q.Pop()
	if !ok || v.(int) != 1 {
		t.Errorf("Expected: 1; Got: %d", v)
	}
	v, ok = q.Pop()
	if !ok || v.(int) != 2 {
		t.Errorf("Expected: 2; Got: %d", v)
	}
	_, ok = q.Pop()
	if ok {
		t.Error("Expected: empty slice (ok=false); Got: ok=true")
	}
}

func TestQueueWithZeroValueAndEmptyShouldReturnAsEmpty(t *testing.T) {
	var q UnboundedQueue
	if _, ok := q.Front(); ok {
		t.Error("Expected: false as the queue is empty; Got: true")
	}
	if _, ok := q.Pop(); ok {
		t.Error("Expected: false as the queue is empty; Got: true")
	}
	if l := q.Len(); l != 0 {
		t.Errorf("Expected: 0 as the queue is empty; Got: %d", l)
	}
}

func TestQueueInitShouldReturnAEmptyQueue(t *testing.T) {
	var q UnboundedQueue
	q.Push(1)

	q.Init()

	if _, ok := q.Front(); ok {
		t.Error("Expected: false as the queue is empty; Got: true")
	}
	if _, ok := q.Pop(); ok {
		t.Error("Expected: false as the queue is empty; Got: true")
	}
	if l := q.Len(); l != 0 {
		t.Errorf("Expected: 0 as the queue is empty; Got: %d", l)
	}
}

func TestQueueWithNilValuesShouldReturnAllValuesInOrder(t *testing.T) {
	q := NewUnboundedQueue()
	q.Push(1)
	q.Push(nil)
	q.Push(2)
	q.Push(nil)

	v, ok := q.Pop()
	if !ok || v.(int) != 1 {
		t.Errorf("Expected: 1; Got: %d", v)
	}
	v, ok = q.Pop()
	if !ok || v != nil {
		t.Errorf("Expected: 1; Got: %d", v)
	}
	v, ok = q.Pop()
	if !ok || v.(int) != 2 {
		t.Errorf("Expected: 1; Got: %d", v)
	}
	v, ok = q.Pop()
	if !ok || v != nil {
		t.Errorf("Expected: 1; Got: %d", v)
	}
	_, ok = q.Pop()
	if ok {
		t.Error("Expected: empty slice (ok=false); Got: ok=true")
	}
}

func TestQueuePushPopFrontShouldRetrieveAllElementsInOrder(t *testing.T) {
	tests := map[string]struct {
		putCount       []int
		getCount       []int
		remainingCount int
	}{
		"Test 1 item": {
			putCount:       []int{1},
			getCount:       []int{1},
			remainingCount: 0,
		},
		"Test 100 items": {
			putCount:       []int{100},
			getCount:       []int{100},
			remainingCount: 0,
		},
		"Test 1000 items": {
			putCount:       []int{1000},
			getCount:       []int{1000},
			remainingCount: 0,
		},
		"Test sequence 1": {
			putCount:       []int{1, 2, 100, 101},
			getCount:       []int{1, 2, 100, 101},
			remainingCount: 0,
		},
		"Test sequence 2": {
			putCount:       []int{10, 1},
			getCount:       []int{1, 10},
			remainingCount: 0,
		},
		"Test sequence 3": {
			putCount:       []int{101, 101},
			getCount:       []int{100, 101},
			remainingCount: 1,
		},
		"Test sequence 4": {
			putCount:       []int{1000, 1000, 1001},
			getCount:       []int{10, 10, 1},
			remainingCount: 2980,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			q := NewUnboundedQueue()
			lastPut := 0
			lastGet := 0
			var ok bool
			var v interface{}
			for count := 0; count < len(test.getCount); count++ {
				for i := 1; i <= test.putCount[count]; i++ {
					lastPut++
					q.Push(lastPut)
					if v, ok = q.Front(); !ok || v != lastGet+1 {
						t.Errorf("Expected: %d; Got: %d", lastGet, v)
					}
				}

				for i := 1; i <= test.getCount[count]; i++ {
					lastGet++
					v, ok = q.Front()
					if !ok || v.(int) != lastGet {
						t.Errorf("Expected: %d; Got: %d", lastGet, v)
					}
					v, ok = q.Pop()
					if !ok || v.(int) != lastGet {
						t.Errorf("Expected: %d; Got: %d", lastGet, v)
					}
				}
			}

			if q.Len() != test.remainingCount {
				t.Errorf("Expected: %d; Got: %d", test.remainingCount, q.Len())
			}

			if test.remainingCount > 0 {
				if v, ok = q.Front(); !ok || v == nil {
					t.Error("Expected: non-empty queue; Got: empty")
				}
			} else {
				if v, ok = q.Front(); ok || v != nil {
					t.Error("Expected: empty queue; Got: non-empty")
				}
			}

			for i := 1; i <= test.remainingCount; i++ {
				lastGet++

				if v, ok = q.Front(); !ok || v.(int) != lastGet {
					t.Errorf("Expected: %d; Got: %d", lastGet, v)
				}
				v, ok = q.Pop()
				if !ok || v.(int) != lastGet {
					t.Errorf("Expected: %d; Got: %d", lastGet, v)
				}
			}
			if v, ok = q.Front(); ok || v != nil {
				t.Errorf("Expected: nil as the queue should be empty; Got: %d", v)
			}
			if v, ok = q.Pop(); ok || v != nil {
				t.Errorf("Expected: nil as the queue should be empty; Got: %d", v)
			}
			if v, ok = q.Front(); ok || v != nil {
				t.Error("Expected: empty queue; Got: non-empty")
			}
			if q.Len() != 0 {
				t.Errorf("Expected: %d; Got: %d", 0, q.Len())
			}
		})
	}
}

const (
	// count holds the number of items to add to the queue.
	count = 1000
)

var (
	// Used to store temp values, avoiding any compiler optimizations.
	tmp  interface{}
	tmp2 bool
)

func BenchmarkQueueMaxFirstSliceSize(b *testing.B) {
	tests := []struct {
		name string
		size int
	}{
		{size: 1},
		{size: 2},
		{size: 4},
		{size: 8},
		{size: 16},
		{size: 32},
		{size: 64},
		{size: 128},
	}

	for _, test := range tests {
		b.Run(strconv.Itoa(test.size), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				maxFirstSliceSize = test.size
				q := NewUnboundedQueue()

				for i := 0; i < count; i++ {
					q.Push(n)
				}
				for tmp, tmp2 = q.Pop(); tmp2; tmp, tmp2 = q.Pop() {
				}
			}
		})
	}
}

func BenchmarkQueueMaxSubsequentSliceSize(b *testing.B) {
	tests := []struct {
		name string
		size int
	}{
		{size: 1},
		{size: 10},
		{size: 20},
		{size: 30},
		{size: 40},
		{size: 50},
		{size: 60},
		{size: 70},
		{size: 80},
		{size: 90},
		{size: 100},
		{size: 200},
		{size: 2},
		{size: 4},
		{size: 8},
		{size: 16},
		{size: 32},
		{size: 64},
		{size: 128},
		{size: 256},
		{size: 512},
		{size: 1024},
	}

	for _, test := range tests {
		b.Run(strconv.Itoa(test.size), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				maxInternalSliceSize = test.size
				q := NewUnboundedQueue()

				for i := 0; i < count; i++ {
					q.Push(n)
				}
				for tmp, tmp2 = q.Pop(); tmp2; tmp, tmp2 = q.Pop() {
				}
			}
		})
	}
}
