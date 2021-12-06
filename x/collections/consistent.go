// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package collections

import (
	"fmt"
	"sort"
)

const (
	ReplicaCount = 20 // 虚拟节点数量
)

// 一致性hash
type Consistent struct {
	circle     map[uint32]string // hash环
	nodes      map[string]bool   // 所有节点
	sortedHash []uint32          // 环hash排序
}

func NewConsistent() *Consistent {
	return &Consistent{
		circle: make(map[uint32]string),
		nodes:  make(map[string]bool),
	}
}

// fnv hash
func (c *Consistent) hashKey(key string) uint32 {
	// see src/hash/fnv.go sum32a.Write
	var hash = uint32(2166136261)
	for i := 0; i < len(key); i++ {
		var c = byte(key[i])
		hash ^= uint32(c)
		hash *= 16777619
	}
	return hash
}

// 添加一个节点
func (c *Consistent) AddNode(node string) {
	for i := 0; i < ReplicaCount; i++ {
		var replica = fmt.Sprintf("%s-%d", node, i)
		c.circle[c.hashKey(replica)] = node
	}
	c.nodes[node] = true
	c.updateSortedHash()
}

func (c *Consistent) RemoveNode(node string) {
	for i := 0; i < ReplicaCount; i++ {
		var replica = fmt.Sprintf("%s-%d", node, i)
		var key = c.hashKey(replica)
		delete(c.circle, key)
	}
	delete(c.nodes, node)
	c.updateSortedHash()
}

// 获取一个节点
func (c *Consistent) GetNodeBy(key string) string {
	var i = c.search(c.hashKey(key))
	var node = c.circle[c.sortedHash[i]]
	return node
}

// 找到第一个大于等于`hash`的节点
func (c *Consistent) search(hash uint32) int {
	var lo = 0
	var hi = len(c.sortedHash)
	for lo < hi {
		var mid = lo + (hi-lo)/2
		if c.sortedHash[mid] <= hash {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	if lo >= len(c.sortedHash) {
		lo = 0
	}
	return lo
}

func (c *Consistent) updateSortedHash() {
	hashes := c.sortedHash[:0]
	// 使用率低于1/4重新分配内存
	if cap(c.sortedHash)/(ReplicaCount*4) > len(c.circle) {
		hashes = nil
	}
	for k, _ := range c.circle {
		hashes = append(hashes, k)
	}
	sort.Sort(Uint32Array(hashes))
	c.sortedHash = hashes
}
