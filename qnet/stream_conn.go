// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"context"
	"sync"
	"sync/atomic"

	"gopkg.in/qchencc/fatchoy.v1"
	"gopkg.in/qchencc/fatchoy.v1/x/cipher"
	"gopkg.in/qchencc/fatchoy.v1/x/stats"
)

// TcpConn和WsConn的公共基类
type StreamConn struct {
	ctx          context.Context         // chained context
	cancel       context.CancelFunc      // cancel func
	wg           sync.WaitGroup          // wait group
	closing      int32                   // closing flag
	node         fatchoy.NodeID          // node id
	addr         string                  // remote address
	userdata     interface{}             // user data
	codecVersion int                     // codec version
	encrypt      cipher.BlockCryptor     // message encryption
	decrypt      cipher.BlockCryptor     // message decryption
	inbound      chan<- fatchoy.IMessage // inbound message queue
	outbound     chan fatchoy.IMessage   // outbound message queue
	stats        *stats.Stats            // message stats
	errChan      chan error              // error signal
}

func (c *StreamConn) init(parentCtx context.Context, node fatchoy.NodeID, codecVersion int, inbound chan<- fatchoy.IMessage,
	outsize int, errChan chan error, stat *stats.Stats) {
	if stat == nil {
		stat = stats.New(NumStat)
	}
	c.node = node
	c.stats = stat
	c.codecVersion = codecVersion
	c.inbound = inbound
	c.errChan = errChan
	c.outbound = make(chan fatchoy.IMessage, outsize)
	c.ctx, c.cancel = context.WithCancel(parentCtx)
}

func (c *StreamConn) NodeID() fatchoy.NodeID {
	return c.node
}

func (c *StreamConn) SetNodeID(node fatchoy.NodeID) {
	c.node = node
}

func (c *StreamConn) SetRemoteAddr(addr string) {
	c.addr = addr
}

func (c *StreamConn) RemoteAddr() string {
	return c.addr
}

func (c *StreamConn) Stats() *stats.Stats {
	return c.stats
}

func (c *StreamConn) SetEncryptPair(encrypt cipher.BlockCryptor, decrypt cipher.BlockCryptor) {
	c.encrypt = encrypt
	c.decrypt = decrypt
}

func (c *StreamConn) IsClosing() bool {
	return atomic.LoadInt32(&c.closing) == 1
}

func (c *StreamConn) testShouldExit() bool {
	select {
	case <-c.ctx.Done():
		return true
	default:
		return false
	}
}

// 把error投递给监听的channel
func (c *StreamConn) notifyErr(err *Error) {
	if c.errChan != nil {
		select {
		case c.errChan <- err:
		default:
			return
		}
	}
}

func (c *StreamConn) SetUserData(ud interface{}) {
	c.userdata = ud
}

func (c *StreamConn) UserData() interface{} {
	return c.userdata
}
