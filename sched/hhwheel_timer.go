// Copyright © 2021 simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

import (
	"log"
	"sync"
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
	TVN_BITS    = 6 // time vector level shift bits
	TVR_BITS    = 8 // timer vector shift bits
	TVN_SIZE    = 1 << TVN_BITS
	TVR_SIZE    = 1 << TVR_BITS
	TVN_MASK    = TVN_SIZE - 1
	TVR_MASK    = TVR_SIZE - 1
	WHEEL_LEVEL = 4
)

// Hashed and Hierarchical Timing Wheels
type HHWheelTimer struct {
	done  chan struct{}
	wg    sync.WaitGroup //
	state fatchoy.State  // 运行状态

	tickInterval time.Duration // tick间隔时长
	timeUnit     time.Duration // 时间单位
	startAt      int64         // 启动时间(timeunit)
	lastTime     int64         //
	tickTime     int64         //

	guard  sync.Mutex              // 多线程
	refer  map[int]*WheelTimerNode // O(1)查找
	nextId int                     // ID生成

	pendingAdd chan *WheelTimerNode // 待插入
	pendingDel chan *WheelTimerNode // 待删除
	C          chan Runnable        // 到期的定时器

	currTick int64                                   // 当前tick（时间指针）
	near     [TVR_SIZE]WheelTimerBucket              // 最近的时间轮
	tvec     [WHEEL_LEVEL][TVR_SIZE]WheelTimerBucket // 层级时间轮
}

func NewDefaultHHWheelTimer() Timer {
	return NewHHWheelTimer(time.Millisecond*5, time.Millisecond)
}

func NewHHWheelTimer(tickInterval, timeUnit time.Duration) Timer {
	return new(HHWheelTimer).init(tickInterval, timeUnit)
}

func (t *HHWheelTimer) init(tickInterval, timeUnit time.Duration) *HHWheelTimer {
	t.tickInterval = tickInterval
	t.timeUnit = timeUnit
	t.done = make(chan struct{})
	t.refer = make(map[int]*WheelTimerNode)
	t.C = make(chan Runnable, PendingQueueCapacity)
	t.pendingAdd = make(chan *WheelTimerNode, PendingQueueCapacity)
	t.pendingDel = make(chan *WheelTimerNode, PendingQueueCapacity)
	return t
}

func (t *HHWheelTimer) Size() int {
	t.guard.Lock()
	var size = len(t.refer)
	t.guard.Unlock()
	return size
}

func (t *HHWheelTimer) Chan() <-chan Runnable {
	return t.C
}

func (t *HHWheelTimer) IsScheduled(id int) bool {
	t.guard.Lock()
	var node = t.refer[id]
	t.guard.Unlock()
	return node != nil
}

func (t *HHWheelTimer) Start() {
	switch state := t.state.Get(); state {
	case fatchoy.StateInit:
		if t.state.CAS(fatchoy.StateInit, fatchoy.StateStarted) {
			t.startAt = t.currentTimeUnit()
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

// 创建一个定时器，在`timeUnits`时间后运行`r`
func (t *HHWheelTimer) RunAfter(timeUnits int, r Runnable) int {
	if timeUnits < 0 {
		timeUnits = 0
	}

	t.guard.Lock()
	defer t.guard.Unlock()

	var id = t.nextID()
	var deadline = t.currentTimeUnit() + int64(timeUnits)
	var node = newWheelTimerNode(id, deadline, 0, r)
	t.pendingAdd <- node
	t.refer[id] = node

	return id
}

// 创建一个定时器，每隔`interval`时间运行一次`r`
func (t *HHWheelTimer) RunEvery(interval int, r Runnable) int {
	if interval < 0 {
		interval = 1
	}

	t.guard.Lock()
	defer t.guard.Unlock()

	var id = t.nextID()
	var deadline = t.currentTimeUnit() + int64(interval)
	var node = newWheelTimerNode(id, deadline, int64(interval), r)
	t.pendingAdd <- node
	t.refer[id] = node

	return id
}

func (t *HHWheelTimer) Cancel(id int) bool {
	t.guard.Lock()
	defer t.guard.Unlock()

	if node, found := t.refer[id]; found {
		t.pendingDel <- node
		delete(t.refer, id)
		return true
	}
	return false
}

// 当前时间
func (t *HHWheelTimer) currentTimeUnit() int64 {
	return time.Now().UnixNano() / int64(t.timeUnit)
}

func (t *HHWheelTimer) convTimeUnit(tm time.Time) int64 {
	return tm.UnixNano() / int64(t.timeUnit)
}

func (t *HHWheelTimer) nextID() int {
	var newId = t.nextId + 1
	for i := 0; i < 1e4; i++ {
		if newId <= 0 {
			newId = 1
		}
		if _, found := t.refer[newId]; found {
			newId++
			continue
		}
		break
	}
	t.nextId = newId
	return newId
}

func (t *HHWheelTimer) worker(ready chan struct{}) {
	defer t.wg.Done()

	var ticker = time.NewTicker(t.tickInterval)
	defer ticker.Stop()
	t.lastTime = t.currentTimeUnit()
	t.tickTime = t.lastTime
	ready <- struct{}{}

	for {
		select {
		case now := <-ticker.C:
			var current = t.convTimeUnit(now)
			t.update(current)

		case node := <-t.pendingAdd:
			t.addNode(node)

		case node := <-t.pendingDel:
			t.delTimer(node)

		case <-t.done:
			return
		}
	}
}

func (t *HHWheelTimer) update(current int64) {
	if current < t.lastTime {
		log.Printf("time gone backwards %d -> %d", t.lastTime, current)
		t.lastTime = current
		return
	}
	if current > t.lastTime {
		var diff = current - t.lastTime
		for i := 0; i < int(diff); i++ {
			t.tick()
		}
		t.lastTime = current
	}
}

func (t *HHWheelTimer) tick() {
	var index = int(t.currTick & TVR_MASK)
	if index == 0 {
		if t.cascade(0) == 0 &&
			t.cascade(1) == 0 &&
			t.cascade(2) == 0 {
			t.cascade(3)
		}
	}
	t.currTick++
	t.tickTime++
	t.expireNear(index)
}

func (t *HHWheelTimer) addNode(node *WheelTimerNode) {
	var ticks = node.deadline - t.tickTime
	if ticks < 0 {
		ticks = 0
	}
	var idx = t.currTick + ticks
	var bucket *WheelTimerBucket
	if idx < TVR_SIZE {
		var i = idx & TVR_MASK
		log.Printf("timer %d tick %d add to near %d\n", node.id, idx, i)
		bucket = &t.near[i]
	} else if ticks < 1<<(TVR_BITS+TVN_BITS) {
		idx = (idx >> (TVR_BITS)) & TVN_MASK
		bucket = &t.tvec[0][idx]
	} else if ticks < 1<<(TVR_BITS+2*TVN_BITS) {
		idx = (idx >> (TVR_BITS + TVN_BITS)) & TVN_MASK
		bucket = &t.tvec[1][idx]
	} else if ticks < 1<<(TVR_BITS+3*TVN_BITS) {
		idx = (idx >> (TVR_BITS + 2*TVN_BITS)) & TVN_MASK
		bucket = &t.tvec[2][idx]
	} else {
		idx = (idx >> (TVR_BITS + 3*TVN_BITS)) & TVN_MASK
		bucket = &t.tvec[3][idx]
	}
	bucket.addNode(node)
}

func (t *HHWheelTimer) delTimer(node *WheelTimerNode) {
	node.bucket.removeNode(node)
}

func (t *HHWheelTimer) cascade(n int) uint32 {
	var idx = uint32(t.currTick>>(TVR_BITS+n*TVN_BITS)) & TVN_MASK
	var node = t.tvec[n][idx].replaceInit()
	for node != nil {
		var next = node.next
		node.unchain()
		t.addNode(node)
		node = next
	}
	return idx
}

func (t *HHWheelTimer) expireNear(index int) {
	var node = t.near[index].replaceInit()
	for node != nil {
		var next = node.next
		node.unchain()

		t.C <- node.r // trigger

		// schedule again
		if node.period > 0 {
			node.deadline = t.currentTimeUnit() + node.period
			t.addNode(node)
		} else {
			t.guard.Lock()
			delete(t.refer, node.id)
			t.guard.Unlock()
		}
		node = next
	}
}

type WheelTimerNode struct {
	next, prev *WheelTimerNode
	bucket     *WheelTimerBucket

	id       int
	deadline int64    // 到期时间
	period   int64    // 间隔
	r        Runnable // 到期任务
}

func newWheelTimerNode(id int, deadline, period int64, r Runnable) *WheelTimerNode {
	return &WheelTimerNode{
		id:       id,
		deadline: deadline,
		period:   period,
		r:        r,
	}
}

func (n *WheelTimerNode) unchain() {
	n.next = nil
	n.prev = nil
	n.bucket = nil
}

type WheelTimerBucket struct {
	head, tail *WheelTimerNode
	size       int
}

func (b *WheelTimerBucket) addNode(node *WheelTimerNode) {
	if node.bucket != nil {
		log.Panicf("wheel node %d bucket is not nil", node.id)
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
		log.Panicf("wheel node %d not belong to this bucket %p != %p", node.id, node.bucket, b)
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

func (b *WheelTimerBucket) replaceInit() *WheelTimerNode {
	var node = b.head
	b.head = nil
	b.tail = nil
	b.size = 0
	return node
}
