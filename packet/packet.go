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
	Sequence int16                   `json:"seq"`            // 序列号
	Typ      fatchoy.PacketType      `json:"typ,omitempty"`  // 类型
	Flg      fatchoy.PacketFlag      `json:"flg,omitempty"`  // 标志位
	Body     interface{}             `json:"body,omitempty"` // 消息内容，int64/float64/string/bytes/proto.Message
	Refer    []fatchoy.NodeID        `json:"ref,omitempty"`  // referenced node IDs
	endpoint fatchoy.MessageEndpoint // 关联的endpoint
}

func Make() *Packet {
	return &Packet{}
}

func New(command int32, seq int16, flag fatchoy.PacketFlag, body interface{}) *Packet {
	return &Packet{
		Typ:      fatchoy.PTypePacket,
		Cmd:      command,
		Flg:      flag,
		Sequence: seq,
		Body:     body,
	}
}

func (m *Packet) Command() int32 {
	return m.Cmd
}

func (m *Packet) SetCommand(v int32) {
	m.Cmd = v
}

func (m *Packet) Seq() int16 {
	return m.Sequence
}

func (m *Packet) SetSeq(v int16) {
	m.Sequence = v
}

func (m *Packet) Type() fatchoy.PacketType {
	return m.Typ
}

func (m *Packet) SetType(v fatchoy.PacketType) {
	m.Typ = v
}

func (m *Packet) Flag() fatchoy.PacketFlag {
	return m.Flg
}

func (m *Packet) SetFlag(v fatchoy.PacketFlag) {
	m.Flg = v
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
	m.Sequence = 0
	m.Flg = 0
	m.Typ = 0
	m.Refer = nil
	m.Body = nil
	m.endpoint = nil
}

func (m *Packet) Clone() fatchoy.IPacket {
	return &Packet{
		Cmd:      m.Cmd,
		Flg:      m.Flg,
		Typ:      m.Typ,
		Sequence: m.Sequence,
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
	var pkt = New(command, m.Sequence, m.Flg, body)
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

// 响应string内容
func (m *Packet) ReplyString(command int32, s string) error {
	var pkt = New(command, m.Sequence, m.Flg, s)
	return m.endpoint.SendPacket(pkt)
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
	var pkt = New(command, m.Sequence, m.Flg|fatchoy.PFlagError, nil)
	pkt.Refer = m.Refer
	pkt.SetErrno(errno)
	return m.endpoint.SendPacket(pkt)
}

func (m Packet) String() string {
	var nodeID fatchoy.NodeID
	if m.endpoint != nil {
		nodeID = m.endpoint.NodeID()
	}
	return fmt.Sprintf("%v c:%d seq:%d 0x%x", nodeID, m.Cmd, m.Sequence, m.Flg)
}
