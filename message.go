// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"github.com/golang/protobuf/proto"
)

type PacketFlag uint8

const (
	PacketFlagError      PacketFlag = 1 << 0
	PacketFlagCompressed PacketFlag = 1 << 1
	PacketFlagEncrypted  PacketFlag = 1 << 2
)

// message type
type PacketType int8

type Handler func(IMessage) error // 消息处理器
type Filter func(IMessage) bool   // 过滤器

// 定义应用层消息接口
type IMessage interface {
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

	Body() interface{}

	SetBodyNumber(n int64)
	BodyAsNumber() int64

	SetBodyString(s string)
	BodyAsString() string

	SetBodyBytes(b []byte)
	BodyAsBytes() []byte

	SetBodyMsg(msg proto.Message)
	BodyAsMsg() proto.Message

	EncodeBodyToBytes() ([]byte, error)
	DecodeTo(msg proto.Message) error

	Reply(ack proto.Message) error
	ReplyString(command int32, s string) error
	ReplyBytes(command int32, b []byte) error

	Refuse(errno int32) error
	RefuseCommand(command, errno int32) error
}
