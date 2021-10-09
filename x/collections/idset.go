// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package collections

import (
	"sort"
)

// 用数组实现的有序数字集合，仅用于存储少量数据且查询多过插入/删除的场合
type OrderedIDSet []int32

// 查询使用二分查找
func (s OrderedIDSet) Find(id int32) int {
	return sort.Search(len(s), func(i int) bool {
		return s[i] >= id
	})
}

func (s OrderedIDSet) Has(id int32) bool {
	i := s.Find(id)
	return i < len(s) && s[i] == id
}

// 插入后保持有序
func (s OrderedIDSet) Insert(id int32) OrderedIDSet {
	i := s.Find(id)
	var n = len(s)
	if i < n && s[i] == id {
		return s // 已经存在
	}
	if i >= n {
		return append(s, id)
	}
	s = append(s, 0) // add an empty value
	copy(s[i+1:], s[i:])
	s[i] = id
	return s
}

// 删除
func (s OrderedIDSet) Delete(id int32) OrderedIDSet {
	for i, n := range s {
		if n == id {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

// 无序整数集合，用于存储少量数据且插入/删除多于查询的场合
type IDSet []int32

func (s IDSet) Find(id int32) int {
	for i, n := range s {
		if n == id {
			return i
		}
	}
	return -1
}

func (s IDSet) Has(id int32) bool {
	return s.Find(id) >= 0
}

// 插入，需要提前判断是否已经存在
func (s IDSet) Insert(id int32) IDSet {
	return append(s, id)
}

func (s IDSet) PutIfAbsent(id int32) IDSet {
	if !s.Has(id) {
		return append(s, id)
	}
	return s
}

// 删除
func (s IDSet) Delete(id int32) IDSet {
	i := s.Find(id)
	if i >= 0 {
		n := len(s)
		s[i], s[n-1] = s[n-1], s[i] // 与最后一个元素交换
		return s[:n-1]
	}
	return s
}
