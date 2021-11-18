// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"go.etcd.io/etcd/clientv3"
)

var (
	etcdAddr  = "localhost:2379"
	redisAddr = "localhost:6379"
	mongoUri  string
	mysqlDSN  string
)

func init() {
	var username = os.Getenv("MONGODB_USER")
	var passwd = os.Getenv("MONGODB_PASSWORD")
	mongoUri = fmt.Sprintf("mongodb://%s:%s@127.0.0.1:27017/?connect=direct", username, passwd)
	println("mongo user:", username)
	println("mongo password:", passwd)

	username = os.Getenv("MYSQL_USER")
	passwd = os.Getenv("MYSQL_PASSWORD")
	var db = os.Getenv("MYSQL_DATABASE")
	println("mysql user:", username)
	println("mysql password:", passwd)
	mysqlDSN = fmt.Sprintf("%s:%s@tcp(localhost:3306)/%s", username, passwd, db)
}

func createCounterStorage(storeTye string, label string) Storage {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*30)
	switch storeTye {
	case "etcd":
		client, err := clientv3.New(clientv3.Config{
			Endpoints:   []string{etcdAddr},
			DialTimeout: time.Second * OpTimeout,
		})
		if err != nil {
			log.Panicf("cannot connect etcd: %v", err)
		}
		return NewEtcdStore(ctx, client, label)

	case "mongo":
		var db = "testdb"
		return NewMongoDBStore(ctx, mongoUri, db, label, DefaultSeqStep)
	case "mysql":
		return NewMySQLStore(ctx, mysqlDSN, "uuid", label, 2000)
	case "redis":
		return NewRedisStore(ctx, redisAddr, label)
	default:
		panic(fmt.Sprintf("invalid storage type %s", storeTye))
	}
	return nil
}

func runSeqIDTestSimple(t *testing.T, storeTye, label string) {
	var store = createCounterStorage(storeTye, label)
	var seq = NewSeqID(store, DefaultSeqStep)
	if err := seq.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	var m = make(map[int64]bool)
	var start = time.Now()
	const tetLoad = 2000000
	for i := 0; i < tetLoad; i++ {
		uid := seq.MustNext()
		if _, found := m[uid]; found {
			t.Fatalf("duplicate key %d exist", uid)
		}
		m[uid] = true
	}
	var elapsed = time.Now().Sub(start).Seconds()
	t.Logf("etcd QPS %.2f/s", float64(tetLoad)/elapsed)
}

// N个并发worker，共享一个生成器, 测试生成id的一致性
func runSeqIDTestConcurrent(t *testing.T, storeTye, label string) {
	var gcnt = 20
	var eachMax = 500000
	var store = createCounterStorage(storeTye, label)
	var seq = NewSeqID(store, DefaultSeqStep)
	if err := seq.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	var workerCtx = NewWorkerContext(eachMax, func() IDGenerator { return seq })
	for i := 0; i < gcnt; i++ {
		workerCtx.Go(t, i)
	}
	workerCtx.Wait()

	var elapsed = workerCtx.Duration().Seconds()
	if !t.Failed() {
		t.Logf("QPS %.2f/s", float64(gcnt*eachMax)/elapsed)
	}
}

// N个并发worker，每个worker单独生成器, 测试生成id的一致性
func runSeqIDTestDistributed(t *testing.T, storeTye, label string) {
	var gcnt = 20
	var eachMax = 500000
	var generator = func() IDGenerator {
		var store = createCounterStorage(storeTye, label)
		var seq = NewSeqID(store, DefaultSeqStep)
		if err := seq.Init(); err != nil {
			t.Fatalf("Init: %v", err)
		}
		return seq
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
}

func TestSeqIDEtcdSimple(t *testing.T) {
	runSeqIDTestSimple(t, "etcd", "/uuid/counter1")
	// Output:
	//  QPS 3209546.48/s
}

func TestSeqIDEtcdConcurrent(t *testing.T) {
	runSeqIDTestConcurrent(t, "etcd", "/uuid/counter2")
	// Output:
	//  QPS 1932994.78/s
}

func TestSeqIDEtcdDistributed(t *testing.T) {
	runSeqIDTestDistributed(t, "etcd", "/uuid/counter3")
	// Output:
	//  QPS 2105700.17/s
}

func TestSeqIDRedisSimple(t *testing.T) {
	runSeqIDTestSimple(t, "redis", "/uuid/counter1")
	// Output:
	//  QPS 4792038.70/s
}

func TestSeqIDRedisConcurrent(t *testing.T) {
	runSeqIDTestConcurrent(t, "redis", "/uuid/counter2")
	// Output:
	//  QPS 2222480.03/s
}

func TestSeqIDRedisDistributed(t *testing.T) {
	runSeqIDTestDistributed(t, "redis", "/uuid/counter3")
	// Output:
	//  QPS 2462537.90/s
}

func TestSeqIDMongoSimple(t *testing.T) {
	runSeqIDTestSimple(t, "mongo", "uuid_counter1")
	// Output:
	//  QPS 3325821.41/s
}

func TestSeqIDMongoConcurrent(t *testing.T) {
	runSeqIDTestConcurrent(t, "mongo", "uuid_counter2")
	// Output:
	//  QPS 1880948.45/s
}

func TestSeqIDMongoDistributed(t *testing.T) {
	runSeqIDTestDistributed(t, "mongo", "uuid_counter3")
	// Output:
	//  QPS 2380477.59/s
}

func TestSeqIDMySQLSimple(t *testing.T) {
	runSeqIDTestSimple(t, "mysql", "uuid_counter1")
	// Output:
	//  QPS 1403723.91/s
}

func TestSeqIDMySQLConcurrent(t *testing.T) {
	runSeqIDTestConcurrent(t, "mysql", "uuid_counter2")
	// Output:
	//  QPS 1084712.19/s
}

func TestSeqIDMySQLDistributed(t *testing.T) {
	runSeqIDTestDistributed(t, "mysql", "uuid_counter3")
	// Output:
	//  QPS 1966049.74/s
}
