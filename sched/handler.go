// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

import (
	"fmt"
	"reflect"

	"qchen.fun/fatchoy"
	"qchen.fun/fatchoy/l0g"
)

// 消息派发
type MessageHandlers struct {
	handlers map[int32][]fatchoy.PacketHandler
}

func NewMsgHandlers() MessageHandlers {
	return MessageHandlers{
		handlers: make(map[int32][]fatchoy.PacketHandler),
	}
}

var defH = NewMsgHandlers()

func MessageDispatcher() MessageHandlers {
	return defH
}

// 注册一个
func (d *MessageHandlers) Register(msgId int32, handler fatchoy.PacketHandler) {
	d.handlers[msgId] = append(d.handlers[msgId], handler)
}

// 取消所有
func (d *MessageHandlers) DeregisterAll(msgId int32) {
	delete(d.handlers, msgId)
}

// 取消单个注册
func (d *MessageHandlers) DeregisterOne(msgId int32, handler fatchoy.PacketHandler) {
	var pointer = reflect.ValueOf(handler).Pointer()
	var retain []fatchoy.PacketHandler
	for _, h := range d.handlers[msgId] {
		if reflect.ValueOf(h).Pointer() != pointer {
			retain = append(retain, h)
		}
	}
	d.handlers[msgId] = retain
}

func (d *MessageHandlers) Dispatch(pkt fatchoy.IPacket) error {
	var msgId = pkt.Command()
	var handlers = d.handlers[msgId]
	if len(handlers) == 0 {
		return fmt.Errorf("no handlers executed for msg %v", msgId)
	}
	var err error
	for _, h := range handlers {
		if er := h(pkt); er != nil {
			err = er
			l0g.Errorf("execute handler for msg %d: %v", msgId, er)
		}
	}
	return err
}
