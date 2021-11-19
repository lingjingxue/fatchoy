// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"testing"
	"time"
)

func TestStartClock(t *testing.T) {
	StartClock()
	defer StopClock()

	t.Logf("now: %s", NowString())
	t.Logf("date: %s", DateTime())

	var clock = WallClock()
	var t1 = Now()
	clock.Travel(-time.Hour)
	var t2 = Now()
	t.Logf("elapsed %v", t1.Sub(t2))
}
