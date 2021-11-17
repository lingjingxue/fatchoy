// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package packet

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"gopkg.in/qchencc/fatchoy.v1"
)

// Packet表示一个应用层消息
type Packet struct {
	Cmd      int32                   `json:"cmd"`            // 协议ID
	Seq_     uint16                  `json:"seq"`            // 序列号
	Type_    fatchoy.PacketType      `json:"typ,omitempty"`  // 类型
	Flg      fatchoy.PacketFlag      `json:"flg,omitempty"`  // 标志位
	Node_    fatchoy.NodeID          `json:"node,omitempty"` // 源/目标节点
	Body     interface{}             `json:"body,omitempty"` // 消息内容，int64/float64/string/bytes/proto.Message
	Refer    []fatchoy.NodeID        `json:"ref,omitempty"`  // 组播session列表
	endpoint fatchoy.MessageEndpoint // 关联的endpoint
}

func Make() *Packet {
	return &Packet{}
}

func New(command int32, seq uint16, flag fatchoy.PacketFlag, body interface{}) *Packet {
	return &Packet{
		Type_: fatchoy.PTypePacket,
		Cmd:   command,
		Flg:   flag,
		Seq_:  seq,
		Body:  body,
	}
}

func (m *Packet) Command() int32 {
	return m.Cmd
}

func (m *Packet) SetCommand(v int32) {
	m.Cmd = v
}

func (m *Packet) Seq() uint16 {
	return m.Seq_
}

func (m *Packet) SetSeq(v uint16) {
	m.Seq_ = v
}

func (m *Packet) Type() fatchoy.PacketType {
	return m.Type_
}

func (m *Packet) SetType(v fatchoy.PacketType) {
	m.Type_ = v
}

func (m *Packet) Flag() fatchoy.PacketFlag {
	return m.Flg
}

func (m *Packet) SetFlag(v fatchoy.PacketFlag) {
	m.Flg = v
}

func (m *Packet) Node() fatchoy.NodeID {
	return m.Node_
}

func (m *Packet) SetNode(n fatchoy.NodeID) {
	m.Node_ = n
}

func (m *Packet) Refers() []fatchoy.NodeID {
	return m.Refer
}

func (m *Packet) SetRefers(v []fatchoy.NodeID) {
	m.Refer = v
}

func (m *Packet) Endpoint() fatchoy.MessageEndpoint {
	return m.endpoint
}

func (m *Packet) SetEndpoint(endpoint fatchoy.MessageEndpoint) {
	m.endpoint = endpoint
}

func (m *Packet) Reset() {
	m.Cmd = 0
	m.Seq_ = 0
	m.Flg = 0
	m.Type_ = 0
	m.Node_ = 0
	m.Refer = nil
	m.Body = nil
	m.endpoint = nil
}

func (m *Packet) Clone() fatchoy.IPacket {
	return &Packet{
		Cmd:      m.Cmd,
		Flg:      m.Flg,
		Type_:    m.Type_,
		Seq_:     m.Seq_,
		Node_:    m.Node_,
		Refer:    m.Refer,
		Body:     m.Body,
		endpoint: m.endpoint,
	}
}

func (m *Packet) Errno() int32 {
	if (m.Flg & fatchoy.PFlagError) != 0 {
		return m.Cmd
	}
	return 0
}

// 如果消息表示一个错误码，设置PacketFlagError标记，并且body为错误码数值
func (m *Packet) SetErrno(ec int32) {
	m.Flg |= fatchoy.PFlagError
	m.SetBody(int64(ec))
}

// body的类型仅支持int64/float64/string/bytes/proto.Message
func (m *Packet) Reply(command int32, body interface{}) error {
	var pkt = New(command, m.Seq_, m.Flg, body)
	pkt.Node_ = m.Node_
	pkt.Refer = m.Refer
	return m.endpoint.SendPacket(pkt)
}

// 响应proto消息内容
func (m *Packet) ReplyMsg(ack proto.Message) error {
	var mid = GetMessageIDOf(ack)
	if mid == 0 {
		mid = m.Cmd
	}
	return m.Reply(mid, ack)
}

// 返回一个错误码消息
func (m *Packet) Refuse(errno int32) error {
	var ackMsgId = GetPairingAckID(m.Cmd)
	if ackMsgId == 0 {
		ackMsgId = m.Cmd
	}
	return m.RefuseWith(ackMsgId, errno)
}

func (m *Packet) RefuseWith(command, errno int32) error {
	var pkt = New(command, m.Seq_, m.Flg|fatchoy.PFlagError, nil)
	pkt.Node_ = m.Node_
	pkt.Refer = m.Refer
	pkt.SetErrno(errno)
	return m.endpoint.SendPacket(pkt)
}

func (m Packet) String() string {
	var nodeID fatchoy.NodeID
	if m.endpoint != nil {
		nodeID = m.endpoint.NodeID()
	}
	return fmt.Sprintf("%v c:%d seq:%d 0x%x", nodeID, m.Cmd, m.Seq_, m.Flg)
}
