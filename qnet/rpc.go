// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"qchen.fun/fatchoy"
	"qchen.fun/fatchoy/codes"
	"qchen.fun/fatchoy/l0g"
	"qchen.fun/fatchoy/packet"
)

type RpcHandler func(proto.Message, int32) error

// RPC上下文
type RpcContext struct {
	dest     fatchoy.NodeID   // 目标节点
	req      proto.Message    // 请求消息
	ack      fatchoy.IPacket  // 响应packet
	deadline time.Time        // 超时
	cb       RpcHandler       // 异步回调
	done     chan *RpcContext // Strobes when RPC is completed
}

func NewRpcContext(node fatchoy.NodeID, req proto.Message, cb RpcHandler) *RpcContext {
	return &RpcContext{
		dest: node,
		req:  req,
		cb:   cb,
	}
}

func (r *RpcContext) DecodeAck() (proto.Message, error) {
	var pkt = r.ack
	if ec := pkt.Errno(); ec > 0 {
		return nil, fmt.Errorf("%v", codes.Code(ec))
	}
	if err := pkt.Decode(); err != nil {
		l0g.Errorf("decode rpc %d response %v", pkt.Command(), err)
		return nil, err
	}
	var ack = pkt.Body().(proto.Message)
	return ack, nil
}

func (r *RpcContext) notify() {
	if r.done == nil {
		return
	}
	select {
	case r.done <- r:
		// ok
	default:
		// We don't want to block here. It is the caller's responsibility to make
		// sure the channel has enough buffer space.
	}
}

func (r *RpcContext) run(pkt fatchoy.IPacket) error {
	r.ack = pkt
	r.notify()
	if r.cb != nil {
		if ec := pkt.Errno(); ec > 0 {
			return r.cb(nil, ec)
		}
		if ack, err := r.DecodeAck(); err != nil {
			l0g.Errorf("decode rpc %d response %v", pkt.Command(), err)
			return r.cb(nil, int32(codes.InternalError))
		} else {
			return r.cb(ack, 0)
		}
	}
	return nil
}

// RPC调用client
type RpcClient struct {
	ctx          context.Context        //
	wg           sync.WaitGroup         //
	guard        sync.Mutex             // 多线程guard
	pendingCtx   map[uint16]*RpcContext // 待响应的RPC
	pendingQueue chan fatchoy.IPacket   // 待发送消息队列
	expired      []*RpcContext          // 超时的
	counter      uint16                 // 序列号生成
}

func NewRpcClient(ctx context.Context, queueSize int) *RpcClient {
	return &RpcClient{
		ctx:          ctx,
		expired:      make([]*RpcContext, 0, 8),
		pendingQueue: make(chan fatchoy.IPacket, queueSize),
		pendingCtx:   make(map[uint16]*RpcContext),
	}
}

func (c *RpcClient) PendingQueue() <-chan fatchoy.IPacket {
	return c.pendingQueue
}

func (c *RpcClient) Go() {
	c.wg.Add(1)
	go c.reaper()
}

func (c *RpcClient) AsyncCall(node fatchoy.NodeID, req proto.Message, cb RpcHandler) error {
	var rpc = NewRpcContext(node, req, cb)
	c.makeCall(rpc)
	return nil
}

func (c *RpcClient) Call(node fatchoy.NodeID, req proto.Message) *RpcContext {
	var rpc = NewRpcContext(node, req, nil)
	rpc.done = make(chan *RpcContext, 1)
	rpc = <-c.makeCall(rpc).done
	return rpc
}

func (c *RpcClient) makeCall(ctx *RpcContext) *RpcContext {
	c.guard.Lock()
	defer c.guard.Unlock()

	ctx.deadline = time.Now().Add(time.Minute) // 1分钟ttl
	c.counter++
	if c.counter == 0 {
		c.counter++
	}
	var seq = c.counter
	c.pendingCtx[seq] = ctx

	var reqMsgID = packet.GetMessageIDOf(ctx.req)
	var pkt = packet.New(reqMsgID, seq, fatchoy.PFlagRpc, ctx.req)
	pkt.SetType(fatchoy.PTypePacket)
	pkt.SetNode(ctx.dest)
	c.pendingQueue <- pkt // this may block

	return ctx
}

func (c *RpcClient) stripRpcContext(seq uint16) *RpcContext {
	c.guard.Lock()
	ctx, found := c.pendingCtx[seq]
	if found {
		delete(c.pendingCtx, seq)
	}
	c.guard.Unlock()
	return ctx
}

func (c *RpcClient) stripExpired() []*RpcContext {
	c.guard.Lock()
	var expired = c.expired
	c.expired = make([]*RpcContext, 0, 8)
	c.guard.Unlock()
	return expired
}

func (c *RpcClient) ReapTimeout() int{
	var expired = c.stripExpired()
	var n = len(expired)
	for _, ctx := range expired {
		var reqMsgID = packet.GetMessageIDOf(ctx.req)
		var ackMsgID = packet.GetPairingAckID(reqMsgID)
		var pkt = packet.Make()
		packet.New(ackMsgID, 0, fatchoy.PFlagRpc, ctx.req)
		pkt.SetType(fatchoy.PTypePacket)
		pkt.SetErrno(int32(codes.RequestTimeout))
		if err := ctx.run(pkt); err != nil {
			l0g.Errorf("rpc %d timed-out done: %v", ackMsgID, err)
		}
	}
	return n
}

// 在主线程运行
func (c *RpcClient) Dispatch(pkt fatchoy.IPacket) error {
	var ctx = c.stripRpcContext(pkt.Seq())
	if ctx != nil {
		return ctx.run(pkt)
	}
	return fmt.Errorf("session %d message %d rpc context not found ", pkt.Seq(), pkt.Command())
}

func (c *RpcClient) reapTimeout(now time.Time) {
	c.guard.Lock()
	defer c.guard.Unlock()
	for seq, ctx := range c.pendingCtx {
		if now.After(ctx.deadline) {
			c.expired = append(c.expired, ctx)
			delete(c.pendingCtx, seq)
		}
	}
}

// 处理超时
func (c *RpcClient) reaper() {
	defer c.wg.Done()
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()
	for {
		select {
		case now := <-ticker.C:
			c.reapTimeout(now)

		case <-c.ctx.Done():
			return
		}
	}
}
