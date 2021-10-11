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
	Body     interface{}             `json:"body,omitempty"` // 消息内容，number/string/bytes/pb.Packet
	endpoint fatchoy.MessageEndpoint // 关联的endpoint
}

func Make() *Packet {
	return &Packet{}
}

func New(command int32, seq int16, typ fatchoy.PacketType, flag fatchoy.PacketFlag, body interface{}) *Packet {
	return &Packet{
		Cmd:      command,
		Flg:      flag,
		Typ:      typ,
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

func (m *Packet) SetBody(v interface{}) {
	m.Body = v
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
	m.Body = nil
	m.endpoint = nil
}

func (m *Packet) Clone() Packet {
	return Packet{
		Cmd:      m.Cmd,
		Flg:      m.Flg,
		Typ:      m.Typ,
		Sequence: m.Sequence,
		Body:     m.Body,
		endpoint: m.endpoint,
	}
}

func (m *Packet) CloneBody() ([]byte, error) {
	data, err := m.EncodeBodyToBytes()
	if err != nil {
		return nil, err
	}
	var clone = make([]byte, len(data))
	copy(clone, data)
	return clone, nil
}

func (m *Packet) Errno() int32 {
	if (m.Flg & fatchoy.PacketFlagError) != 0 {
		return m.Cmd
	}
	return 0
}

// 返回响应
func (m *Packet) ReplyCommand(command int32, ack proto.Message) error {
	var pkt = New(command, m.Sequence, m.Typ, m.Flg, ack)
	return m.endpoint.SendPacket(pkt)
}

// 响应proto消息内容
func (m *Packet) Reply(ack proto.Message) error {
	var mid = GetMessageIDOf(ack)
	if mid == 0 {
		return fmt.Errorf("message ID of %T not found", ack)
	}
	return m.ReplyCommand(mid, ack)
}

// 响应string内容
func (m *Packet) ReplyString(command int32, s string) error {
	var pkt = New(command, m.Sequence, m.Typ, m.Flg, s)
	return m.endpoint.SendPacket(pkt)
}

// 响应字节内容
func (m *Packet) ReplyBytes(command int32, b []byte) error {
	var pkt = New(command, m.Sequence, m.Typ, m.Flg, b)
	return m.endpoint.SendPacket(pkt)
}

// 返回一个错误码消息
func (m *Packet) Refuse(errno int32) error {
	var ackMsgId = GetPairingAckID(m.Cmd)
	return m.RefuseCommand(ackMsgId, errno)
}

func (m *Packet) RefuseCommand(command, errno int32) error {
	var pkt = New(command, m.Sequence, m.Typ, m.Flg|fatchoy.PacketFlagError, nil)
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
