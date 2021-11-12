// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"github.com/golang/protobuf/proto"
)

type PacketFlag uint8

const (
	PFlagCompressed PacketFlag = 0x01
	PFlagEncrypted  PacketFlag = 0x02
	PFlagError      PacketFlag = 0x10
)

// message type
type PacketType int8

const (
	PTypeMessage   PacketType = 0
	PTypeRoute     PacketType = 1
	PTypeMulticast PacketType = 2
)

type Handler func(IPacket) error // 消息处理器
type Filter func(IPacket) bool   // 过滤器

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
