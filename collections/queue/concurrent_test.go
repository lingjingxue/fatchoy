// Copyright © 2020 simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package queue

import (
	"context"
	"testing"
	"time"
)

func TestUnboundedConcurrentQueue_Enqueue(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var que = NewConcurrentUnboundedQueue()
	var done = make(chan struct{}, 1)
	var ready = make(chan struct{}, 1)

	const maxN = 10

	go func() {
		var count = maxN
		for count > 0 {
			ready <- struct{}{}
			select {
			case <-que.Signal():
				for !que.IsEmpty() {
					if v, ok := que.Dequeue(); ok {
						t.Logf("dequeue item %v", v)
					}
					count--
				}
			}
		}
		done <- struct{}{}
	}()

	<-ready
	for i := 10; i < 10+maxN; i++ {
		que.Enqueue(i)
	}

	select {
	case <-done:
		break
	case err := <-ctx.Done():
		t.Fatalf("%v", err)
		break
	}
}
