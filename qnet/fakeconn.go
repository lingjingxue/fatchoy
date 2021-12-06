// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"net"

	"qchen.fun/fatchoy"
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

func (c *FakeConn) Go(flag fatchoy.EndpointFlag) {
}

func (c *FakeConn) Close() error {
	return nil
}

func (c *FakeConn) ForceClose(error) {
}
