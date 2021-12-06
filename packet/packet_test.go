// Copyright © 2020-present simon@qchen.fun All rights reserved.
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

	pkt := New(1234, 1001, 0x1, "")

	pkt.SetBody("hello")

	t.Logf("new: %v", pkt)
	clone := pkt.Clone()
	pkt.Reset()
	t.Logf("reset: %v", pkt)
	t.Logf("clone: %v", clone)

	clone.SetErrno(1002)
	t.Logf("clone: %v", clone)
}
