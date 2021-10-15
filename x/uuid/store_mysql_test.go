// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"context"
	"sync"
	"testing"
	"time"
)

var mysqlDSN = "root:HuppusoxOzs@tcp(127.0.0.1:3306)/testdb"

func createMySQLStore(label string) Storage {
	return NewMySQLStore(context.Background(), mysqlDSN, "uuid", label, 2000)
}

func TestMySQLStoreExample(t *testing.T) {
	var store = createMySQLStore("uuid_test1")
	defer store.Close()
	var (
		count = 10000
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
	var elapsed = time.Since(start).Seconds()
	t.Logf("QPS %.2f/s", float64(count)/elapsed)
	// Output:
	//  QPS 918.53/s
}

// N个并发worker，每个worker单独连接, 测试生成id的一致性
func TestMySQLStoreDistributed(t *testing.T) {
	var (
		wg      sync.WaitGroup
		guard   sync.Mutex
		gcnt    = 10
		eachMax = 10000
		m       = make(map[int64]int, 10000)
	)
	var start = time.Now()
	for i := 1; i <= gcnt; i++ {
		ctx := newWorkerContext(&wg, &guard, m, eachMax)
		ctx.idGenCreator = func() IDGenerator {
			store := createMySQLStore("distribute_uid2")
			return NewStorageGen(store)
		}
		wg.Add(1)
		go runIDWorker(i, ctx, t)
	}
	wg.Wait()
	var elapsed = time.Since(start).Seconds()
	if !t.Failed() {
		t.Logf("QPS %.2f/s", float64(gcnt*eachMax)/elapsed)
	}
	// Output:
	//  QPS 1070.08/s
}
