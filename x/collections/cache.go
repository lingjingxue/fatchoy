// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package collections

import (
	"container/list"
)

type LRUEntry struct {
	Key   string
	Value interface{}
}

// LRU缓存
// https://en.wikipedia.org/wiki/Cache_replacement_policies#Least_recently_used_(LRU)
type LRUCache struct {
	list     *list.List
	dict     map[string]*list.Element
	capacity int
}

func NewLRUCache(capacity int) *LRUCache {
	if capacity <= 0 {
		panic("cache capacity out of range")
	}
	cache := &LRUCache{
		capacity: capacity,
		list:     list.New(),
		dict:     make(map[string]*list.Element, capacity),
	}
	return cache
}

func (c *LRUCache) Len() int {
	return c.list.Len()
}

func (c *LRUCache) Cap() int {
	return c.capacity
}

// 重置cache
func (c *LRUCache) Reset() {
	for k := range c.dict {
		delete(c.dict, k)
	}
	c.list.Init()
}

//
func (c *LRUCache) removeElement(e *list.Element) {
	kv := e.Value.(*LRUEntry)
	c.list.Remove(e)
	delete(c.dict, kv.Key)
}

// 查看key是否存在，不移动链表
func (c *LRUCache) Exist(key string) bool {
	_, found := c.dict[key]
	return found
}

// 获取key对应的值，并把其移动到链表头部
func (c *LRUCache) Get(key string) (interface{}, bool) {
	e, found := c.dict[key]
	if found {
		c.list.MoveToFront(e)
		kv := e.Value.(*LRUEntry)
		if kv == nil {
			return nil, false
		}
		return kv.Value, true
	}
	return nil, false
}

// 获取key对应的值，不移动链表
func (c *LRUCache) Peek(key string) (interface{}, bool) {
	e, found := c.dict[key]
	if found {
		kv := e.Value.(*LRUEntry)
		return kv.Value, true
	}
	return nil, false
}

// 获取最老的值（链表尾节点）
func (c *LRUCache) GetOldest() (key string, value interface{}, ok bool) {
	ent := c.list.Back()
	if ent != nil {
		kv := ent.Value.(*LRUEntry)
		return kv.Key, kv.Value, true
	}
	return "", nil, false
}

// 把key-value加入到cache中，并移动到链表头部
func (c *LRUCache) Put(key string, value interface{}) bool {
	e, exist := c.dict[key]
	if exist {
		c.list.MoveToFront(e)
		e.Value.(*LRUEntry).Value = value
		return false
	} else {
		kv := &LRUEntry{Key: key, Value: value}
		e = c.list.PushFront(kv) // push entry to list front
		c.dict[key] = e
		if c.Len() > c.capacity {
			c.RemoveOldest()
		}
		return true
	}
}

// 把key从cache中删除
func (c *LRUCache) Remove(key string) bool {
	if e, ok := c.dict[key]; ok {
		c.removeElement(e)
		return true
	}
	return false
}

// 删除最老的的key-value，并返回
func (c *LRUCache) RemoveOldest() (string, interface{}, bool) {
	e := c.list.Back()
	if e != nil {
		kv := e.Value.(*LRUEntry)
		c.removeElement(e)
		return kv.Key, kv.Value, true
	}
	return "", nil, false
}

// 返回所有的key（从旧到新）
func (c *LRUCache) Keys() []string {
	keys := make([]string, len(c.dict))
	i := 0
	for e := c.list.Back(); e != nil; e = e.Prev() {
		keys[i] = e.Value.(*LRUEntry).Key
		i++
	}
	return keys
}
