// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

//go:build !ignore

package collections

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"testing"
	"time"
)

var _ = os.Open

var testRandSeed = time.Now().UnixNano()

func init() {
	rand.Seed(testRandSeed)
}

type testPlayer struct {
	Uid   int64
	Point int64
	Level int16
}

func (p *testPlayer) CompareTo(other Comparable) int {
	var rhs = other.(*testPlayer)
	if p.Uid > rhs.Uid {
		return 1
	} else if p.Uid < rhs.Uid {
		return -1
	}
	return 0
}

func ExampleZSkipList() {
	var playerMap = make(map[int64]*testPlayer)
	var zset = NewSortedSet()

	//简单的测试角色数据
	var p1 = &testPlayer{Uid: 1001, Level: 12, Point: 2012}
	var p2 = &testPlayer{Uid: 1002, Level: 13, Point: 2015}
	var p3 = &testPlayer{Uid: 1003, Level: 14, Point: 2014}
	var p4 = &testPlayer{Uid: 1004, Level: 11, Point: 2014}
	var p5 = &testPlayer{Uid: 1005, Level: 14, Point: 2011}
	playerMap[p1.Uid] = p1
	playerMap[p2.Uid] = p2
	playerMap[p3.Uid] = p3
	playerMap[p4.Uid] = p4
	playerMap[p5.Uid] = p5

	//插入角色数据到zskiplist
	for _, v := range playerMap {
		zset.Add(v, v.Point)
	}

	//打印调试信息
	// fmt.Printf("%v\n", zset.zsl)

	//获取角色的排行信息
	var rank = zset.GetRank(p1, false) // in ascend order
	var myRank = zset.Len() - rank + 1 // get descend rank
	fmt.Printf("rank of %d: %d\n", p1.Uid, myRank)

	//根据排行获取角色信息
	//var node = zset.GetRank(rank)
	//var player = playerMap[node.Obj.Uuid()]
	//fmt.Printf("rank at %d is: %s\n", rank, player.name)
	//
	////遍历整个zskiplist
	//zsl.Walk(true, func(rank int, v RankInterface) bool {
	//	fmt.Printf("rank %d: %v", rank, v)
	//	return true
	//})
	//
	////从zset中删除p1
	//if !zset.Remove(p1) {
	//	// error handling
	//}
	//
	//p1.score += 10
	//if zset.Insert(p1.score, p1) == nil {
	//	// error handling
	//}
}

func makeTestPlayers(count, maxScore int, dupScore bool) map[int64]*testPlayer {
	var set = make(map[int64]*testPlayer, count)
	var nextID int64 = 100000000
	for i := 0; i < count; i++ {
		nextID++
		obj := &testPlayer{
			Uid:   nextID,
			Level: int16(rand.Int() % 60),
		}
		if dupScore {
			obj.Point = int64(rand.Int()%maxScore) + 1
		} else {
			obj.Point = int64(maxScore)
			maxScore--
		}
		set[obj.Uid] = obj
	}
	return set
}

func dumpToFile(zsl *ZSkipList, filename string) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		log.Fatalf("OpenFile: %v", err)
	}
	zsl.Dump(f)
}

func dumpSliceToFile(players []*testPlayer, filename string) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		log.Fatalf("OpenFile: %v", err)
	}

	fmt.Fprint(f, "    uid    rank    score    level\n")
	var count = 0
	for i := 0; i < len(players); i++ {
		var item = players[i]
		count++
		fmt.Fprintf(f, "%8d, %5d %5d, %4d\n", item.Uid, count, item.Point, item.Level)
	}
	f.Close()
}

func mapToSlice(set map[int64]*testPlayer) []*testPlayer {
	var slice = make([]*testPlayer, 0, len(set))
	for _, v := range set {
		slice = append(slice, v)
	}
	return slice
}

type tester interface {
	Fatalf(format string, args ...interface{})
}

// Update score(zskiplist insert and delete) in many times
func manyUpdate(t tester, zset *SortedSet, set map[int64]*testPlayer, count int) {
	for _, v := range set {
		var oldScore = v.Point
		if !zset.Remove(v) {
			t.Fatalf("manyUpdate: delete old item[%d-%d] fail", v.Uid, v.Point)
			break
		}
		v.Point += int64(rand.Uint32()%100) + 1
		if !zset.Add(v, v.Point) {
			t.Fatalf("manyUpdate: insert new item[%d-%d] fail, old score: %d", v.Uid, v.Point, oldScore)
			break
		}
		count--
		if count == 0 {
			break
		}
	}
}

func TestZSkipListInsertRemove(t *testing.T) {
	const units = 20000
	var set = makeTestPlayers(units, 1000, true)
	var zset = NewSortedSet()
	var maxTurn = 10
	for i := 0; i < maxTurn; i++ {
		// First insert all player to zskiplist
		for _, v := range set {
			if !zset.Add(v, v.Point) {
				t.Fatalf("insert item[%d-%d] failed", v.Point, v.Uid)
			}
		}
		if zset.Len() != units {
			t.Fatalf("unexpected skiplist element count, %d != %d", zset.Len(), units)
		}

		// Second remove all players in zskiplist
		for _, v := range set {
			if !zset.Remove(v) {
				t.Fatalf("delete item[%d-%d] failed", v.Point, v.Uid)
			}
		}

		if zset.Len() != 0 {
			t.Fatalf("skiplist not empty")
		}
	}
}

func TestZSkipListChangedInsert(t *testing.T) {
	const units = 20000
	var set = makeTestPlayers(units, 1000, true)
	var zset = NewSortedSet()

	// Insert all player to zskiplist
	for _, v := range set {
		if !zset.Add(v, v.Point) {
			t.Fatalf("insert item[%d-%d] failed", v.Point, v.Uid)
		}
	}

	// Update half elements
	manyUpdate(t, zset, set, units/2)

	if zset.Len() != units {
		t.Fatalf("unexpected skiplist element count")
	}

	// Delete all elements
	for _, v := range set {
		if !zset.Remove(v) {
			t.Fatalf("delete set item[%d-%d] failed", v.Point, v.Uid)
		}
	}
	if zset.Len() != 0 {
		t.Fatalf("skiplist expected empty, but got size: %d", zset.Len())
	}
}

func TestZSkipListGetRank(t *testing.T) {
	const units = 20000
	var set = makeTestPlayers(units, units, false)
	var zset = NewSortedSet()
	for _, v := range set {
		if !zset.Add(v, v.Point) {
			t.Fatalf("insert item[%d-%d] failed", v.Point, v.Uid)
		}
	}

	// rank by sort package
	var ranks = mapToSlice(set)
	sort.SliceStable(ranks, func(i, j int) bool {
		return ranks[i].Point < ranks[j].Point
	})

	for i := len(ranks); i > 0; i-- {
		var v = ranks[i-1]
		var thisRank = len(ranks) - i + 1
		var rank = zset.Len() - zset.GetRank(v, false)
		if rank != thisRank {
			t.Fatalf("%v not equal at rank, %d != %d", v, rank, thisRank)
			break
		}
	}
}

func TestZSkipListUpdateGetRank(t *testing.T) {
	const units = 20000
	var set = makeTestPlayers(units, units, false)
	var zset = NewSortedSet()
	for _, v := range set {
		if !zset.Add(v, v.Point) {
			t.Fatalf("insert item[%d-%d] failed", v.Point, v.Uid)
		}
	}

	var maxTurn = 10
	for i := 0; i < maxTurn; i++ {
		manyUpdate(t, zset, set, units/2)

		// rank by sort package
		var ranks = mapToSlice(set)
		sort.SliceStable(ranks, func(i, j int) bool {
			return ranks[i].Point < ranks[j].Point
		})

		for i := len(ranks); i > 0; i-- {
			var v = ranks[i-1]
			var rank = zset.GetRank(v, false)
			var myRank = zset.Len() - rank + 1
			var thisRank = len(ranks) - i + 1
			if myRank != thisRank {
				var nodes = zset.GetRange(rank, rank, false)
				if len(nodes) == 0 {
					t.Fatalf("%v GetElementByRank return nil: %d", v, rank)
					break
				}
			}
		}
	}
}

func BenchmarkZSkipListInsert(b *testing.B) {
	b.StopTimer()
	var zset = NewSortedSet()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		obj := &testPlayer{
			Uid:   int64(i),
			Level: int16(i),
			Point: int64(i),
		}
		if !zset.Add(obj, obj.Point) {
			b.Fatalf("insert item[%d-%d] failed", obj.Point, obj.Uid)
		}
	}
}

func BenchmarkZSkipListUpdate(b *testing.B) {
	b.StopTimer()
	const units = 20000
	var set = makeTestPlayers(units, units, true)
	var zset = NewSortedSet()
	for _, v := range set {
		if !zset.Add(v, v.Point) {
			b.Fatalf("insert item[%d-%d] failed", v.Point, v.Uid)
		}
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		manyUpdate(b, zset, set, units/2)
	}
}

func BenchmarkZSkipListGetRank(b *testing.B) {
	b.StopTimer()
	const units = 20000
	var set = makeTestPlayers(units, units, false)
	var zset = NewSortedSet()
	for _, v := range set {
		if !zset.Add(v, v.Point) {
			b.Fatalf("insert item[%d-%d] failed", v.Point, v.Uid)
		}
	}
	b.StartTimer()
	for i := 1; i < b.N; i++ {
		var obj *testPlayer
		for _, v := range set {
			obj = v
			break
		}
		zset.GetRank(obj, false)
	}
}

func BenchmarkZSkipListGetElementByRank(b *testing.B) {
	b.StopTimer()
	const units = 20000
	var set = makeTestPlayers(units, units, false)
	var zset = NewSortedSet()
	for _, v := range set {
		if !zset.Add(v, v.Point) {
			b.Fatalf("insert item[%d-%d] failed", v.Point, v.Uid)
		}
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		var rank = (i % units) + 1
		zset.GetRange(rank, rank, false)
	}
}
