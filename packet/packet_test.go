// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package packet

import (
	"testing"
	"unsafe"
)

func TestNewPacket(t *testing.T) {
	size := unsafe.Sizeof(Packet{})
	println("sizeof packet:", size)

	pkt := New(1234, 1001, 0, 0x1, "")
	if _, err := pkt.EncodeBodyToBytes(); err != nil {
		t.Fail()
	}

	pkt.SetBodyString("hello")
	if _, err := pkt.EncodeBodyToBytes(); err != nil {
		t.Fail()
	}

	t.Logf("new: %v", pkt)
	clone := pkt.Clone()
	pkt.Reset()
	t.Logf("reset: %v", pkt)
	t.Logf("clone: %v", clone)

	clone.SetErrno(1002)
	t.Logf("clone: %v", clone)
}
