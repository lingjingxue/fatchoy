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
)

// 最小堆实现的定时器
// 标准库的四叉堆实现的time.Timer已经可以满足大多数高精度的定时需求
// 这个实现主要是为了在大量timer的场景，把timer的压力从runtime放到应用上
type TimerQueue struct {
	done      chan struct{}
	wg        sync.WaitGroup          //
	state     int32                   // worker state
	ticks     int64                   //
	guard     sync.Mutex              // 多线程
	timers    timerHeap               // 二叉最小堆
	lastId    int                     // id生成
	refer     map[int]*timerQueueNode // O(1)查找
	C         <-chan Runnable         // 到期的定时器
	startedAt int64                   //
}

func NewTimerQueue() *TimerQueue {
	return &TimerQueue{
		done:   make(chan struct{}, 1),
		timers: make(timerHeap, 0, 64),
		refer:  make(map[int]*timerQueueNode, 64),
	}
}

func (s *TimerQueue) Size() int {
	s.guard.Lock()
	var n = len(s.timers)
	s.guard.Unlock()
	return n
}

func (s *TimerQueue) Shutdown() {
	close(s.done)
	s.wg.Wait()

	s.guard.Lock()
	s.C = nil
	s.refer = nil
	s.timers = nil
	s.guard.Unlock()
}

// 创建一个定时器，在`ts`毫秒时间戳运行`task`
func (s *TimerQueue) RunAt(ts int64, task Runnable) int {
	var now = currentMs()
	if ts < now {
		ts = now
	}
	return s.schedule(ts, 0, false, task)
}

// 创建一个定时器，在`interval`毫秒后运行`task`
func (s *TimerQueue) RunAfter(interval int, task Runnable) int {
	if interval >= math.MaxInt32 {
		log.Panicf("interval %d out of range", interval)
		return -1
	}
	if interval < 0 {
		interval = 0
	}
	var ts = currentMs() + int64(interval)
	return s.schedule(ts, 0, false, task)
}

// 创建一个定时器，每隔`interval`毫秒运行一次`task`
func (s *TimerQueue) RunEvery(interval int, task Runnable) int {
	if interval >= math.MaxInt32 {
		log.Panicf("interval %d out of range", interval)
		return -1
	}
	if interval < 0 {
		interval = int(TimeUnit)
	}
	var ts = currentMs() + int64(interval)
	return s.schedule(ts, int32(interval), true, task)
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
	var state = atomic.LoadInt32(&s.state)
	switch state {
	case WorkerInit:
		if atomic.CompareAndSwapInt32(&s.state, WorkerInit, WorkerStarted) {
			var ready = make(chan struct{}, 1)
			s.wg.Add(1)
			go s.worker(ready)
			<-ready
		}
	case WorkerStarted:
		return

	default:
		log.Panicf("invalid worker state %v", state)
	}
}

func (s *TimerQueue) worker(ready chan struct{}) {
	defer func() {
		atomic.StoreInt32(&s.state, WorkerShutdown)
		s.wg.Done()
	}()

	var ticker = time.NewTicker(TimeUnit)
	defer ticker.Stop()

	var bus = make(chan Runnable, 1000)
	s.startedAt = currentMs()
	s.C = bus
	ready <- struct{}{}

	for {
		select {
		case now, ok := <-ticker.C:
			if ok {
				atomic.AddInt64(&s.ticks, 1)
				s.tick(now, bus)
			}

		case <-s.done:
			return
		}
	}
}

func (s *TimerQueue) tick(deadline time.Time, bus chan<- Runnable) {
	s.guard.Lock()
	var expires = s.trigger(deadline)
	s.guard.Unlock()

	for _, node := range expires {
		if node.task != nil {
			bus <- node.task
		}
	}
}

// 返回触发的timer列表
func (s *TimerQueue) trigger(t time.Time) []*timerQueueNode {
	var deadline = timeMs(t)
	var maxId = s.lastId
	var expires []*timerQueueNode
	for len(s.timers) > 0 {
		var node = s.timers[0] // peek first item of heap
		if deadline < node.expiry {
			break // no new timer expired
		}
		// make sure we don't process timer created by timer events
		if node.id > maxId {
			continue
		}

		// 如果timer需要重复执行，只修正heap，id保持不变
		if node.repeatable {
			node.expiry = deadline + int64(node.interval)
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
	s.start()

	s.guard.Lock()
	defer s.guard.Unlock()

	var id = s.nextID()
	var node = &timerQueueNode{
		expiry:     ts,
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
type timerQueueNode struct {
	id         int      // 唯一ID
	index      int      // 数组索引
	expiry     int64    // 到期时间
	interval   int32    // 间隔（毫秒)，最多24.8天
	repeatable bool     // 是否重复
	task       Runnable // 触发任务
}

type timerHeap []*timerQueueNode

func (q timerHeap) Len() int {
	return len(q)
}

func (q timerHeap) Less(i, j int) bool {
	return q[i].expiry < q[j].expiry
}

func (q timerHeap) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].index = i
	q[j].index = j
}

func (q *timerHeap) Push(x interface{}) {
	v := x.(*timerQueueNode)
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
