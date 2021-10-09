// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package datetime

import (
	"testing"
	"time"
)

func TestClockExample(t *testing.T) {
	clock := NewClock(DefaultTickInterval)
	clock.Go()
	defer clock.Stop()
	time.Sleep(10 * time.Millisecond)

	now := clock.DateTime()
	t.Logf("now: %v", now)

	clock.Travel(time.Hour * 2) // 往前拨2小时
	t.Logf("t1: %v", clock.DateTime())

	clock.Travel(time.Hour * -3) // 往后拨3小时
	t.Logf("t2: %v", clock.DateTime())
}
