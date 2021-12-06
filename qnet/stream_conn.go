// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"sync"
	"sync/atomic"

	"qchen.fun/fatchoy"
	"qchen.fun/fatchoy/codec"
	"qchen.fun/fatchoy/x/cipher"
	"qchen.fun/fatchoy/x/stats"
)

// stream connection
type StreamConn struct {
	done     chan struct{}
	wg       sync.WaitGroup         // wait group
	closing  int32                  // closing flag
	node     fatchoy.NodeID         // node id
	addr     string                 // remote address
	userdata interface{}            // user data
	encrypt  cipher.BlockCryptor    // message encryption
	decrypt  cipher.BlockCryptor    // message decryption
	enc      codec.Encoder          // message encode/decode
	inbound  chan<- fatchoy.IPacket // inbound message queue
	outbound chan fatchoy.IPacket   // outbound message queue
	stats    *stats.Stats           // message stats
	errChan  chan error             // error signal
}

func (c *StreamConn) Init(node fatchoy.NodeID, enc codec.Encoder, inbound chan<- fatchoy.IPacket,
	outsize int, errChan chan error, stat *stats.Stats) {
	if stat == nil {
		stat = stats.New(NumStat)
	}
	c.node = node
	c.stats = stat
	c.inbound = inbound
	c.enc = enc
	c.errChan = errChan
	c.outbound = make(chan fatchoy.IPacket, outsize)
	c.done = make(chan struct{})
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
	case <-c.done:
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
