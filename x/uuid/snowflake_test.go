// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"sync"
	"testing"
	"time"
)

func putIfAbsent(dict map[int64]bool, id int64) bool {
	if _, found := dict[id]; !found {
		dict[id] = true
		return true
	}
	return false
}

func TestSnowflakeNext(t *testing.T) {
	const count = 2000000
	var dict = make(map[int64]bool)
	var sf = NewSnowFlake(1234)
	var start = time.Now()
	for i := 0; i < count; i++ {
		id := sf.Next()
		if !putIfAbsent(dict, id) {
			t.Fatalf("duplicate id %d after %d", id, i)
			return
		}
	}
	if len(dict) != count {
		t.Fatalf("duplicate id found")
	}
	var expired = time.Now().Sub(start)
	t.Logf("QPS: %.02f/s", count/expired.Seconds())
	// Output:
	//   QPS: 101889.51/s
}

func TestSnowflakeLimit(t *testing.T) {
	var sf = NewSnowFlake(1234)
	var uuid = sf.Next()
	println("an uuid is", uuid)

	tsUp := Twepoch/1000_000_000 + (int64(1<<TimeUnitBits)-1)/100
	var ts = time.Unix(tsUp, 0)
	t.Logf("snowfake will exhausted at %v", ts.UTC())
}

var (
	uuidMap   = make(map[int64]bool, 1000000)
	uuidGuard sync.Mutex
)

func newSnowflakeIDWorker(gid int, done chan int, count int, sf *SnowFlake, t *testing.T) {
	defer func() {
		done <- gid
	}()
	for i := 0; i < count; i++ {
		id := sf.Next()
		uuidGuard.Lock()
		if !putIfAbsent(uuidMap, id) {
			uuidGuard.Unlock()
			t.Logf("duplicate id %d after %d", id, i)
		} else {
			uuidGuard.Unlock()
		}
	}
	//t.Logf("worker %d completed\n", gid)
}

// 开启N个goroutine，测试NewID的并发性
func TestSnowflakeConcurrent(t *testing.T) {
	var gcount = 100
	var eachCnt = 150000
	var sf = NewSnowFlake(1234)
	var done = make(chan int, gcount)
	for i := 0; i < gcount; i++ {
		go newSnowflakeIDWorker(i, done, eachCnt, sf, t)
	}
	for i := 0; i < gcount; i++ {
		<-done
	}
	if len(uuidMap) != gcount*eachCnt {
		t.Fatalf("duplicate id found")
	}
}

func BenchmarkSnowflakeGen(b *testing.B) {
	var sf = NewSnowFlake(1234)
	for i := 0; i < b.N; i++ {
		sf.Next()
	}
}
