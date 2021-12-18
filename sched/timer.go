// Copyright © 2020 simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

const (
	PendingQueueCapacity = 128  // pending add/delete
	TimeoutQueueCapacity = 1000 // pending timed-out
)

// 定时器
type Timer interface {
	Start()

	// 关闭定时器
	Shutdown()

	// 在`timeUnits`时间后执行`r`
	RunAfter(timeUnits int, r Runnable) int

	// 每隔`interval`时间执行`r`
	RunEvery(interval int, r Runnable) int

	// 取消一个timer
	Cancel(id int) bool

	// 判断timer是否在计划中
	IsPending(id int) bool

	// 超时的待执行runner
	Chan() <-chan Runnable

	// timer数量
	Size() int
}
