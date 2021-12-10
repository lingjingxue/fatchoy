// Copyright © 2020 simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package queue

import (
	"testing"
)

func TestUnboundedConcurrentQueue_Enqueue(t *testing.T) {
	var q = NewUnboundedConcurrentQueue()
	for i := 0; i < 10; i++ {
		q.Enqueue(i)
	}
}
