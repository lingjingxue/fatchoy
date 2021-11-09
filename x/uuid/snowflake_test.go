// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"sync"
	"testing"
	"time"
)

func TestSnowflakeNext(t *testing.T) {
	const count = 1000000
	var dict = make(map[int64]bool)
	var sf = NewSnowFlake(1234)
	var start = time.Now()
	var l NoLock
	for i := 0; i < count; i++ {
		id := sf.Next()
		if !putIfAbsent(&l, dict, id) {
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

func newSnowflakeIDWorker(t *testing.T, sf *SnowFlake, wg *sync.WaitGroup, gid int, count int) {
	defer wg.Done()
	//t.Logf("snowflake worker %d started", gid)
	for i := 0; i < count; i++ {
		id := sf.Next()
		if !putIfAbsent(&uuidGuard, uuidMap, id) {
			t.Fatalf("duplicate id %d after %d", id, i)
		}
	}
	//t.Logf("snowflake worker %d done", gid)
}

// 开启N个goroutine，测试NewID的并发性
func TestSnowflakeConcurrent(t *testing.T) {
	var gcount = 20
	var eachCnt = 100000
	var sf = NewSnowFlake(1234)
	var wg sync.WaitGroup
	wg.Add(gcount)
	for i := 0; i < gcount; i++ {
		go newSnowflakeIDWorker(t, sf, &wg, i, eachCnt)
	}
	wg.Wait()
	if len(uuidMap) != gcount*eachCnt {
		t.Fatalf("duplicate id found")
	}
}

func BenchmarkSnowflakeGen(b *testing.B) {
	var sf = NewSnowFlake(1234)
	for i := 0; i < 10000; i++ {
		sf.Next()
	}
}
