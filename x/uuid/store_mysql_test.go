// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"context"
	"testing"
	"time"
)

var mysqlDSN = "root:LGrk4IaS0Wflxw@tcp(localhost:3306)/testdb"

func createMySQLStore(label string) Storage {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*60)
	return NewMySQLStore(ctx, mysqlDSN, "uuid", label, 2000)
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

func TestMySQLStoreConcurrent(t *testing.T) {
	var gcnt = 10
	var eachMax = 100000
	var store = createMySQLStore("uuid_test2")
	defer store.Close()
	var generator = func() IDGenerator { return NewPersistIDGenAdapter(store) }
	var workerCtx = NewWorkerContext(eachMax, generator)
	for i := 0; i < gcnt; i++ {
		workerCtx.Go(t, i)
	}
	workerCtx.Wait()

	var elapsed = workerCtx.Duration().Seconds()
	if !t.Failed() {
		t.Logf("QPS %.2f/s", float64(gcnt*eachMax)/elapsed)
	}
	// Output:
	//  QPS 1070.08/s
}

// N个并发worker，每个worker单独连接, 测试生成id的一致性
func TestMySQLStoreDistributed(t *testing.T) {
	var gcnt = 10
	var eachMax = 100000

	var generator = func() IDGenerator {
		var store = createMySQLStore("uuid_test3")
		return NewPersistIDGenAdapter(store)
	}
	var workerCtx = NewWorkerContext(eachMax, generator)
	for i := 0; i < gcnt; i++ {
		workerCtx.Go(t, i)
	}
	workerCtx.Wait()

	var elapsed = workerCtx.Duration().Seconds()
	if !t.Failed() {
		t.Logf("QPS %.2f/s", float64(gcnt*eachMax)/elapsed)
	}
	// Output:
	//  QPS 1070.08/s
}
