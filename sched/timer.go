// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

import (
	"container/heap"
	"context"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"qchen.fun/fatchoy/qlog"
)

const (
	TimerPrecision    = 50   // 精度为50ms
	TimerChanCapacity = 1000 //
	TimerCapacity     = 64
)

// 最小堆定时器
type TimeoutScheduler struct {
	ctx     context.Context    //
	cancel  context.CancelFunc //
	ticker  *time.Ticker       // 精度ticker
	ticks   int64              //
	guard   sync.Mutex         // 多线程guard
	timers  TimerHeap          // timer heap
	nextId  int                // time id生成
	ref     map[int]*timerNode // O(1)查找
	C       chan *timerNode    // 到期的定时器
	expires []*timerNode
}

func NewTimeoutScheduler(parentCtx context.Context) *TimeoutScheduler {
	t := &TimeoutScheduler{}
	t.Init(parentCtx)
	return t
}

func currentMilliSec() int64 {
	return time.Now().UnixNano() / 1000_000 // to millisecond
}

func (s *TimeoutScheduler) Init(parentCtx context.Context) {
	s.nextId = 1
	ctx, cancel := context.WithCancel(parentCtx)
	s.ctx = ctx
	s.cancel = cancel
	s.ticker = time.NewTicker(TimerPrecision * time.Millisecond)
	s.timers = make(TimerHeap, 0, TimerCapacity)
	s.ref = make(map[int]*timerNode, TimerCapacity)
	s.C = make(chan *timerNode, TimerChanCapacity)
	s.expires = make([]*timerNode, 0, 4)
}

func (s *TimeoutScheduler) Start() {
	go s.serve()
}

func (s *TimeoutScheduler) StopAndWait() {
	s.ticker.Stop()
	s.cancel()
	timer := time.NewTimer(time.Second * 5)
	select {
	case <-timer.C:
		return
	case <-s.ctx.Done():
		return
	}
}

func (s *TimeoutScheduler) Shutdown() {
	s.StopAndWait()
	close(s.C)
	s.C = nil
	s.ticker = nil
	s.ref = nil
	s.timers = nil
}

func (s *TimeoutScheduler) serve() {
	qlog.Debugf("scheduler start serving")
	for {
		select {
		case now := <-s.ticker.C:
			s.update(now)

		case <-s.ctx.Done():
			return
		}
	}
}

func (s *TimeoutScheduler) update(now time.Time) {
	atomic.AddInt64(&s.ticks, 1)
	s.guard.Lock()
	s.trigger(now)
	s.guard.Unlock()

	if len(s.expires) == 0 {
		return
	}
	for _, timer := range s.expires {
		s.C <- timer
	}
	if len(s.expires) > TimerCapacity {
		s.expires = make([]*timerNode, 0, 4)
	} else {
		s.expires = s.expires[:0]
	}
}

// 返回触发的timer列表
func (s *TimeoutScheduler) trigger(now time.Time) {
	var ts = now.UnixNano() / 1000_000 // to millisecond
	var maxId = s.nextId
	for s.timers.Len() > 0 {
		var node = s.timers[0] // peek first item of heap
		if ts < node.priority {
			return // no timer expired
		}
		// make sure we don't process timer created by timer events
		if node.id > maxId {
			continue
		}

		// 如果timer需要重复执行，只修正heap，id保持不变
		if node.repeatable {
			node.priority = ts + int64(node.interval)
			heap.Fix(&s.timers, node.index)
		} else {
			heap.Pop(&s.timers)
			delete(s.ref, node.id)
		}
		s.expires = append(s.expires, node)
	}
}

func (s *TimeoutScheduler) schedule(ts int64, interval uint32, repeat bool, r Runner) int {
	s.guard.Lock()

	// 假设ID一直自增不会溢出
	var id = s.nextId
	s.nextId++

	var node = &timerNode{
		priority:   ts,
		interval:   interval,
		repeatable: repeat,
		id:         id,
		R:          r,
	}
	heap.Push(&s.timers, node)
	s.ref[id] = node
	s.guard.Unlock()
	return id
}

// 创建一个定时器，在`ts`毫秒时间戳运行`r`
func (s *TimeoutScheduler) RunAt(ts int64, r Runner) int {
	return s.schedule(ts, 0, false, r)
}

// 创建一个定时器，在`interval`毫秒后运行`r`
func (s *TimeoutScheduler) RunAfter(interval int, r Runner) int {
	if interval < 0 {
		interval = 0
	}
	var ts = currentMilliSec() + int64(interval)
	return s.schedule(ts, 0, false, r)
}

// 创建一个定时器，每隔`interval`毫秒运行一次`r`
func (s *TimeoutScheduler) RunEvery(interval int, r Runner) int {
	if interval <= 0 {
		interval = TimerPrecision
	} else if interval > math.MaxInt32 {
		interval = math.MaxInt32
	}
	var ts = currentMilliSec() + int64(interval)
	return s.schedule(ts, uint32(interval), true, r)
}

func (s *TimeoutScheduler) Cancel(id int) bool {
	s.guard.Lock()
	defer s.guard.Unlock()
	if timer, found := s.ref[id]; found {
		delete(s.ref, id)
		heap.Remove(&s.timers, timer.index)
		return true
	}
	return false
}

type timerNode struct {
	id         int    // 唯一ID
	index      int    // 数组索引
	priority   int64  // 优先级（即超时时间戳）
	interval   uint32 // 间隔（毫秒)，最多49.7天
	repeatable bool   // 是否重复
	R          Runner // 待执行runner
}

type TimerHeap []*timerNode

func (q TimerHeap) Len() int {
	return len(q)
}

func (q TimerHeap) Less(i, j int) bool {
	return q[i].priority < q[j].priority
}

func (q TimerHeap) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].index = i
	q[j].index = j
}

func (q *TimerHeap) Push(x interface{}) {
	v := x.(*timerNode)
	v.index = len(*q)
	*q = append(*q, v)
}

func (q *TimerHeap) Pop() interface{} {
	old := *q
	n := len(old)
	if n > 0 {
		v := old[n-1]
		v.index = -1 // for safety
		*q = old[:n-1]
		return v
	}
	return nil
}
