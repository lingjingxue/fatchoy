// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func createTestDBConn() *mongo.Client {
	username := "admin"
	password := "cuKpVrfZzUvg"
	uri := fmt.Sprintf("mongodb://%s:%s@127.0.0.1:27017/?connect=direct", username, password)

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		log.Panicf("%v", err)
	}
	if err = client.Ping(context.TODO(), nil); err != nil {
		log.Panicf("%v", err)
	}
	return client
}

func createMongoStore(label string) Storage {
	db := "testdb"
	cli := createTestDBConn()
	return NewMongoDBStore(cli, context.TODO(), db, label, DefaultSeqStep)
}

func TestMongoStoreExample(t *testing.T) {
	var store = createMongoStore("ctr001")
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
	//    QPS 5282.06/s
}

// N个并发worker，每个worker单独连接, 测试生成id的一致性
func TestMongoStoreDistributed(t *testing.T) {
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
			store := createMongoStore("ctr003")
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
	//  QPS 4998.85/s
}
