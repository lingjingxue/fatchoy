// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package stats

import (
	"sync/atomic"
)

// 一组计数器
type Stats struct {
	counters []int64
}

func New(n int) *Stats {
	return &Stats{counters: make([]int64, n)}
}

func (s *Stats) Get(i int) int64 {
	if i >= 0 && i < len(s.counters) {
		return atomic.LoadInt64(&s.counters[i])
	}
	return 0
}

func (s *Stats) Set(i int, v int64) {
	if i >= 0 && i < len(s.counters) {
		atomic.StoreInt64(&s.counters[i], v)
	}
}

func (s *Stats) Add(i int, delta int64) int64 {
	if i >= 0 && i < len(s.counters) {
		return atomic.AddInt64(&s.counters[i], delta)
	}
	return 0
}

func (s *Stats) Copy() []int64 {
	arr := make([]int64, len(s.counters))
	for i := 0; i < len(arr); i++ {
		arr[i] = atomic.LoadInt64(&s.counters[i])
	}
	return arr
}

func (s *Stats) Clone() *Stats {
	return &Stats{counters: s.Copy()}
}
