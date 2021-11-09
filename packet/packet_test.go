// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package packet

import (
	"encoding/json"
	"testing"
	"unsafe"
)

func TestNewPacket(t *testing.T) {
	size := unsafe.Sizeof(Packet{})
	println("sizeof packet:", size)

	pkt := New(1234, 1001, 0, 0x1, "")

	pkt.SetBodyString("hello")
	
	t.Logf("new: %v", pkt)
	clone := pkt.Clone()
	pkt.Reset()
	t.Logf("reset: %v", pkt)
	t.Logf("clone: %v", clone)

	clone.SetErrno(1002)
	t.Logf("clone: %v", clone)
}

func TestPacketEncode(t *testing.T) {
	pkt := New(1234, 1001, 0, 0x1, "hello")
	if err := Encode(pkt, 4); err != nil {
		t.Fatalf("%v", err)
	}
	data, err := json.Marshal(pkt)
	if err != nil {
		t.Fatalf("%v", err)
	}
	var pkt2 = Make()
	if err := json.Unmarshal(data, pkt2); err != nil {
		t.Fatalf("%v", err)
	}
	if err := Decode(pkt2); err != nil {
		t.Fatalf("%v", err)
	}
}
