// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

import (
	"fmt"
	"reflect"

	"qchen.fun/fatchoy"
	"qchen.fun/fatchoy/qlog"
)

// 消息派发
type MessageHandlers struct {
	fns map[int32][]fatchoy.PacketHandler
}

func NewMsgHandlers() MessageHandlers {
	return MessageHandlers{
		fns: make(map[int32][]fatchoy.PacketHandler),
	}
}

var defH = NewMsgHandlers()

func MsgHandlers() MessageHandlers {
	return defH
}

// 注册一个
func (d *MessageHandlers) Register(msgId int32, handler fatchoy.PacketHandler) {
	d.fns[msgId] = append(d.fns[msgId], handler)
}

// 取消所有
func (d *MessageHandlers) DeregisterAll(msgId int32) {
	delete(d.fns, msgId)
}

// 取消单个注册
func (d *MessageHandlers) DeregisterOne(msgId int32, handler fatchoy.PacketHandler) {
	var pointer = reflect.ValueOf(handler).Pointer()
	var retain []fatchoy.PacketHandler
	for _, h := range d.fns[msgId] {
		if reflect.ValueOf(h).Pointer() != pointer {
			retain = append(retain, h)
		}
	}
	d.fns[msgId] = retain
}

func (d *MessageHandlers) Dispatch(pkt fatchoy.IPacket) error {
	var cnt = 0
	var err error
	var msgId = pkt.Command()
	var handlers = d.fns[msgId]
	for _, h := range handlers {
		cnt++
		if er := h(pkt); er != nil {
			err = er
			qlog.Errorf("execute handler for msg %d: %v", msgId, er)
		}
	}
	if cnt == 0 {
		return fmt.Errorf("no handlers executed for msg %v", msgId)
	}
	return err
}
