// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"net"

	"gopkg.in/qchencc/fatchoy.v1"
)

// a fake endpoint
type FakeConn struct {
	StreamConn
}

func NewFakeConn(node fatchoy.NodeID, addr string) fatchoy.Endpoint {
	return &FakeConn{
		StreamConn: StreamConn{
			node: node,
			addr: addr,
		},
	}
}

func (c *FakeConn) RawConn() net.Conn {
	return nil
}

func (c *FakeConn) SendPacket(fatchoy.IPacket) error {
	return nil
}

func (c *FakeConn) Go(bool, bool) {
}

func (c *FakeConn) Close() error {
	return nil
}

func (c *FakeConn) ForceClose(error) {
}
