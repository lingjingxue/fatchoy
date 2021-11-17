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
	PFlagRpc        PacketFlag = 0x20 // RPC标记
)

// 消息编码类型
type PacketType int8

const (
	PTypePacket    PacketType = 0 // 应用消息
	PTypeRoute     PacketType = 1 // 路由消息
	PTypeMulticast PacketType = 2 // 组播消息
)

type PacketHandler func(IPacket) error // 消息处理器

// 定义应用层消息接口
type IPacket interface {
	Command() int32
	SetCommand(int32)

	Seq() uint16
	SetSeq(uint16)

	Type() PacketType
	SetType(PacketType)

	Flag() PacketFlag
	SetFlag(PacketFlag)

	Errno() int32
	SetErrno(ec int32)

	Node() NodeID
	SetNode(NodeID)

	Refers() []NodeID
	SetRefers([]NodeID)

	Endpoint() MessageEndpoint
	SetEndpoint(MessageEndpoint)

	Clone() IPacket

	SetBody(v interface{})
	IBody() interface{}

	BodyToInt() int64
	BodyToFloat() float64
	BodyToString() string
	BodyToBytes() []byte

	DecodeTo(msg proto.Message) error
	Decode() error

	Reply(command int32, body interface{}) error
	ReplyMsg(ack proto.Message) error

	Refuse(errno int32) error
	RefuseWith(command, errno int32) error
}
