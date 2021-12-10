// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

import (
	"fmt"
	"testing"
	"time"
)

func TestNewHashedWheelTimer(t *testing.T) {
	var timer = NewHashedWheelTimer()
	defer timer.Shutdown()

	var task = NewTask(func() error {
		var now = time.Now()
		fmt.Printf("timeout at %v", now.Format(time.RFC3339))
		return nil
	})

	var timeout = timer.CreateTimeout(1200, task)

	var C1 = time.After(time.Second * 3)
	var C2 = time.After(time.Second * 2)
	for {
		select {
		case tsk := <-timer.C:
			tsk.Run()

		case <-C2:
			timeout.Cancel()

		case <-C1:
			return
		}
	}
}
