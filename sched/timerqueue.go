// Copyright © 2020 simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

import (
	"container/heap"
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"qchen.fun/fatchoy"
)

// 最小堆实现的定时器
// 标准库的四叉堆实现的time.Timer已经可以满足大多数高精度的定时需求
// 这个实现主要是为了在大量timer的场景，把timer的压力从runtime放到应用上
type TimerQueue struct {
	done      chan struct{}
	wg        sync.WaitGroup     //
	state     fatchoy.State      //
	ticks     int64              //
	guard     sync.Mutex         // 多线程
	timers    timerHeap          // 二叉最小堆
	lastId    int                // id生成
	refer     map[int]*timerNode // O(1)查找
	C         <-chan Runnable    // 到期的定时器
	startedAt int64              //
}

func NewTimerQueue() Timer {
	t := &TimerQueue{
		done:   make(chan struct{}, 1),
		timers: make(timerHeap, 0, 64),
		refer:  make(map[int]*timerNode, 64),
	}
	t.start()
	return t
}

func (s *TimerQueue) Size() int {
	s.guard.Lock()
	var n = len(s.timers)
	s.guard.Unlock()
	return n
}

func (s *TimerQueue) Chan() <-chan Runnable {
	return s.C
}

func (s *TimerQueue) Shutdown() {
	switch s.state.Get() {
	case fatchoy.StateShutdown, fatchoy.StateTerminated:
		return
	}

	s.state.Set(fatchoy.StateShutdown)
	close(s.done)
	s.wg.Wait()

	s.guard.Lock()
	s.C = nil
	s.refer = nil
	s.timers = nil
	s.guard.Unlock()

	s.state.Set(fatchoy.StateTerminated)
}

// 创建一个定时器，在`durationMs`毫秒后运行`task`
func (s *TimerQueue) RunAfter(durationMs int, task Runnable) int {
	if durationMs >= math.MaxInt32 {
		log.Panicf("duration %d out of range", durationMs)
		return -1
	}
	if durationMs < 0 {
		durationMs = 0
	}
	var ts = currentMs() + int64(durationMs)
	return s.schedule(ts, 0, false, task)
}

// 创建一个定时器，每隔`interval`毫秒运行一次`task`
func (s *TimerQueue) RunEvery(intervalMs int, task Runnable) int {
	if intervalMs >= math.MaxInt32 {
		log.Panicf("interval %d out of range", intervalMs)
		return -1
	}
	if intervalMs < 0 {
		intervalMs = TimeUnit
	}
	var ts = currentMs() + int64(intervalMs)
	return s.schedule(ts, int32(intervalMs), true, task)
}

// 取消一个timer
func (s *TimerQueue) Cancel(id int) bool {
	s.guard.Lock()
	defer s.guard.Unlock()

	if node, found := s.refer[id]; found {
		delete(s.refer, id)
		heap.Remove(&s.timers, node.index)
		node.task = nil
		return true
	}
	return false
}

// Starts the background thread explicitly
func (s *TimerQueue) start() {
	switch state := s.state.Get(); state {
	case fatchoy.StateInit:
		if s.state.CAS(fatchoy.StateInit, fatchoy.StateStarted) {
			var ready = make(chan struct{}, 1)
			s.wg.Add(1)
			go s.worker(ready)
			<-ready
			s.state.Set(fatchoy.StateRunning)
		}

	case fatchoy.StateRunning:
		return

	default:
		log.Panicf("invalid worker state %v", state)
	}
}

func (s *TimerQueue) worker(ready chan struct{}) {
	defer s.wg.Done()

	var ticker = time.NewTicker(TimeUnit * time.Millisecond)
	defer ticker.Stop()

	var expired = make(chan Runnable, 1000)
	s.startedAt = currentMs()
	s.C = expired
	ready <- struct{}{}

	for {
		select {
		case now, ok := <-ticker.C:
			if ok {
				atomic.AddInt64(&s.ticks, 1)
				s.tick(now, expired)
			}

		case <-s.done:
			return
		}
	}
}

func (s *TimerQueue) tick(deadline time.Time, expired chan<- Runnable) {
	s.guard.Lock()
	var expires = s.trigger(deadline)
	s.guard.Unlock()

	for _, node := range expires {
		if node.task != nil {
			expired <- node.task
		}
	}
}

// 返回触发的timer列表
func (s *TimerQueue) trigger(t time.Time) []*timerNode {
	var deadline = timeMs(t)
	var maxId = s.lastId
	var expires []*timerNode
	for len(s.timers) > 0 {
		var node = s.timers[0] // peek first item of heap
		if deadline < node.deadline {
			break // no new timer expired
		}
		// make sure we don't process timer created by timer events
		if node.id > maxId {
			continue
		}

		// 如果timer需要重复执行，只修正heap，id保持不变
		if node.repeatable {
			node.deadline = deadline + int64(node.interval)
			heap.Fix(&s.timers, node.index)
		} else {
			heap.Pop(&s.timers)
			delete(s.refer, node.id)
		}
		expires = append(expires, node)
	}
	return expires
}

func (s *TimerQueue) nextID() int {
	var newId = s.lastId + 1
	for i := 0; i < 1e4; i++ {
		if newId <= 0 {
			newId = 1
		}
		if _, found := s.refer[newId]; found {
			newId++
			continue
		}
		break
	}
	s.lastId = newId
	return newId
}

func (s *TimerQueue) schedule(ts int64, interval int32, repeat bool, task Runnable) int {
	s.guard.Lock()
	defer s.guard.Unlock()

	var id = s.nextID()
	var node = &timerNode{
		deadline:   ts,
		interval:   interval,
		repeatable: repeat,
		id:         id,
		task:       task,
	}
	heap.Push(&s.timers, node)
	s.refer[id] = node
	return id
}

// 二叉堆节点
type timerNode struct {
	id         int      // 唯一ID
	index      int      // 数组索引
	deadline   int64    // 到期时间
	interval   int32    // 间隔（毫秒)，最多24.8天
	repeatable bool     // 是否重复
	task       Runnable // 触发任务
}

type timerHeap []*timerNode

func (q timerHeap) Len() int {
	return len(q)
}

func (q timerHeap) Less(i, j int) bool {
	if q[i].deadline == q[j].deadline {
		return q[i].id > q[j].id
	}
	return q[i].deadline < q[j].deadline
}

func (q timerHeap) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].index = i
	q[j].index = j
}

func (q *timerHeap) Push(x interface{}) {
	v := x.(*timerNode)
	v.index = len(*q)
	*q = append(*q, v)
}

func (q *timerHeap) Pop() interface{} {
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
