// Copyright © 2020 simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package xtimer

import (
	"container/heap"
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"qchen.fun/fatchoy/l0g"
)

const (
	TimeUnit    = 10 * time.Millisecond           //
	CustomEpoch = int64(1577836800 * time.Second) // 起始纪元 2020-01-01 00:00:00 UTC
)

// 定时器回调函数
type TimerHandle func()

// 最小堆实现的定时器
// 标准库的四叉堆实现的time.Timer已经可以满足大多数高精度的定时需求
// 这个实现主要是为了在大量timer的场景，把timer的压力从runtime放到应用上
type TimerQueue struct {
	done   chan struct{}
	wg     sync.WaitGroup     //
	ticks  int64              //
	guard  sync.Mutex         // 多线程
	timers timerHeap          // 二叉最小堆
	nextId int                // id生成
	ref    map[int]*queueNode // O(1)查找
	C      chan TimerHandle   // 到期的定时器
}

func NewTimerQueue() *TimerQueue {
	return &TimerQueue{
		done:   make(chan struct{}, 1),
		timers: make(timerHeap, 0, 64),
		ref:    make(map[int]*queueNode, 64),
		C:      make(chan TimerHandle, 1000),
	}
}

func (s *TimerQueue) Start() {
	s.wg.Add(1)
	go s.serve()
}

func (s *TimerQueue) Size() int {
	return len(s.timers)
}

func (s *TimerQueue) Shutdown() {
	close(s.done)
	s.wg.Wait()

	s.C = nil
	s.ref = nil
	s.timers = nil
}

func (s *TimerQueue) serve() {
	l0g.Debugf("scheduler start serving")
	defer s.wg.Done()
	var ticker = time.NewTicker(TimeUnit)
	defer ticker.Stop()

	for {
		select {
		case now, ok := <-ticker.C:
			if ok {
				s.update(now)
			}

		case <-s.done:
			return
		}
	}
}

func (s *TimerQueue) update(now time.Time) {
	atomic.AddInt64(&s.ticks, 1)

	s.guard.Lock()
	defer s.guard.Unlock()

	var expires = s.trigger(now)
	for _, node := range expires {
		if node.cb != nil {
			s.C <- node.cb
		}
	}
}

// 返回触发的timer列表
func (s *TimerQueue) trigger(now time.Time) []*queueNode {
	var ts = nowMs(now)
	var maxId = s.nextId
	var expires []*queueNode
	for len(s.timers) > 0 {
		var node = s.timers[0] // peek first item of heap
		if ts < node.expiry {
			break // no new timer expired
		}
		// make sure we don't process timer created by timer events
		if node.id > maxId {
			continue
		}

		// 如果timer需要重复执行，只修正heap，id保持不变
		if node.repeatable {
			node.expiry = ts + int64(node.interval)
			heap.Fix(&s.timers, node.index)
		} else {
			heap.Pop(&s.timers)
			delete(s.ref, node.id)
		}
		expires = append(expires, node)
	}
	return expires
}

func (s *TimerQueue) schedule(ts int64, interval int32, repeat bool, cb TimerHandle) int {
	s.guard.Lock()
	defer s.guard.Unlock()

	var id = s.nextId
	s.nextId++ // 假设ID一直自增不会溢出

	var node = &queueNode{
		expiry:     ts,
		interval:   interval,
		repeatable: repeat,
		id:         id,
		cb:         cb,
	}
	heap.Push(&s.timers, node)
	s.ref[id] = node
	return id
}

// 创建一个定时器，在`ts`毫秒时间戳运行`cb`
func (s *TimerQueue) RunAt(ts int64, cb TimerHandle) int {
	var now = currentMs()
	if ts < now {
		ts = now
	}
	return s.schedule(ts, 0, false, cb)
}

// 创建一个定时器，在`interval`毫秒后运行`cb`
func (s *TimerQueue) RunAfter(interval int, cb TimerHandle) int {
	if interval >= math.MaxInt32 {
		log.Panicf("interval %d out of range", interval)
		return -1
	}
	if interval < 0 {
		interval = 0
	}
	var ts = currentMs() + int64(interval)
	return s.schedule(ts, 0, false, cb)
}

// 创建一个定时器，每隔`interval`毫秒运行一次`cb`
func (s *TimerQueue) RunEvery(interval int, cb TimerHandle) int {
	if interval >= math.MaxInt32 {
		log.Panicf("interval %d out of range", interval)
		return -1
	}
	if interval < 0 {
		interval = 0
	}

	var ts = currentMs() + int64(interval)
	return s.schedule(ts, int32(interval), true, cb)
}

func (s *TimerQueue) Cancel(id int) bool {
	s.guard.Lock()
	defer s.guard.Unlock()

	if node, found := s.ref[id]; found {
		delete(s.ref, id)
		heap.Remove(&s.timers, node.index)
		node.cb = nil
		return true
	}
	return false
}

// 二叉堆节点
type queueNode struct {
	id         int         // 唯一ID
	index      int         // 数组索引
	expiry     int64       // 到期时间
	interval   int32       // 间隔（毫秒)，最多24.8天
	repeatable bool        // 是否重复
	cb         TimerHandle // 超时回调函数
}

type timerHeap []*queueNode

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
	v := x.(*queueNode)
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

// 当前毫秒
func currentMs() int64 {
	return (time.Now().UnixNano() - CustomEpoch) / int64(time.Millisecond)
}

// 转为当前毫秒
func nowMs(now time.Time) int64 {
	return (now.UnixNano() - CustomEpoch) / int64(time.Millisecond)
}
