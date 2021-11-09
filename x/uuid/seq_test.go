// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"testing"
	"time"
)

func TestSeqIDEtcdSimple(t *testing.T) {
	var store = createEtcdStore("/uuid/ctr101")
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
	var store = createEtcdStore("/uuid/ctr103")
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
	var store = createEtcdStore("/uuid/ctr103")
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
