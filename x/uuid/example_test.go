// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
)

func TestExampleNewV4(t *testing.T) {
	u4, err := NewV4()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(u4)
}

func TestExampleNewV5(t *testing.T) {
	u5, err := NewV5(NamespaceURL, []byte("nu7hat.ch"))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(u5)
}

func TestExampleParseHex(t *testing.T) {
	u, err := ParseHex("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(u)
}

func TestExampleUUID(t *testing.T) {
	rand.Seed(int64(os.Getpid()))
	var store = NewRedisStore("127.0.0.1:6379", "uuid")
	Init(1234, store)
	for i := 0; i < 5; i++ {
		var uid = NextID()
		var uuid = NextUUID()
		t.Logf("uuid short: %v", uid)
		t.Logf("uuid full: %v", uuid)
	}
	for i := 0; i < 2*DefaultSeqStep; i++ {
		NextID()
	}
}
