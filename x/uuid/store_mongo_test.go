// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func createMongoStore(label string) Storage {
	db := "testdb"
	username := "admin"
	password := "cuKpVrfZzUvg"
	uri := fmt.Sprintf("mongodb://%s:%s@127.0.0.1:27017/?connect=direct", username, password)
	ctx, _ := context.WithTimeout(context.Background(), time.Second*60)
	return NewMongoDBStore(ctx, uri, db, label, DefaultSeqStep)
}

func TestMongoStoreExample(t *testing.T) {
	var store = createMongoStore("ctr101")
	var (
		count = 10000
		ids   []int64
		m     = make(map[int64]bool)
	)
	var start = time.Now()
	for i := 0; i < count; i++ {
		id, err := store.Incr()
		if err != nil {
			t.Fatalf("cannot incr %v", err)
		}
		if _, found := m[id]; found {
			t.Fatalf("duplicate id %d", id)
		}
		ids = append(ids, id)
	}
	var elapsed = time.Since(start).Seconds()
	t.Logf("QPS %.2f/s", float64(count)/elapsed)
	// Output:
	//    QPS 4910.91/s
}

func TestMongoStoreConcurrent(t *testing.T) {
	var gcnt = 10
	var eachMax = 100000
	var store = createMongoStore("ctr102")
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
	//  QPS 16647.61/s
}

// N个并发worker，每个worker单独连接, 测试生成id的一致性
func TestMongoStoreDistributed(t *testing.T) {
	var gcnt = 10
	var eachMax = 100000
	var store = createMongoStore("ctr103")
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
	//  QPS 16647.61/s
}
