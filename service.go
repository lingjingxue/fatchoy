// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"context"

	"qchen.fun/fatchoy/discovery"
	"qchen.fun/fatchoy/x/uuid"
)

// 定义应用层服务接口
type Service interface {
	Type() uint8
	Name() string

	NodeID() NodeID
	SetNodeID(id NodeID)

	Context() *ServiceContext

	Init(*ServiceContext) error
	Startup() error
}

// 服务的上下文
type ServiceContext struct {
	done      chan struct{}     // 同步等待
	workCtx   context.Context   // 用于执行业务
	instance  Service           // service实例
	queue     chan IPacket      // 消息队列
	registrar *discovery.Client // etcd注册
	runId     string            //
}

func NewServiceContext(ctx context.Context, srv Service, queueSize int) *ServiceContext {
	return &ServiceContext{
		workCtx:  ctx,
		instance: srv,
		runId:    uuid.MustCreateGUID(),
		done:     make(chan struct{}, 1),
		queue:    make(chan IPacket, queueSize),
	}
}

func (c *ServiceContext) InitRegistrar(hostAddr, namespace string) error {
	c.registrar = discovery.NewClient(hostAddr, namespace)
	if err := c.registrar.Init(); err != nil {
		return err
	}
	return nil
}

// 业务context
func (c *ServiceContext) WorkCtx() context.Context {
	return c.workCtx
}

// 唯一运行ID
func (c *ServiceContext) RunID() string {
	return c.runId
}

// service实例
func (c *ServiceContext) Instance() Service {
	return c.instance
}

// 服务注册器
func (c *ServiceContext) Registrar() *discovery.Client {
	return c.registrar
}

// 消息队列，仅接收
func (c *ServiceContext) InboundQueue() chan<- IPacket {
	return c.queue
}

// 消息队列，仅消费
func (c *ServiceContext) MessageQueue() <-chan IPacket {
	return c.queue
}

// 投递一条消息到context
func (c *ServiceContext) Send(pkt IPacket) {
	c.queue <- pkt // block send
}

// 等待close完成
func (c *ServiceContext) WaitDone() <-chan struct{} {
	return c.done
}

func (c *ServiceContext) Close() {
	c.registrar.Close()
	c.registrar = nil
	close(c.queue)
	c.queue = nil

	select {
	case c.done <- struct{}{}:
	default:
	}
}
