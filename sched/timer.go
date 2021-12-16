// Copyright © 2020 simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

import (
	"time"
)

const (
	TimeUnit             = 10                              // centi-seconds (10 ms)
	PendingQueueCapacity = 1000                            //
	CustomEpoch          = int64(1577836800 * time.Second) // 起始纪元 2020-01-01 00:00:00 UTC
)

type Timer interface {
	RunAfter(durationMs int, task Runnable) int
	RunEvery(intervalMs int, task Runnable) int
	Cancel(id int) bool
	IsScheduled(id int) bool
	Chan() <-chan Runnable
	Size() int
	Shutdown()
}
