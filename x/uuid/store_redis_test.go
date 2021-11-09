// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

//+build !ignore

package uuid

import (
	"context"
	"testing"
	"time"
)

var (
	redisAddr = "127.0.0.1:6379"
)

func createRedisStore(key string) Storage {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*60)
	var store = NewRedisStore(ctx, redisAddr, key)
	return store
}

func TestRedisStoreExample(t *testing.T) {
	var (
		count = 100000
		ids   []int64
		uuids = make(map[int64]bool)
	)
	var store = createRedisStore("/uuid/cnt1")
	defer store.Close()
	var start = time.Now()
	for i := 0; i < count; i++ {
		id, err := store.Incr()
		if err != nil {
			t.Fatalf("incr failed: %v", err)
		}
		if _, found := uuids[id]; found {
			t.Fatalf("duplicate id %d", id)
		}
		uuids[id] = true
		ids = append(ids, id)
	}
	var elapsed = time.Since(start).Seconds()
	t.Logf("QPS %.2f/s", float64(count)/elapsed)
	// Output:
	//  QPS 20607.68/s
}

func TestRedisStoreConcurrent(t *testing.T) {
	var gcnt = 10
	var eachMax = 20000
	var store = createRedisStore("/uuid/cnt2")
	defer store.Close()

	var idGen = NewPersistIDGen(store)
	var workerCtx = NewWorkerContext(eachMax, func() IDGenerator { return idGen })
	for i := 0; i < gcnt; i++ {
		workerCtx.Go(t, i)
	}
	workerCtx.Wait()

	var elapsed = workerCtx.Duration().Seconds()
	if !t.Failed() {
		t.Logf("QPS %.2f/s", float64(gcnt*eachMax)/elapsed)
	}
	// Output:
	//  QPS 18676.36/s
}

// N个并发worker，每个worker单独连接, 测试生成id的一致性
func TestRedisStoreDistributed(t *testing.T) {
	var gcnt = 10
	var eachMax = 100000
	var store = createRedisStore("/uuid/cnt3")
	defer store.Close()

	var workerCtx = NewWorkerContext(eachMax, func() IDGenerator { return NewPersistIDGen(store) })
	for i := 0; i < gcnt; i++ {
		workerCtx.Go(t, i)
	}
	workerCtx.Wait()

	var elapsed = workerCtx.Duration().Seconds()
	if !t.Failed() {
		t.Logf("QPS %.2f/s", float64(gcnt*eachMax)/elapsed)
	}
	// Output:
	//  QPS 19106.58/s
}
