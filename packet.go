// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"github.com/golang/protobuf/proto"
)

// 消息标志位
type PacketFlag uint8

const (
	PFlagCompressed PacketFlag = 0x01 // 压缩
	PFlagEncrypted  PacketFlag = 0x02 // 加密
	PFlagError      PacketFlag = 0x10 // 错误标记
)

// 消息编码类型
type PacketType int8

const (
	PTypeMessage   PacketType = 0 // 4字节长度前缀的消息
	PTypePacket    PacketType = 1 // 以header前缀的消息
	PTypeMulticast PacketType = 2 // 多播
)

type Handler func(IPacket) error // 消息处理器

// 定义应用层消息接口
type IPacket interface {
	Command() int32
	SetCommand(int32)

	Seq() int16
	SetSeq(int16)

	Type() PacketType
	SetType(PacketType)

	Flag() PacketFlag
	SetFlag(PacketFlag)

	Errno() int32
	SetErrno(ec int32)

	Endpoint() MessageEndpoint
	SetEndpoint(MessageEndpoint)

	SetBodyNumber(n int64)
	BodyToNumber() int64

	SetBodyString(s string)
	BodyToString() string

	SetBodyBytes(b []byte)
	BodyToBytes() []byte

	SetBodyMsg(msg proto.Message)
	DecodeTo(msg proto.Message) error

	ReplyWith(command int32, ack proto.Message) error
	Reply(ack proto.Message) error
	ReplyString(command int32, s string) error

	RefuseWith(command, errno int32) error
	Refuse(errno int32) error
}
