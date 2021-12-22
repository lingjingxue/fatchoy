// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

// This skiplist implementation is almost a translation of the original
// algorithm described by William Pugh in "Skip Lists: A Probabilistic
// Alternative to Balanced Trees", modified in three ways:
// a) this implementation allows for repeated scores.
// b) the comparison is not just by key (our 'score') but by satellite data.
// c) there is a back pointer, so it's a doubly linked list with the back
// pointers being only at "level 1". This allows to traverse the list
// from tail to head.
//
// https://github.com/antirez/redis/blob/3.2/src/t_zset.c
// https://en.wikipedia.org/wiki/Skip_list

package zset

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
)

const (
	ZSKIPLIST_MAXLEVEL = 12   // Should be enough
	ZSKIPLIST_P        = 0.25 // Skiplist P = 1/4
)

// each level of list node
type zskipListLevel struct {
	forward *ZSkipListNode // link to next node
	span    int            // node # between this and forward link
}

// list node
type ZSkipListNode struct {
	Ele      KeyType
	Score    int64
	backward *ZSkipListNode
	level    []zskipListLevel
}

func newZSkipListNode(level int, score int64, element KeyType) *ZSkipListNode {
	return &ZSkipListNode{
		Ele:   element,
		Score: score,
		level: make([]zskipListLevel, level),
	}
}

func (n *ZSkipListNode) Before() *ZSkipListNode {
	return n.backward
}

// Next return next forward pointer
func (n *ZSkipListNode) Next() *ZSkipListNode {
	return n.level[0].forward
}

// 带索引的排序链表
type ZSkipList struct {
	head   *ZSkipListNode // 头结点
	tail   *ZSkipListNode // 尾节点（最大值节点）
	length int            // 节点数
	level  int            // 层级
}

func NewZSkipList() *ZSkipList {
	return &ZSkipList{
		level: 1,
		head:  newZSkipListNode(ZSKIPLIST_MAXLEVEL, 0, nil),
	}
}

// 返回新节点的随机层级[1-ZSKIPLIST_MAXLEVEL]
func (zsl *ZSkipList) randLevel() int {
	var level = 1
	for {
		var seed = rand.Uint32() & 0xFFFF
		if float32(seed) < ZSKIPLIST_P*0xFFFF {
			level++
		} else {
			break
		}
	}
	if level > ZSKIPLIST_MAXLEVEL {
		level = ZSKIPLIST_MAXLEVEL
	}
	return level
}

// 链表的节点数量
func (zsl *ZSkipList) Len() int {
	return zsl.length
}

// 链表的层级
func (zsl *ZSkipList) Height() int {
	return zsl.level
}

// 头结点
func (zsl *ZSkipList) HeadNode() *ZSkipListNode {
	return zsl.head.level[0].forward
}

// 尾节点
func (zsl *ZSkipList) TailNode() *ZSkipListNode {
	return zsl.tail
}

// 插入一个不存在的节点
func (zsl *ZSkipList) Insert(score int64, ele KeyType) *ZSkipListNode {
	var update [ZSKIPLIST_MAXLEVEL]*ZSkipListNode
	var rank [ZSKIPLIST_MAXLEVEL]int

	var x = zsl.head
	for i := zsl.level - 1; i >= 0; i-- {
		// store rank that is crossed to reach the insert position
		if i != zsl.level-1 {
			rank[i] = rank[i+1]
		}
		for x.level[i].forward != nil &&
			(x.level[i].forward.Score < score ||
				(x.level[i].forward.Score == score &&
					x.level[i].forward.Ele.CompareTo(ele) < 0)) {
			rank[i] += x.level[i].span
			x = x.level[i].forward
		}
		update[i] = x
	}
	// we assume the element is not already inside, since we allow duplicated
	// scores, reinserting the same element should never happen since the
	// caller of zslInsert() should test in the hash table if the element is
	// already inside or not.
	var level = zsl.randLevel()
	if level > zsl.level {
		for i := zsl.level; i < level; i++ {
			rank[i] = 0
			update[i] = zsl.head
			update[i].level[i].span = zsl.length
		}
		zsl.level = level
	}
	x = newZSkipListNode(level, score, ele)
	for i := 0; i < level; i++ {
		x.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = x

		// update span covered by update[i] as x is inserted here
		x.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}
	// increment span for untouched levels
	for i := level; i < zsl.level; i++ {
		update[i].level[i].span++
	}
	if update[0] != zsl.head {
		x.backward = update[0]
	} else {
		x.backward = nil
	}
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x
	} else {
		zsl.tail = x
	}
	zsl.length++
	return x
}

// 删除一个节点
func (zsl *ZSkipList) deleteNode(x *ZSkipListNode, update []*ZSkipListNode) {
	for i := 0; i < zsl.level; i++ {
		if update[i].level[i].forward == x {
			update[i].level[i].span += x.level[i].span - 1
			update[i].level[i].forward = x.level[i].forward
		} else {
			update[i].level[i].span -= 1
		}
	}
	if x.level[0].forward != nil {
		x.level[0].forward.backward = x.backward
	} else {
		zsl.tail = x.backward
	}
	for zsl.level > 1 && zsl.head.level[zsl.level-1].forward == nil {
		zsl.level--
	}
	zsl.length--
}

// 删除对应score的节点
func (zsl *ZSkipList) Delete(score int64, ele KeyType) *ZSkipListNode {
	var update [ZSKIPLIST_MAXLEVEL]*ZSkipListNode
	var x = zsl.head
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(x.level[i].forward.Score < score ||
				(x.level[i].forward.Score == score &&
					x.level[i].forward.Ele.CompareTo(ele) < 0)) {
			x = x.level[i].forward
		}
		update[i] = x
	}

	// We may have multiple elements with the same score, what we need
	// is to find the element with both the right score and object.
	x = x.level[0].forward
	if x != nil {
		if score == x.Score && x.Ele.CompareTo(ele) == 0 {
			zsl.deleteNode(x, update[0:])
			return x
		}
		// log.Printf("zskiplist need delete %v, but found %v\n", ele, x.Ele)
	}
	return nil // not found
}

// 删除排名在[start-end]之间的节点，排名从1开始
func (zsl *ZSkipList) DeleteRangeByRank(start, end int, dict map[KeyType]int64) int {
	var update [ZSKIPLIST_MAXLEVEL]*ZSkipListNode
	var traversed, removed int
	var x = zsl.head
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && (traversed+x.level[i].span < start) {
			traversed += x.level[i].span
			x = x.level[i].forward
		}
		update[i] = x
	}
	traversed++
	x = x.level[0].forward
	for x != nil && traversed <= end {
		var next = x.level[0].forward
		zsl.deleteNode(x, update[0:])
		delete(dict, x.Ele)
		removed++
		traversed++
		x = next
	}
	return removed
}

// 删除score在[min-max]之间的节点
func (zsl *ZSkipList) DeleteRangeByScore(min, max int64, dict map[KeyType]int64) int {
	var update [ZSKIPLIST_MAXLEVEL]*ZSkipListNode
	var removed int
	var x = zsl.head
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && x.level[i].forward.Score <= min {
			x = x.level[i].forward
		}
		update[i] = x
	}

	// Current node is the last with score < or <= min
	x = x.level[0].forward

	// Delete nodes while in range
	for x != nil && x.Score <= max {
		var next = x.level[0].forward
		zsl.deleteNode(x, update[0:])
		delete(dict, x.Ele)
		removed++
		x = next
	}
	return removed
}

// 获取score所在的排名，排名从1开始
func (zsl *ZSkipList) GetRank(score int64, ele KeyType) int {
	var rank = 0
	var x = zsl.head
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(x.level[i].forward.Score < score ||
				(x.level[i].forward.Score == score &&
					x.level[i].forward.Ele.CompareTo(ele) <= 0)) {
			rank += x.level[i].span
			x = x.level[i].forward
		}

		// x might be equal to zsl->header, so test if obj is non-nil
		if x.Ele != nil && x.Ele.CompareTo(ele) == 0 {
			return rank
		}
	}
	return 0
}

// 根据排名获得节点，排名从1开始
func (zsl *ZSkipList) GetElementByRank(rank int) *ZSkipListNode {
	var tranversed = 0
	var x = zsl.head
	for i := zsl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && (tranversed+x.level[i].span <= rank) {
			tranversed += x.level[i].span
			x = x.level[i].forward
		}
		if tranversed == rank {
			return x
		}
	}
	return nil
}

// Returns if there is a part of the zset is in range.
func (zsl *ZSkipList) IsInRange(min, max int64) bool {
	if min > max {
		return false
	}
	var x = zsl.tail
	if x == nil || x.Score < min {
		return false
	}
	x = zsl.head.level[0].forward
	if x == nil || x.Score > max {
		return false
	}
	return true
}

// Find the first node that is contained in the specified range.
// Returns NULL when no element is contained in the range.
func (zsl *ZSkipList) FirstInRange(min, max int64) *ZSkipListNode {
	if !zsl.IsInRange(min, max) {
		return nil
	}
	var x = zsl.head
	for i := zsl.level - 1; i >= 0; i-- {
		// Go forward while *OUT* of range.
		for x.level[i].forward != nil && x.level[i].forward.Score < min {
			x = x.level[i].forward
		}
	}
	// This is an inner range, so the next node cannot be NULL.
	x = x.level[0].forward
	if x != nil && x.Score > max {
		return nil
	}
	return x
}

// Find the last node that is contained in the specified range.
// Returns NULL when no element is contained in the range.
func (zsl *ZSkipList) LastInRange(min, max int64) *ZSkipListNode {
	if !zsl.IsInRange(min, max) {
		return nil
	}
	var x = zsl.head
	for i := zsl.level - 1; i >= 0; i-- {
		// Go forward while *OUT* of range.
		for x.level[i].forward != nil && x.level[i].forward.Score <= max {
			x = x.level[i].forward
		}
	}
	// Check if score >= min.
	if x.Score < min {
		return nil
	}
	return x
}

func (zsl ZSkipList) String() string {
	var buf bytes.Buffer
	zsl.Dump(&buf)
	return buf.String()
}

// dump whole list to w, mostly for debugging
func (zsl *ZSkipList) Dump(w io.Writer) {
	var x = zsl.head
	// dump header
	var line bytes.Buffer
	n, _ := fmt.Fprintf(w, "<             head> ")
	prePadding(&line, n)
	for i := 0; i < zsl.level; i++ {
		if i < len(x.level) {
			if x.level[i].forward != nil {
				fmt.Fprintf(w, "[%2d] ", x.level[i].span)
				line.WriteString("  |  ")
			}
		}
	}
	fmt.Fprint(w, "\n")
	line.WriteByte('\n')
	line.WriteTo(w)

	// dump list
	var count = 0
	x = x.level[0].forward
	for x != nil {
		count++
		zsl.dumpNode(w, x, count)
		if len(x.level) > 0 {
			x = x.level[0].forward
		}
	}

	// dump tail end
	fmt.Fprintf(w, "<             end> ")
	for i := 0; i < zsl.level; i++ {
		fmt.Fprintf(w, "  _  ")
	}
	fmt.Fprintf(w, "\n")
}

func (zsl *ZSkipList) dumpNode(w io.Writer, node *ZSkipListNode, count int) {
	var line bytes.Buffer
	var ss = fmt.Sprintf("%v", node.Ele)
	n, _ := fmt.Fprintf(w, "<%6d %4d, %s> ", node.Score, count, ss)
	prePadding(&line, n)
	for i := 0; i < zsl.level; i++ {
		if i < len(node.level) {
			fmt.Fprintf(w, "[%2d] ", node.level[i].span)
			line.WriteString("  |  ")
		} else {
			if shouldLinkVertical(zsl.head, node, i) {
				fmt.Fprintf(w, "  |  ")
				line.WriteString("  |  ")
			}
		}
	}
	fmt.Fprint(w, "\n")
	line.WriteByte('\n')
	line.WriteTo(w)
}

func shouldLinkVertical(head, node *ZSkipListNode, level int) bool {
	if node.backward == nil { // first element
		return head.level[level].span >= 1
	}
	var tranversed = 0
	var prev *ZSkipListNode
	var x = node.backward
	for x != nil {
		if level >= len(x.level) {
			return true
		}
		if x.level[level].span > tranversed {
			return true
		}
		tranversed++
		prev = x
		x = x.backward
	}
	if prev != nil && level < len(prev.level) {
		return prev.level[level].span >= tranversed
	}
	return false
}

func prePadding(line *bytes.Buffer, n int) {
	for i := 0; i < n; i++ {
		line.WriteByte(' ')
	}
}
