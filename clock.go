// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"time"

	"qchen.fun/fatchoy/x/datetime"
)

var (
	ClockEpoch int64           = 1577836800 // 2020-01-01 00:00:00 UTC
	gClock     *datetime.Clock              // default clock
)

// 开启时钟
func StartClock() {
	gClock = datetime.NewClock(0)
	gClock.Go()
}

// 关闭时钟
func StopClock() {
	if gClock != nil {
		gClock.Stop()
		gClock = nil
	}
}

func WallClock() *datetime.Clock {
	return gClock
}

func Now() time.Time {
	return gClock.Now()
}

func NowString() string {
	return gClock.Now().Format(datetime.TimestampFormat)
}

func DateTime() string {
	return gClock.DateTime()
}
