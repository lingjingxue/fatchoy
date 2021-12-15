// Copyright © 2021 simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

import (
	"fmt"
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"qchen.fun/fatchoy"
)

// timer queue implemented with hashed hierarchical wheel.
//
// inspired by linux kernel, see links below
// https://git.kernel.org/pub/scm/linux/kernel/git/stable/linux.git/tree/kernel/timer.c?h=v3.2.98
//
// We model timers as the number of ticks until the next
// due event. We allow 32-bits of space to track this
// due interval, and break that into 4 regions of 8 bits.
// Each region indexes into a bucket of 256 lists.

const (
	TVN_BITS = 6 // time vector level shift bits
	TVR_BITS = 8 // timer vector shift bits
	TVN_SIZE = 1 << TVN_BITS
	TVR_SIZE = 1 << TVR_BITS
	TVN_MASK = TVN_SIZE - 1
	TVR_MASK = TVR_SIZE - 1

	MAX_TVAL = TVN_BITS * TVN_BITS * TVN_BITS * TVN_BITS * TVR_BITS * int64(TimeUnit)
)

// Hashed and Hierarchical Timing Wheels
type HHWheelTimer struct {
	done     chan struct{}
	wg       sync.WaitGroup
	state    fatchoy.State
	startAt  int64
	lastTime int64
	size     int
	nextId   int32
	currTick uint32
	near     [TVR_SIZE]WheelTimerBucket
	tvec     [4][TVR_SIZE]WheelTimerBucket
	addQueue chan *WheelTimerNode
	delQueue chan *WheelTimerNode
	C        <-chan Runnable // 到期的定时器
	guard    sync.Mutex
	refer    map[int]*WheelTimerNode
}

func NewHHWheelTimer() Timer {
	t := &HHWheelTimer{
		startAt:  currentMs(),
		lastTime: currentMs(),
		refer:    make(map[int]*WheelTimerNode),
		done:     make(chan struct{}),
		addQueue: make(chan *WheelTimerNode, 1024),
		delQueue: make(chan *WheelTimerNode, 1024),
	}
	t.start()
	return t
}

func (t *HHWheelTimer) Size() int {
	t.guard.Lock()
	var size = t.size
	t.guard.Unlock()
	return size
}

func (t *HHWheelTimer) Chan() <-chan Runnable {
	return t.C
}

func (t *HHWheelTimer) Shutdown() {
	switch t.state.Get() {
	case fatchoy.StateShutdown, fatchoy.StateTerminated:
		return
	}
	t.state.Set(fatchoy.StateShutdown)
	close(t.done)
	t.wg.Wait()

	t.C = nil
	t.refer = nil
	t.state.Set(fatchoy.StateTerminated)
}

// 创建一个定时器，在`interval`毫秒后运行`task`
func (t *HHWheelTimer) RunAfter(interval int, task Runnable) int {
	if int64(interval) >= MAX_TVAL {
		log.Panicf("timer interval %d out of range", interval)
		return -1
	}

	var id = int(atomic.AddInt32(&t.nextId, 1))
	var node = &WheelTimerNode{
		id:       id,
		task:     task,
		deadline: currentMs() + int64(interval),
	}
	t.addQueue <- node

	return id
}

// 创建一个定时器，每隔`interval`毫秒运行一次`task`
func (t *HHWheelTimer) RunEvery(interval int, task Runnable) int {
	if interval >= math.MaxInt32 {
		log.Panicf("interval %d out of range", interval)
		return -1
	}

	var id = int(atomic.AddInt32(&t.nextId, 1))
	var node = &WheelTimerNode{
		id:         id,
		task:       task,
		repeatable: true,
		interval:   int32(interval),
		deadline:   currentMs() + int64(interval),
	}
	t.addQueue <- node
	return id
}

func (t *HHWheelTimer) Cancel(id int) bool {
	t.guard.Lock()
	defer t.guard.Unlock()

	if node, found := t.refer[id]; found {
		t.delQueue <- node
		return true
	}
	return false
}

func (t *HHWheelTimer) start() {
	switch state := t.state.Get(); state {
	case fatchoy.StateInit:
		if t.state.CAS(fatchoy.StateInit, fatchoy.StateStarted) {
			var ready = make(chan struct{}, 1)
			t.wg.Add(1)
			go t.worker(ready)
			<-ready
			t.state.Set(fatchoy.StateRunning)
		}

	case fatchoy.StateRunning:
		return

	default:
		log.Panicf("invalid worker state %v", state)
	}
}

func (t *HHWheelTimer) worker(ready chan struct{}) {
	defer t.wg.Done()

	var expireChan = make(chan Runnable, 1000)
	var ticker = time.NewTicker(TimeUnit * time.Millisecond)
	defer ticker.Stop()
	t.C = expireChan

	ready <- struct{}{}

	for {
		select {
		case now := <-ticker.C:
			t.tick(timeMs(now), expireChan)

		case node := <-t.addQueue:
			t.addTimer(node)

		case node := <-t.delQueue:
			t.delTimer(node)

		case <-t.done:
			return
		}
	}
}

func (t *HHWheelTimer) addTimer(node *WheelTimerNode) {
	t.addNode(node)

	t.guard.Lock()
	t.refer[node.id] = node
	t.size++
	t.guard.Unlock()
}

func (t *HHWheelTimer) addNode(node *WheelTimerNode) {
	var ticks = (node.deadline - t.startAt) / int64(TimeUnit)
	if ticks < TVR_SIZE {
		var i = ticks & TVR_MASK
		t.near[i].addNode(node)
		return
	}
	for level := 0; level < 4; level++ {
		var n = int64(1 << (TVR_BITS + (level+1)*TVN_BITS))
		if ticks < n {
			var i = (ticks >> (TVR_BITS + (level)*TVN_BITS)) & TVN_MASK
			t.tvec[level][i].addNode(node)
		}
	}
}

func (t *HHWheelTimer) delTimer(node *WheelTimerNode) {
	delete(t.refer, node.id)
	node.bucket.removeNode(node)
	t.size--
}

func (t *HHWheelTimer) cascade(level int) uint32 {
	var idx = (t.currTick >> (TVR_BITS + (level-1)*TVN_BITS)) & TVN_MASK
	var node = t.tvec[level][idx].splice()
	for node != nil {
		var next = node.next
		node.unchain()
		t.addNode(node)
		node = next
	}
	return uint32(idx)
}

func (t *HHWheelTimer) tick(now int64, expireChan chan<- Runnable) {
	var idx = t.currTick & TVR_MASK
	if idx == 0 &&
		t.cascade(1) == 0 &&
		t.cascade(2) == 0 {
		t.cascade(3)
	}
	var node = t.near[idx].splice()
	for node != nil {
		var next = node.next
		node.unchain()
		if node.deadline > now {
			panic("timeout node deadline")
		}
		expireChan <- node.task // trigger
		if node.repeatable {
			node.deadline += int64(node.interval)
			t.addNode(node)
		}
		node = next
	}
	t.currTick++
}

type WheelTimerNode struct {
	next, prev *WheelTimerNode
	bucket     *WheelTimerBucket

	id         int
	deadline   int64    // 到期时间
	interval   int32    // 间隔（毫秒)，最多24.8天
	repeatable bool     // 是否重复
	task       Runnable // 到期任务
}

func (n *WheelTimerNode) unchain() {
	fmt.Printf("nil node %d and its bucket %p\n", n.id, n.bucket)
	n.next = nil
	n.prev = nil
	n.bucket = nil
}

type WheelTimerBucket struct {
	head, tail *WheelTimerNode
	size       int
}

func (b *WheelTimerBucket) addNode(node *WheelTimerNode) {
	fmt.Printf("add node %d to bucket %p\n", node.id, b)
	if node.bucket != nil {
		panic("wheel node bucket is not nil")
	}
	node.bucket = b
	if b.head == nil {
		b.head = node
		b.tail = node
		b.size = 1
	} else {
		b.tail.next = node
		node.prev = b.tail
		b.tail = node
		b.size++
	}
}

func (b *WheelTimerBucket) removeNode(node *WheelTimerNode) *WheelTimerNode {
	if node.bucket != b {
		panic("wheel node not belong to this bucket")
	}
	var next = node.next
	if node.prev != nil {
		node.prev.next = next
	}
	if node.next != nil {
		node.next.prev = node.prev
	}
	if node == b.head {
		if node == b.tail {
			b.head = nil
			b.tail = nil
		} else {
			b.head = next
		}
	} else if node == b.tail {
		b.tail = node.prev
	}
	// unchain from bucket
	node.unchain()
	b.size--
	return next
}

func (b *WheelTimerBucket) splice() *WheelTimerNode {
	var node = b.head
	b.head = nil
	b.tail = nil
	b.size = 0
	return node
}
