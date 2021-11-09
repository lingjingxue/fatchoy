// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"sync"
	"testing"
	"time"
)

type IDGenWorkerContext struct {
	wg           sync.WaitGroup
	guard        sync.Mutex
	uuids        map[int64]bool
	eachMaxCount int
	newGen       func() IDGenerator
	startAt      time.Time
	stopAt       time.Time
}

func NewWorkerContext(eachMaxCount int, f func() IDGenerator) *IDGenWorkerContext {
	return &IDGenWorkerContext{
		newGen:       f,
		eachMaxCount: eachMaxCount,
		uuids:        make(map[int64]bool, 10000),
		startAt:      time.Now(),
	}
}

func (ctx *IDGenWorkerContext) serve(t *testing.T, gid int) {
	defer ctx.wg.Done()
	var idGen = ctx.newGen()
	for i := 0; i < ctx.eachMaxCount; i++ {
		id, err := idGen.Next()
		if err != nil {
			t.Fatalf("worker %d generate error: %v", gid, err)
		}
		// fmt.Printf("worker %d generate id %d\n", worker, id)
		ctx.guard.Lock()
		if !putIfAbsent(ctx.uuids, id) {
			ctx.guard.Unlock()
			t.Fatalf("worker %d: tick %d, id %d is already produced by worker", gid, i, id)
		} else {
			ctx.guard.Unlock()
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

func TestSeqIDEtcdSimple(t *testing.T) {
	var store = createEtcdStore(t, "/uuid/ctr101")
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

// N个并发worker，共享一个生成器, 测试生成id的一致性
func TestSeqIDEtcdConcurrent(t *testing.T) {
	var gcnt = 10
	var eachMax = 100000
	var store = createEtcdStore(t, "/uuid/ctr103")
	var seq = NewSeqID(store, 2000)
	var workerCtx = NewWorkerContext(eachMax, func() IDGenerator { return seq })
	for i := 0; i < gcnt; i++ {
		workerCtx.Go(t, i)
	}
	workerCtx.Wait()

	var elapsed = workerCtx.Duration().Seconds()
	if !t.Failed() {
		t.Logf("QPS %.2f/s", float64(gcnt*eachMax)/elapsed)
	}
	// Output:
	//    QPS 2096858.84/s
}

// N个并发worker，每个worker单独生成器, 测试生成id的一致性
func TestSeqIDEtcdDistributed(t *testing.T) {
	var gcnt = 10
	var eachMax = 100000
	var store = createEtcdStore(t, "/uuid/ctr103")
	var workerCtx = NewWorkerContext(eachMax, func() IDGenerator { return NewSeqID(store, 2000) })
	for i := 0; i < gcnt; i++ {
		workerCtx.Go(t, i)
	}
	workerCtx.Wait()

	var elapsed = workerCtx.Duration().Seconds()
	if !t.Failed() {
		t.Logf("QPS %.2f/s", float64(gcnt*eachMax)/elapsed)
	}
	// Output:
	//  QPS 2174803.37/s
}
