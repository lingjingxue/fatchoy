// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"fmt"
	"log"
	"sync"
)

const (
	DefaultSeqStep = 2000 // 默认每个counter分配2000个号
)

// 使用发号器算法的uuid生成器
// 1. 算法把一个64位的整数按step范围划分为N个号段；
// 2. service从发号器拿到号段后才可分配此号段内的ID;
// 3. 发号器依赖存储(etcd, redis)把当前待分配号段持续自增；
type SeqID struct {
	guard   sync.Mutex //
	store   Storage    // 存储组件
	step    int64      // 号段区间
	counter int64      // 当前号段
	lastID  int64      // 上次生成的ID
}

func NewSeqID(store Storage, step int32) *SeqID {
	if step <= 0 {
		step = DefaultSeqStep
	}
	return &SeqID{
		store: store,
		step:  int64(step),
	}
}

func (s *SeqID) Init() error {
	if err := s.reload(); err != nil {
		return err
	}
	return nil
}

func (s *SeqID) reload() error {
	counter, err := s.store.Incr()
	if err != nil {
		return err
	}
	s.counter = counter
	s.lastID = counter * s.step
	var rangeEnd = (counter + 1) * s.step
	if rangeEnd < s.lastID {
		return fmt.Errorf("SeqID: integer overflow: %d -> %d", s.lastID, rangeEnd)
	}
	return nil
}

func (s *SeqID) Next() (int64, error) {
	s.guard.Lock()
	defer s.guard.Unlock()

	var next = s.lastID + 1
	var rangEnd = (s.counter + 1) * s.step

	// 在当前号段内，直接分配
	if next <= rangEnd {
		s.lastID = next
		return next, nil
	}

	// 需要重新申请号段
	if err := s.reload(); err != nil {
		return 0, err
	}
	next = s.lastID + 1
	s.lastID = next
	return next, nil
}

func (s *SeqID) MustNext() int64 {
	n, err := s.Next()
	if err != nil {
		log.Panicf("next ID: %v", err)
	}
	return n
}
