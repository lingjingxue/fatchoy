// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package packet

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"gopkg.in/qchencc/fatchoy"
)

// Packet表示一个应用层消息
type Packet struct {
	command  int32                   // 协议ID
	seq      int16                   // 序列号
	type_    fatchoy.PacketType      // 类型
	flag     fatchoy.PacketFlag      // 标志位
	body     interface{}             // 消息内容，number/string/bytes/pb.Packet
	endpoint fatchoy.MessageEndpoint // 关联的endpoint
}

func Make() *Packet {
	return &Packet{}
}

func New(command int32, seq int16, flag fatchoy.PacketFlag, body interface{}) *Packet {
	return &Packet{
		command: command,
		flag:    flag,
		seq:     seq,
		body:    body,
	}
}

func (m *Packet) Command() int32 {
	return m.command
}

func (m *Packet) SetCommand(v int32) {
	m.command = v
}

func (m *Packet) Seq() int16 {
	return m.seq
}

func (m *Packet) SetSeq(v int16) {
	m.seq = v
}

func (m *Packet) Type() fatchoy.PacketType {
	return m.type_
}

func (m *Packet) SetType(v fatchoy.PacketType) {
	m.type_ = v
}

func (m *Packet) Flag() fatchoy.PacketFlag {
	return m.flag
}

func (m *Packet) SetFlag(v fatchoy.PacketFlag) {
	m.flag = v
}

func (m *Packet) Body() interface{} {
	return m.body
}

func (m *Packet) SetBody(v interface{}) {
	m.body = v
}

func (m *Packet) Endpoint() fatchoy.MessageEndpoint {
	return m.endpoint
}

func (m *Packet) SetEndpoint(endpoint fatchoy.MessageEndpoint) {
	m.endpoint = endpoint
}

func (m *Packet) Reset() {
	m.command = 0
	m.seq = 0
	m.flag = 0
	m.body = nil
	m.endpoint = nil
}

func (m *Packet) Clone() Packet {
	return Packet{
		command:  m.command,
		flag:     m.flag,
		seq:      m.seq,
		body:     m.body,
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
	if (m.flag & fatchoy.PacketFlagError) != 0 {
		return m.command
	}
	return 0
}

// 返回响应
func (m *Packet) ReplyCommand(command int32, ack proto.Message) error {
	var pkt = New(command, m.seq, m.flag, ack)
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
	var pkt = New(command, m.seq, m.flag, s)
	return m.endpoint.SendPacket(pkt)
}

// 响应字节内容
func (m *Packet) ReplyBytes(command int32, b []byte) error {
	var pkt = New(command, m.seq, m.flag, b)
	return m.endpoint.SendPacket(pkt)
}

// 返回一个错误码消息
func (m *Packet) Refuse(errno int32) error {
	var ackMsgId = GetPairingAckID(m.command)
	return m.RefuseCommand(ackMsgId, errno)
}

func (m *Packet) RefuseCommand(command, errno int32) error {
	var pkt = New(command, m.seq, m.flag|fatchoy.PacketFlagError, nil)
	pkt.SetErrno(errno)
	return m.endpoint.SendPacket(pkt)
}

func (m Packet) String() string {
	var nodeID fatchoy.NodeID
	if m.endpoint != nil {
		nodeID = m.endpoint.NodeID()
	}
	return fmt.Sprintf("%v c:%d seq:%d 0x%x", nodeID, m.Command(), m.Seq(), m.Flag())
}
