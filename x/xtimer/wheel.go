// Copyright © 2020 simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package xtimer

import (
	"sync"
	"time"
)

const (
	SecondsPerMinute = 60
	SecondsPerHour   = 60 * SecondsPerMinute
	SecondsPerDay    = 60 * SecondsPerHour

	MaxWheelDay      = 400 // 最多400天
)

type TimerSlot map[int]*TimerNode

// 用于Timer存档
type TimerNode struct {
	Id       int   // 唯一ID
	Expire   int64 // 超时时间
	Receiver int64 // 接收者（通常是userID)
	Action   int   // 执行的动作
	Param    int   // 参数
}

// 用于低精度（秒）的定时器（主要用于业务中的倒计时）
type TimingWheel struct {
	current int64
	nextId  int
	guard   sync.Mutex
	buckets [3][60]TimerSlot       // 时分秒3级时间轮
	faraway [MaxWheelDay]TimerSlot // 按天划分
}

func NewTimingWheel() *TimingWheel {
	t := &TimingWheel{}
	t.Init()
	return t
}

func (t *TimingWheel) Init() {
	for i, slots := range t.buckets {
		for j := 0; j <= len(slots); j++ {
			t.buckets[i][j] = make(TimerSlot)
		}
	}
	for i := 0; i < len(t.faraway); i++ {
		t.faraway[i] = make(TimerSlot)
	}
}

func (t *TimingWheel) A() {

}

func (t *TimingWheel) addNode(node *TimerNode) {
	var duration = node.Expire - time.Now().Unix()
	if duration < 0 {
		duration = 0
	}
	var slot TimerSlot
	if duration < SecondsPerMinute {
		slot = t.buckets[0][duration]
	} else if duration < SecondsPerHour {
		idx := duration / SecondsPerMinute
		slot = t.buckets[1][idx]
	} else if duration < SecondsPerDay {
		idx := duration / SecondsPerHour
		slot = t.buckets[2][idx]
	} else {
		idx := duration / SecondsPerDay
		slot = t.faraway[idx]
	}
	slot[node.Id] = node
}

func (t *TimingWheel) cascadeTimer() {

}

func (t *TimingWheel) schedule(node *TimerNode) {
	t.guard.Lock()
	defer t.guard.Unlock()

	t.nextId++
	node.Id = t.nextId
	t.addNode(node)
}

func (t *TimingWheel) StartTimer(receiver int64, action, param int, delay int64) int {
	if delay < 0 {
		delay = 0
	}
	if delay >= MaxWheelDay*SecondsPerDay {
		panic("timeout out of range")
		return -1
	}
	var node = &TimerNode{
		Expire:   time.Now().Unix() + delay,
		Receiver: receiver,
		Action:   action,
		Param:    param,
	}
	t.schedule(node)
	return node.Id
}

func (t *TimingWheel) StopTimer(id int) {

}
