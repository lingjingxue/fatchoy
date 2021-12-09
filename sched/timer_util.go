// Copyright © 2020 simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

import (
	"time"
)

const (
	TimeUnit    = 10 * time.Millisecond           //
	CustomEpoch = int64(1577836800 * time.Second) // 起始纪元 2020-01-01 00:00:00 UTC
)

const (
	WorkerInit     = 0
	WorkerStarted  = 1
	WorkerShutdown = 2
)

// 当前毫秒
func currentMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// 转为当前毫秒
func timeMs(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}
