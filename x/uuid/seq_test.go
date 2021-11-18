// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

//go:build ignore

package uuid

import (
	"testing"
)


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
