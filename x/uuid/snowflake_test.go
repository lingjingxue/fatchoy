// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"sync"
	"testing"
	"time"
)

func TestSnowflakeLimit(t *testing.T) {
	var sf = NewSnowflake(1234)
	t.Logf("a typical uuid is %d", sf.MustNext())

	var epoch = time.Unix(CustomEpoch/int64(time.Second), 0)
	var endOfWorld = epoch.Add(time.Duration(TimeUnit) * ((1 << TimeUnitBits) - 1))
	t.Logf("the end time of uuid is %v", endOfWorld.UTC())
}

func TestSnowflakeNext(t *testing.T) {
	const count = 1000000
	var dict = make(map[int64]bool)
	var sf = NewSnowflake(1234)
	var start = time.Now()
	var l NoLock
	for i := 0; i < count; i++ {
		id := sf.MustNext()
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
	//   QPS: 66043.09/s
}

var (
	uuidMap   = make(map[int64]bool, 1000000)
	uuidGuard sync.Mutex
)

func newSnowflakeIDWorker(t *testing.T, sf *Snowflake, wg *sync.WaitGroup, gid int, count int) {
	defer wg.Done()
	//t.Logf("snowflake worker %d started", gid)
	for i := 0; i < count; i++ {
		id := sf.MustNext()
		if !putIfAbsent(&uuidGuard, uuidMap, id) {
			t.Fatalf("duplicate id %d after %d", id, i)
		}
	}
	//t.Logf("snowflake worker %d done", gid)
}

// 开启N个goroutine，测试UUID的并发性
func TestSnowflakeConcurrent(t *testing.T) {
	var gcount = 20
	var eachCnt = 100000
	var sf = NewSnowflake(1234)
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
	var sf = NewSnowflake(1234)
	for i := 0; i < 10000; i++ {
		sf.MustNext()
	}
}
