// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package packet

import (
	"testing"
	"unsafe"

	"qchen.fun/fatchoy"
)

type AAA struct {
	Cmd     int32
	Seq     uint32
	Flag    uint16
	Reserve uint16
	Node    fatchoy.NodeID
	Refer   []fatchoy.NodeID
}

func TestNewPacket(t *testing.T) {
	var a PacketBase
	size := unsafe.Sizeof(a)
	println("sizeof packet:", size)

	var pkt = FakePacket{}

	pkt.SetBody("hello")

	t.Logf("new: %v", pkt)
	clone := pkt.Clone()
	pkt.Reset()
	t.Logf("reset: %v", pkt)
	t.Logf("clone: %v", clone)

	clone.SetErrno(1002)
	t.Logf("clone: %v", clone)
}
