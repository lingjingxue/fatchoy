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

func TestSeqIDEtcdSimple(t *testing.T) {
	cli := createEtcdClient()
	var store = NewEtcdStore(context.Background(), cli, "/uuid/ctr101")
	var seq = NewSeqID(store, 2000)
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
	// Output:
	//   etcd QPS: 3236293.09/s
}

func createEtcdIDGen(key string, t *testing.T) IDGenerator {
	cli := createEtcdClient()
	var store = NewEtcdStore(context.Background(), cli, key)
	var seq = NewSeqID(store, 2000)
	if err := seq.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return seq
}

// N个并发worker，共享一个生成器, 测试生成id的一致性
func TestSeqIDEtcdConcurrent(t *testing.T) {
	var (
		wg      sync.WaitGroup
		guard   sync.Mutex
		gcnt    = 20
		eachMax = 200000
		m       = make(map[int64]int, 10000)
	)
	ctx := newWorkerContext(&wg, &guard, m, eachMax)
	ctx.idGenCreator = func() IDGenerator {
		return createEtcdIDGen("/uuid/ctr102", t)
	}
	var start = time.Now()
	for i := 0; i < gcnt; i++ {
		wg.Add(1)
		go runIDWorker(i, ctx, t)
	}
	wg.Wait()
	if n := len(m); n != gcnt*eachMax {
		t.Fatalf("duplicate id found, %d != %d", n, gcnt*eachMax)
	}
	var elapsed = time.Now().Sub(start).Seconds()
	if !t.Failed() {
		t.Logf("QPS %.2f/s", float64(gcnt*eachMax)/elapsed)
	}
	// Output:
	//    QPS 2096858.84/s
}

// N个并发worker，每个worker单独生成器, 测试生成id的一致性
func TestSeqIDEtcdDistributed(t *testing.T) {
	var (
		wg      sync.WaitGroup
		guard   sync.Mutex
		gcnt    = 20
		eachMax = 200000
		m       = make(map[int64]int, 10000)
	)

	var start = time.Now()
	for i := 0; i < gcnt; i++ {
		ctx := newWorkerContext(&wg, &guard, m, eachMax)
		ctx.idGenCreator = func() IDGenerator {
			return createEtcdIDGen("/uuid/ctr103", t)
		}
		wg.Add(1)
		go runIDWorker(i, ctx, t)
	}
	wg.Wait()
	if n := len(m); n != gcnt*eachMax {
		t.Fatalf("duplicate id found, %d != %d", n, gcnt*eachMax)
	}
	var elapsed = time.Now().Sub(start).Seconds()
	if !t.Failed() {
		t.Logf("QPS %.2f/s", float64(gcnt*eachMax)/elapsed)
	}
	// Output:
	//  QPS 2174803.37/s
}

func TestSeqIDRedis(t *testing.T) {
	var store = NewRedisStore(context.Background(), redisAddr, "uuid:ctr101")
	var seq = NewSeqID(store, 2000)
	if err := seq.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	var m = make(map[int64]bool)
	var start = time.Now()
	const tetLoad = 2000000
	for i := 0; i < tetLoad; i++ {
		uid := seq.MustNext()
		if _, found := m[uid]; found {
			t.Fatalf("key %d exist", uid)
		}
		m[uid] = true
	}
	var elapsed = time.Now().Sub(start).Seconds()
	t.Logf("redis QPS: %.2f/s", float64(tetLoad)/elapsed)
	// Output:
	//   redis QPS: 4852265.36/s
}

func createRedisIDGen(key string, t *testing.T) IDGenerator {
	var store = NewRedisStore(context.Background(), redisAddr, key)
	var seq = NewSeqID(store, 2000)
	if err := seq.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return seq
}

// N个并发worker，共享一个生成器, 测试生成id的一致性
func TestSeqIDRedisConcurrent(t *testing.T) {
	var (
		wg      sync.WaitGroup
		guard   sync.Mutex
		gcnt    = 20
		eachMax = 200000
		m       = make(map[int64]int, 10000)
	)
	ctx := newWorkerContext(&wg, &guard, m, eachMax)
	ctx.idGenCreator = func() IDGenerator {
		return createRedisIDGen("uuid:ctr102", t)
	}
	var start = time.Now()
	for i := 0; i < gcnt; i++ {
		wg.Add(1)
		go runIDWorker(i, ctx, t)
	}
	wg.Wait()
	if n := len(m); n != gcnt*eachMax {
		t.Fatalf("duplicate id found, %d != %d", n, gcnt*eachMax)
	}
	var elapsed = time.Now().Sub(start).Seconds()
	if !t.Failed() {
		t.Logf("QPS %.2f/s", float64(gcnt*eachMax)/elapsed)
	}
	// Output:
	//    QPS 2164858.64/s
}

// N个并发worker，每个worker单独一个生成器, 测试生成id的一致性
func TestSeqIDRedisDistributed(t *testing.T) {
	var (
		wg      sync.WaitGroup
		guard   sync.Mutex
		gcnt    = 20
		eachMax = 200000
		m       = make(map[int64]int, 10000)
	)

	var start = time.Now()
	for i := 0; i < gcnt; i++ {
		ctx := newWorkerContext(&wg, &guard, m, eachMax)
		ctx.idGenCreator = func() IDGenerator {
			return createRedisIDGen("uuid:ctr103", t)
		}
		wg.Add(1)
		go runIDWorker(i, ctx, t)
	}
	wg.Wait()
	if n := len(m); n != gcnt*eachMax {
		t.Fatalf("duplicate id found, %d != %d", n, gcnt*eachMax)
	}
	var elapsed = time.Now().Sub(start).Seconds()
	t.Logf("QPS %.2f/s", float64(gcnt*eachMax)/elapsed)
	// Output:
	//  QPS 2180111.72/s
}

type idGenWorkerContext struct {
	wg           *sync.WaitGroup
	guard        *sync.Mutex
	dict         map[int64]int
	maxCount     int
	idGen        IDGenerator
	idGenCreator func() IDGenerator
}

func newWorkerContext(wg *sync.WaitGroup, guard *sync.Mutex, dict map[int64]int, maxCount int) *idGenWorkerContext {
	return &idGenWorkerContext{
		wg:       wg,
		guard:    guard,
		dict:     dict,
		maxCount: maxCount,
	}
}

type StorageGen struct {
	store Storage
}

func NewStorageGen(store Storage) IDGenerator {
	return &StorageGen{
		store: store,
	}
}

func (s *StorageGen) Next() (int64, error) {
	return s.store.Incr()
}

type IDGenerator interface {
	Next() (int64, error)
}

func runIDWorker(worker int, ctx *idGenWorkerContext, t *testing.T) {
	defer ctx.wg.Done()
	var idGen = ctx.idGen
	if idGen == nil {
		idGen = ctx.idGenCreator()
	}
	for i := 0; i < ctx.maxCount; i++ {
		id, err := idGen.Next()
		if err != nil {
			t.Fatalf("worker %d storage error: %v", worker, err)
		}
		// fmt.Printf("worker %d generate id %d\n", worker, id)
		ctx.guard.Lock()
		if old, found := ctx.dict[id]; found {
			ctx.guard.Unlock()
			t.Fatalf("worker %d: tick %d, id %d is already produced by worker %d", worker, i, id, old)
		} else {
			ctx.dict[id] = worker
			ctx.guard.Unlock()
		}
	}
}
