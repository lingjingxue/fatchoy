// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"log"
	"sync"
	"testing"
	"time"

	"go.etcd.io/etcd/clientv3"
)

var (
	etcdAddr = "127.0.0.1:2379"
)

func createEtcdClient() *clientv3.Client {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdAddr},
		DialTimeout: time.Second * TimeoutSec,
	})
	if err != nil {
		log.Panicf("cannot connect etcd: %v", err)
	}
	return client
}

func TestEtcdStoreExample(t *testing.T) {
	cli := createEtcdClient()
	var store = NewEtcdStore(cli, "/uuid/ctr001")

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
	var elapsed = time.Now().Sub(start).Seconds()
	t.Logf("QPS %.2f/s", float64(count)/elapsed)
	// Output:
	//    QPS 2619.06/s
}

func createEtcdStore(key string, t *testing.T) Storage {
	cli := createEtcdClient()
	var store = NewEtcdStore(cli, key)
	return store
}

// N个并发worker，每个worker单独连接, 测试生成id的一致性
func TestEtcdStoreDistributed(t *testing.T) {
	var (
		wg      sync.WaitGroup
		guard   sync.Mutex
		gcnt    = 1
		eachMax = 1000
		m       = make(map[int64]int, 10000)
	)
	var start = time.Now()
	for i := 1; i <= gcnt; i++ {
		ctx := newWorkerContext(&wg, &guard, m, eachMax)
		ctx.idGenCreator = func() IDGenerator {
			store := createEtcdStore("uuid:ctr003", t)
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
	//  QPS 2552.59/s
}
