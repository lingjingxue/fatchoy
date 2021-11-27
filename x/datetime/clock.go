// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package datetime

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	DefaultTickInterval = 250 * time.Millisecond // 1秒tick4次
)

// Clock提供一些对壁钟时间的操作，不适用于高精度的计时场景
// 设计初衷是为精度至少为秒的上层业务服务, 支持时钟的往前/后调拨
type Clock struct {
	done     chan struct{}
	wg       sync.WaitGroup
	traveled time.Duration // 旅行时间，提供对时钟的往前/后拨动
	nanosec  int64         // 当前tick的时间戳(in nanoseconds, up to 2262)
	ticks    int64         // 当前tick
	ticker   *time.Ticker  //
}

func NewClock(interval time.Duration) *Clock {
	if interval <= 0 {
		interval = DefaultTickInterval
	}
	c := &Clock{
		done:    make(chan struct{}),
		nanosec: time.Now().UnixNano(),
		ticker:  time.NewTicker(interval),
	}
	return c
}

func (c *Clock) Go() {
	c.wg.Add(1)
	go c.serve()
}

func (c *Clock) serve() {
	defer c.wg.Done()
	for {
		select {
		case t, ok := <-c.ticker.C:
			if !ok {
				return
			}
			atomic.StoreInt64(&c.nanosec, t.UnixNano())
			atomic.AddInt64(&c.ticks, 1)

		case <-c.done:
			return
		}
	}
}

func (c *Clock) Stop() {
	close(c.done)
	c.wg.Wait()
	c.ticker.Stop()
	c.ticker = nil
}

func (c *Clock) TickCount() int64 {
	return atomic.LoadInt64(&c.ticks)
}

func (c *Clock) Now() time.Time {
	ts := atomic.LoadInt64(&c.nanosec)
	now := time.Unix(ts/1e9, ts%1e9)
	if c.traveled != 0 {
		return now.Add(c.traveled)
	}
	return now
}

func (c *Clock) DateTime() string {
	now := c.Now()
	return now.Format(DateFormat)
}

// 恢复时钟
func (c *Clock) Reset() {
	c.traveled = 0
}

// 时间旅行(拨动时钟)
func (c *Clock) Travel(d time.Duration) {
	c.traveled += d
}
