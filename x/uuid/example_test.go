// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"testing"
)

func TestExampleUUID(t *testing.T) {
	var store = createCounterStorage("redis", "/uuid/eg11")
	if err := Init(0, store); err != nil {
		t.Fatalf("%v", err)
	}
	t.Logf("seq id: %d", NextID())
	t.Logf("uuid: %d", NextUUID())
	t.Logf("guid: %s", NextGUID())
}

func TestExampleParseHex(t *testing.T) {
	u, err := ParseHex("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Logf("%v", u)
}

func TestExampleSeqIDEtcd(t *testing.T) {
	runSeqIDTestSimple(t, "etcd", "/uuid/counter101")
}

func TestExampleSeqIDRedis(t *testing.T) {
	runSeqIDTestSimple(t, "redis", "/uuid/counter101")
}
