// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"google.golang.org/protobuf/proto"
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

// 消息处理器
type PacketHandler func(IPacket) error

// 定义应用层消息接口
type IPacket interface {
	// 消息命令（ID）
	Command() int32
	SetCommand(int32)

	// session序列号
	Seq() uint16
	SetSeq(uint16)

	// 消息类型
	Type() PacketType
	SetType(PacketType)

	// 消息标记
	Flag() PacketFlag
	SetFlag(PacketFlag)

	// 消息错误码
	Errno() int32
	SetErrno(ec int32)

	// 节点ID
	Node() NodeID
	SetNode(NodeID)

	// 引用节点
	Refers() []NodeID
	SetRefers([]NodeID)
	AddRefers(...NodeID)

	// 绑定的endpoint
	Endpoint() MessageEndpoint
	SetEndpoint(MessageEndpoint)

	// clone一个packet
	Clone() IPacket

	// 消息body，仅支持int64/float64/string/bytes/proto.Message类型
	Body() interface{}
	SetBody(v interface{})

	// body类型转换
	BodyToInt() int64
	BodyToFloat() float64
	BodyToString() string
	BodyToBytes() []byte

	// 自动解码为pb消息
	Decode() error

	// 解码body到`msg`里
	DecodeTo(msg proto.Message) error

	// 响应ack消息
	ReplyWith(command int32, body interface{}) error
	Reply(ack proto.Message) error

	// 响应错误码
	RefuseWith(command, errno int32) error
	Refuse(errno int32) error
}
