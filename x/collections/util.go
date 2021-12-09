// Copyright © 2020 simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package collections

import (
	"time"
)

const (
	TimeUnit    = 10 * time.Millisecond           //
	CustomEpoch = int64(1577836800 * time.Second) // 起始纪元 2020-01-01 00:00:00 UTC
)

// 定时器回调函数
type TimerTask func()

// 当前毫秒
func currentMs() int64 {
	return (time.Now().UTC().UnixNano() - CustomEpoch) / int64(time.Millisecond)
}

// 转为当前毫秒
func nowMs(now time.Time) int64 {
	return (now.UTC().UnixNano() - CustomEpoch) / int64(time.Millisecond)
}
