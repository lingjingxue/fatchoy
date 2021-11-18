// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
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

// empty lock
type NoLock struct {
}

func (l *NoLock) Lock() {
}

func (l *NoLock) Unlock() {
}

func putIfAbsent(guard sync.Locker, uuids map[int64]bool, id int64) bool {
	guard.Lock()
	defer guard.Unlock()

	if _, found := uuids[id]; !found {
		uuids[id] = true
		return true
	} else {
		return false
	}
}

type IDGenWorkerContext struct {
	wg           sync.WaitGroup
	guard        sync.Mutex
	uuids        map[int64]bool
	eachMaxCount int
	genMaker     func() IDGenerator
	startAt      time.Time
	stopAt       time.Time
}

func NewWorkerContext(eachMaxCount int, f func() IDGenerator) *IDGenWorkerContext {
	return &IDGenWorkerContext{
		genMaker:     f,
		eachMaxCount: eachMaxCount,
		uuids:        make(map[int64]bool, 10000),
		startAt:      time.Now(),
	}
}

func (ctx *IDGenWorkerContext) serve(t *testing.T, gid int) {
	defer ctx.wg.Done()
	var idGen = ctx.genMaker()
	for i := 0; i < ctx.eachMaxCount; i++ {
		id, err := idGen.Next()
		if err != nil {
			t.Fatalf("worker %d generate error: %v", gid, err)
		}
		// fmt.Printf("worker %d generate id %d\n", worker, id)
		if !putIfAbsent(&ctx.guard, ctx.uuids, id) {
			t.Fatalf("worker %d: tick %d, id %d is already produced by worker", gid, i, id)
		}
	}
}

func (ctx *IDGenWorkerContext) Go(t *testing.T, gid int) {
	ctx.wg.Add(1)
	go ctx.serve(t, gid)
}

func (ctx *IDGenWorkerContext) Wait() {
	ctx.wg.Wait()
	ctx.stopAt = time.Now()
}

func (ctx *IDGenWorkerContext) Duration() time.Duration {
	return ctx.stopAt.Sub(ctx.startAt)
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
