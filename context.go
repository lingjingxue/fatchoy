// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"context"
)

// 服务的上下文
type ServiceContext struct {
	done     chan struct{}   // 同步等待
	ctx      context.Context // context对象
	instance Service         // service实例
	queue    chan IPacket    // 消息队列
}

func NewServiceContext(ctx context.Context, srv Service, queueSize int) *ServiceContext {
	return &ServiceContext{
		ctx:      ctx,
		instance: srv,
		done:     make(chan struct{}, 1),
		queue:    make(chan IPacket, queueSize),
	}
}

func (c *ServiceContext) Context() context.Context {
	return c.ctx
}

func (c *ServiceContext) Instance() Service {
	return c.instance
}

func (c *ServiceContext) Close() {
	close(c.queue)
	c.queue = nil
	c.done <- struct{}{}
}

// 等待close完成
func (c *ServiceContext) WaitDone() <-chan struct{} {
	return c.done
}

// 消息队列
func (c *ServiceContext) MessageQueue() chan IPacket {
	return c.queue
}
