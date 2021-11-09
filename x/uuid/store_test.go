// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"sync"
	"testing"
	"time"
)

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
