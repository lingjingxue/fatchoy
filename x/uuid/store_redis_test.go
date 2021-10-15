// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

//+build !ignore

package uuid

import (
	"context"
	"sync"
	"testing"
	"time"
)

var (
	redisAddr = "192.168.132.129:6379"
)

func createRedisStore(key string, t *testing.T) Storage {
	var store = NewRedisStore(context.Background(), redisAddr, key)
	return store
}

func TestRedisStoreExample(t *testing.T) {
	var store = NewRedisStore(context.Background(), redisAddr, "/uuid/cnt1")
	defer store.Close()
	var (
		count = 100000
		ids   []int64
		rkeys = make(map[int64]bool)
	)
	var start = time.Now()
	for i := 0; i < count; i++ {
		id, err := store.Incr()
		if err != nil {
			t.Fatalf("incr failure: %v", err)
		}
		if _, found := rkeys[id]; found {
			t.Fatalf("duplicate id %d", id)
		}
		rkeys[id] = true
		ids = append(ids, id)
	}
	var elapsed = time.Now().Sub(start).Seconds()
	t.Logf("QPS %.2f/s", float64(count)/elapsed)
	// Output:
	//  QPS 30526.92/s
}

// N个并发worker，每个worker单独连接, 测试生成id的一致性
func TestRedisStoreDistributed(t *testing.T) {
	var (
		wg      sync.WaitGroup
		guard   sync.Mutex
		gcnt    = 20
		eachMax = 100000
		m       = make(map[int64]int, 10000)
	)
	var start = time.Now()
	for i := 0; i <= gcnt; i++ {
		ctx := newWorkerContext(&wg, &guard, m, eachMax)
		ctx.idGenCreator = func() IDGenerator {
			store := createRedisStore("uuid:cnt3", t)
			return NewStorageGen(store)
		}
		wg.Add(1)
		go runIDWorker(i, ctx, t)
	}
	wg.Wait()
	var elapsed = time.Now().Sub(start).Seconds()
	if !t.Failed() {
		t.Logf("QPS %.2f/s", float64(gcnt*eachMax)/elapsed)
	}
	// Output:
	//  QPS 75249.02/s
}
