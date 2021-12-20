// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"container/heap"
	"fmt"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"qchen.fun/fatchoy"
	"qchen.fun/fatchoy/codes"
	"qchen.fun/fatchoy/log"
	"qchen.fun/fatchoy/packet"
)

type RpcHandler func(proto.Message, int32) error

// RPC上下文
type RpcContext struct {
	deadline int64            // 超时
	index    int32            // 在最小堆里的索引
	seq      uint32           //
	dest     fatchoy.NodeID   // 目标节点
	req      proto.Message    // 请求消息
	ack      fatchoy.IPacket  // 响应packet
	action   RpcHandler       // 异步回调
	done     chan *RpcContext // Strobes when RPC is completed
}

func NewRpcContext(node fatchoy.NodeID, req proto.Message, action RpcHandler) *RpcContext {
	return &RpcContext{
		dest:   node,
		req:    req,
		action: action,
	}
}

func (r *RpcContext) nilOut() {
	r.req = nil
	r.ack = nil
	r.action = nil
	r.done = nil
}

func (r *RpcContext) DecodeAck() (proto.Message, error) {
	var pkt = r.ack
	if ec := pkt.Errno(); ec > 0 {
		return nil, fmt.Errorf("%v", codes.Code(ec))
	}
	if err := pkt.Decode(); err != nil {
		log.Errorf("decode rpc %d response %v", pkt.Command(), err)
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
	defer r.nilOut()
	r.ack = pkt
	r.notify()
	if r.action != nil {
		if ec := pkt.Errno(); ec > 0 {
			return r.action(nil, ec)
		}
		if ack, err := r.DecodeAck(); err != nil {
			log.Errorf("decode rpc %d response %v", pkt.Command(), err)
			return r.action(nil, int32(codes.InternalError))
		} else {
			return r.action(ack, 0)
		}
	}
	return nil
}

// RPC client stub
type RpcClient struct {
	done         chan struct{}
	wg           sync.WaitGroup         //
	state        fatchoy.State          //
	guard        sync.Mutex             // 多线程guard
	minheap      rpcNodeHeap            // 使用最小堆减少主动检测超时节点
	pendingCtx   map[uint32]*RpcContext // 待响应的RPC
	pendingQueue chan fatchoy.IPacket   // 待发送消息队列
	expired      []*RpcContext          // 超时的
	counter      uint32                 // 序列号生成
}

func NewRpcClient(queueSize int) *RpcClient {
	return &RpcClient{
		done:         make(chan struct{}),
		pendingQueue: make(chan fatchoy.IPacket, queueSize),
		pendingCtx:   make(map[uint32]*RpcContext),
	}
}

func (c *RpcClient) PendingQueue() <-chan fatchoy.IPacket {
	return c.pendingQueue
}

func (c *RpcClient) IsRunning() bool {
	return c.state.IsRunning()
}

func (c *RpcClient) Start() {
	switch state := c.state.Get(); state {
	case fatchoy.StateInit:
		if c.state.CAS(fatchoy.StateInit, fatchoy.StateStarted) {
			var ready = make(chan struct{}, 1)
			c.wg.Add(1)
			go c.reaper(ready)
			<-ready
			c.state.Set(fatchoy.StateRunning)
		}

	case fatchoy.StateRunning:
		return

	default:
		log.Panicf("invalid worker state %v", state)
	}
}

func (c *RpcClient) Shutdown() {
	switch c.state.Get() {
	case fatchoy.StateShutdown, fatchoy.StateTerminated:
		return
	}
	c.state.Set(fatchoy.StateShutdown)
	close(c.done)
	c.wg.Wait()
	c.pendingCtx = nil
	c.pendingQueue = nil
	c.state.Set(fatchoy.StateTerminated)
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

	ctx.deadline = time.Now().Add(time.Minute).UnixNano() / int64(time.Millisecond) // 1分钟ttl
	c.counter++
	if c.counter == 0 {
		c.counter++
	}
	var seq = c.counter
	ctx.seq = seq
	c.pendingCtx[seq] = ctx
	heap.Push(&c.minheap, ctx)

	var reqMsgID = packet.GetMessageIDOf(ctx.req)
	var pkt = packet.New(reqMsgID, seq, fatchoy.PFlagRpc, ctx.req)
	pkt.SetNode(ctx.dest)
	c.pendingQueue <- pkt // this may block

	return ctx
}

func (c *RpcClient) stripPending(seq uint32) *RpcContext {
	c.guard.Lock()
	ctx, found := c.pendingCtx[seq]
	if found {
		delete(c.pendingCtx, seq)
		heap.Remove(&c.minheap, int(ctx.index))
		ctx.index = -1
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

func (c *RpcClient) ReapTimeout() int {
	var expired = c.stripExpired()
	var n = len(expired)
	for _, ctx := range expired {
		var reqMsgID = packet.GetMessageIDOf(ctx.req)
		var ackMsgID = packet.GetPairingAckID(reqMsgID)
		var pkt = packet.Make()
		packet.New(ackMsgID, 0, fatchoy.PFlagRpc, ctx.req)
		pkt.SetErrno(int32(codes.RequestTimeout))
		if err := ctx.run(pkt); err != nil {
			log.Errorf("rpc %d timed-out done: %v", ackMsgID, err)
		}
	}
	return n
}

// 在主线程运行
func (c *RpcClient) Dispatch(pkt fatchoy.IPacket) error {
	var ctx = c.stripPending(pkt.Seq())
	if ctx != nil {
		return ctx.run(pkt)
	}
	return fmt.Errorf("session %d message %d rpc context not found ", pkt.Seq(), pkt.Command())
}

func (c *RpcClient) reapTimeout(now int64) {
	c.guard.Lock()
	defer c.guard.Unlock()
	for len(c.minheap) > 0 {
		var ctx = c.minheap[0] // peek first item of heap
		if now < ctx.deadline {
			break // no new context expired
		}
		heap.Pop(&c.minheap)
		delete(c.pendingCtx, ctx.seq)
		ctx.index = -1
		c.expired = append(c.expired, ctx)
	}
}

// 处理超时
func (c *RpcClient) reaper(ready chan struct{}) {
	defer c.wg.Done()

	var ticker = time.NewTicker(time.Second)
	defer ticker.Stop()
	<-ready

	for {
		select {
		case now := <-ticker.C:
			c.reapTimeout(now.UnixNano() / int64(time.Millisecond))

		case <-c.done:
			return
		}
	}
}

type rpcNodeHeap []*RpcContext

func (q rpcNodeHeap) Len() int {
	return len(q)
}

func (q rpcNodeHeap) Less(i, j int) bool {
	return q[i].deadline < q[j].deadline
}

func (q rpcNodeHeap) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].index = int32(i)
	q[j].index = int32(j)
}

func (q *rpcNodeHeap) Push(x interface{}) {
	v := x.(*RpcContext)
	v.index = int32(len(*q))
	*q = append(*q, v)
}

func (q *rpcNodeHeap) Pop() interface{} {
	old := *q
	n := len(old)
	if n > 0 {
		v := old[n-1]
		v.index = -1 // for safety
		*q = old[:n-1]
		return v
	}
	return nil
}
