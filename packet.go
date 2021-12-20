// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"google.golang.org/protobuf/proto"
)

// 消息标志位
type PacketFlag uint16

const (
	// 低8位用于表达一些传输flag
	PFlagCompressed PacketFlag = 0x0001 // 压缩
	PFlagEncrypted  PacketFlag = 0x0002 // 加密
	PFlagError      PacketFlag = 0x0004 // 错误标记
	PFlagRpc        PacketFlag = 0x0010 // RPC标记

	// 高8位用于表达消息类型
	PFlagRoute     PacketFlag = 0x0100 // 路由消息
	PFlagMulticast PacketFlag = 0x0200 // 组播消息
)

func (g PacketFlag) Has(n PacketFlag) bool {
	return g&n != 0
}

func (g PacketFlag) Clear(n PacketFlag) PacketFlag {
	return g &^ n
}

// 消息处理器
type PacketHandler func(IPacket) error

// 定义应用层消息接口
type IPacket interface {
	// 消息命令（ID）
	Command() int32
	SetCommand(int32)

	// session序列号
	Seq() uint32
	SetSeq(uint32)

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

	// 消息body，仅支持int32/int64/float64/string/bytes/proto.Message类型
	Body() interface{}
	SetBody(v interface{})

	BodyToBytes() []byte

	// 自动解码为pb消息
	Decode() error

	// 解码body到`msg`里
	DecodeTo(msg proto.Message) error

	// 响应ack消息
	ReplyWith(command int32, body interface{}) error
	Reply(ack proto.Message) error

	// 响应错误码
	RefuseWith(command int32, errno int32) error
	Refuse(errno int32) error
}
