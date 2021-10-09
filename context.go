// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"context"
)

// 服务的上下文
type ServiceContext struct {
	ctx      context.Context //
	instance Service         // service对象
	queue    chan IMessage   // 收取消息队列
}

func NewServiceContext(ctx context.Context, srv Service, queueSize int) *ServiceContext {
	return &ServiceContext{
		ctx:      ctx,
		instance: srv,
		queue:    make(chan IMessage, queueSize),
	}
}

func (c *ServiceContext) Context() context.Context {
	return c.ctx
}

func (c *ServiceContext) Service() Service {
	return c.instance
}

func (c *ServiceContext) Close() {
	close(c.queue)
	c.queue = nil
}

func (c *ServiceContext) MessageQueue() chan IMessage {
	return c.queue
}
