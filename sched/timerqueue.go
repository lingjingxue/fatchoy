// Copyright © 2020 simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

import (
	"container/heap"
	"log"
	"sync"
	"time"

	"qchen.fun/fatchoy"
)

// 最小堆实现的定时器
// 标准库的四叉堆实现的time.Timer已经可以满足大多数高精度的定时需求
// 这个实现主要是为了在大量timer的场景，把timer的压力从runtime放到应用上
type TimerQueue struct {
	done         chan struct{}
	wg           sync.WaitGroup     //
	state        fatchoy.State      //
	tickInterval time.Duration      // tick间隔
	timeUnit     time.Duration      // 时间单位
	guard        sync.Mutex         // 多线程
	lastId       int                // id生成
	pending      chan *timerNode    //
	timers       timerHeap          // 二叉最小堆
	refer        map[int]*timerNode // O(1)查找
	C            <-chan Runnable    // 到期的定时器
	startedAt    int64              //
}

func NewDefaultTimerQueue() Timer {
	return NewTimerQueue(time.Millisecond*10, time.Millisecond)
}

func NewTimerQueue(tickInterval, timeUnit time.Duration) Timer {
	t := &TimerQueue{
		tickInterval: tickInterval,
		timeUnit:     timeUnit,
		done:         make(chan struct{}),
		timers:       make(timerHeap, 0, 64),
		refer:        make(map[int]*timerNode, 64),
		pending:      make(chan *timerNode, 1000),
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

func (s *TimerQueue) IsScheduled(id int) bool {
	s.guard.Lock()
	var node = s.refer[id]
	s.guard.Unlock()
	return node != nil
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

// 当前时间
func (s *TimerQueue) currentTimeUnit() int64 {
	return time.Now().UnixNano() / int64(s.timeUnit)
}

func (s *TimerQueue) convTimeUnit(t time.Time) int64 {
	return t.UnixNano() / int64(s.timeUnit)
}

// 创建一个定时器，在`timeUnits`时间后运行`r`
func (s *TimerQueue) RunAfter(timeUnits int, r Runnable) int {
	if timeUnits < 0 {
		timeUnits = 0
	}
	var deadline = s.currentTimeUnit() + int64(timeUnits)
	return s.schedule(deadline, 0, r)
}

// 创建一个定时器，每隔`interval`时间运行一次`r`
func (s *TimerQueue) RunEvery(interval int, r Runnable) int {
	if interval < 0 {
		interval = 1
	}
	var deadline = s.currentTimeUnit() + int64(interval)
	return s.schedule(deadline, int64(interval), r)
}

// 取消一个timer
func (s *TimerQueue) Cancel(id int) bool {
	s.guard.Lock()
	defer s.guard.Unlock()

	if node, found := s.refer[id]; found {
		delete(s.refer, id)
		heap.Remove(&s.timers, node.index)
		node.r = nil
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

	var ticker = time.NewTicker(s.tickInterval)
	defer ticker.Stop()

	var expired = make(chan Runnable, 1000)
	s.startedAt = s.currentTimeUnit()
	s.C = expired
	ready <- struct{}{}

	for {
		select {
		case now, ok := <-ticker.C:
			if ok {
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
		if node.r != nil {
			expired <- node.r
		}
	}
}

// 返回触发的timer列表
func (s *TimerQueue) trigger(t time.Time) []*timerNode {
	var deadline = s.convTimeUnit(t)
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
		if node.period > 0 {
			node.deadline = deadline + node.period
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

func (s *TimerQueue) schedule(deadline, period int64, r Runnable) int {
	s.guard.Lock()
	defer s.guard.Unlock()

	var id = s.nextID()
	var node = newTimerNode(id, deadline, period, r)
	heap.Push(&s.timers, node)
	s.refer[id] = node
	return id
}

// 二叉堆节点
type timerNode struct {
	id       int      // unique id
	index    int      // array index of heap
	deadline int64    // Next execution time for this task in milliseconds
	period   int64    // Period in milliseconds for repeating tasks
	r        Runnable //
}

func newTimerNode(id int, deadline, period int64, r Runnable) *timerNode {
	return &timerNode{
		id:       id,
		deadline: deadline,
		period:   period,
		r:        r,
	}
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
