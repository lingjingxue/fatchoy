// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"context"
	"log"
	"testing"
	"time"

	"go.etcd.io/etcd/clientv3"
)

var etcdAddr = "localhost:2379"

func createEtcdClient() *clientv3.Client {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdAddr},
		DialTimeout: time.Second * OpTimeout,
	})
	if err != nil {
		log.Panicf("cannot connect etcd: %v", err)
	}
	return client
}

func createEtcdStore(key string) Storage {
	cli := createEtcdClient()
	ctx, _ := context.WithTimeout(context.Background(), time.Second*30)
	var store = NewEtcdStore(ctx, cli, key)
	return store
}

func TestEtcdStoreExample(t *testing.T) {
	var store = createEtcdStore("/uuid/ctr101")

	var (
		count = 100000
		ids   []int64
		m     = make(map[int64]bool)
	)
	var start = time.Now()
	for i := 0; i < count; i++ {
		id, err := store.Incr()
		if err != nil {
			t.Fatalf("cannot incr at %d: %v", i, err)
		}
		if _, found := m[id]; found {
			t.Fatalf("duplicate id %d", id)
		}
		ids = append(ids, id)
	}
	var elapsed = time.Since(start).Seconds()
	t.Logf("QPS %.2f/s", float64(count)/elapsed)
	// Output:
	//    QPS 2475.96/s
}

func TestEtcdStoreConcurrent(t *testing.T) {
	var gcnt = 10
	var eachMax = 100000
	var store = createEtcdStore("/uuid/ctr102")
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
}

// N个并发worker，每个worker单独连接, 测试生成id的一致性
func TestEtcdStoreDistributed(t *testing.T) {
	var gcnt = 10
	var eachMax = 100000
	var generator = func() IDGenerator {
		var store = createEtcdStore("/uuid/ctr103")
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
	//  QPS 18234.02/s
}
