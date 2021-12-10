// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package lru

import (
	"container/list"
)

type Entry struct {
	Key   interface{}
	Value interface{}
}

// LRU缓存
// https://en.wikipedia.org/wiki/Cache_replacement_policies#Least_recently_used_(LRU)
type Cache struct {
	list      *list.List
	items     map[interface{}]*list.Element
	onEvicted func(k, v interface{})
	size      int
}

func NewCache(size int, onEvicted func(k, v interface{})) *Cache {
	if size <= 0 {
		panic("cache capacity out of range")
	}
	cache := &Cache{
		size:      size,
		onEvicted: onEvicted,
		list:      list.New(),
		items:     make(map[interface{}]*list.Element, size),
	}
	return cache
}

func (c *Cache) Len() int {
	return c.list.Len()
}

func (c *Cache) Cap() int {
	return c.size
}

// 查看key是否存在，不移动链表
func (c *Cache) Contains(key interface{}) bool {
	_, found := c.items[key]
	return found
}

// 获取key对应的值，并把其移动到链表头部
func (c *Cache) Get(key interface{}) (interface{}, bool) {
	e, found := c.items[key]
	if found {
		c.list.MoveToFront(e)
		kv := e.Value.(*Entry)
		if kv == nil {
			return nil, false
		}
		return kv.Value, true
	}
	return nil, false
}

// 获取key对应的值，不移动链表
func (c *Cache) Peek(key interface{}) (interface{}, bool) {
	e, found := c.items[key]
	if found {
		kv := e.Value.(*Entry)
		return kv.Value, true
	}
	return nil, false
}

// 获取最老的值（链表尾节点）
func (c *Cache) GetOldest() (key, value interface{}, ok bool) {
	ent := c.list.Back()
	if ent != nil {
		kv := ent.Value.(*Entry)
		return kv.Key, kv.Value, true
	}
	return "", nil, false
}

// 返回所有的key（从旧到新）
func (c *Cache) Keys() []interface{} {
	keys := make([]interface{}, len(c.items))
	i := 0
	for e := c.list.Back(); e != nil; e = e.Prev() {
		keys[i] = e.Value.(*Entry).Key
		i++
	}
	return keys
}

// 把key-value加入到cache中，并移动到链表头部
func (c *Cache) Put(key interface{}, value interface{}) bool {
	e, exist := c.items[key]
	if exist {
		c.list.MoveToFront(e)
		e.Value.(*Entry).Value = value
		return false
	}
	entry := &Entry{Key: key, Value: value}
	e = c.list.PushFront(entry) // push entry to list front
	c.items[key] = e
	if c.Len() > c.size {
		c.removeOldest()
	}
	return true
}

// Resize changes the cache size.
func (c *Cache) Resize(size int) int {
	diff := c.Len() - size
	if diff < 0 {
		diff = 0
	}
	for i := 0; i < diff; i++ {
		c.removeOldest()
	}
	c.size = size
	return diff
}

// 把key从cache中删除
func (c *Cache) Remove(key string) bool {
	if e, ok := c.items[key]; ok {
		c.removeElement(e)
		return true
	}
	return false
}

// 删除最老的的key-value，并返回
func (c *Cache) RemoveOldest() (key, value interface{}, ok bool) {
	e := c.list.Back()
	if e != nil {
		entry := e.Value.(*Entry)
		c.removeElement(e)
		return entry.Key, entry.Value, true
	}
	return
}

// 清除所有
func (c *Cache) Purge() {
	for k, v := range c.items {
		if c.onEvicted != nil {
			c.onEvicted(k, v)
		}
		delete(c.items, k)
	}
	c.list.Init()
}

// removeOldest removes the oldest item from the cache.
func (c *Cache) removeOldest() {
	ent := c.list.Back()
	if ent != nil {
		c.removeElement(ent)
	}
}

// remove a given list element from the cache
func (c *Cache) removeElement(e *list.Element) {
	entry := e.Value.(*Entry)
	c.list.Remove(e)
	delete(c.items, entry.Key)
	if c.onEvicted != nil {
		c.onEvicted(entry.Key, entry.Value)
	}
}
