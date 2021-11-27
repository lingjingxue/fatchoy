// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"time"

	"qchen.fun/fatchoy/x/datetime"
)

// epoch of clock
const ClockEpoch int64 = 1577836800 // 2020-01-01 00:00:00 UTC

// default clock
var gClock *datetime.Clock

// 开启时钟
func StartClock() {
	gClock = datetime.NewClock(time.Millisecond * 250)
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
