// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"

	"github.com/golang/protobuf/proto"
	"gopkg.in/qchencc/fatchoy.v1"
)

type testPacket struct {
	command  int32
	seq      uint16
	flag     fatchoy.PacketFlag
	typ      fatchoy.PacketType
	node     fatchoy.NodeID
	body     []byte
	refer    []fatchoy.NodeID
	endpoint fatchoy.MessageEndpoint
}

func (m *testPacket) Command() int32 {
	return m.command
}

func (m *testPacket) SetCommand(v int32) {
	m.command = v
}

func (m *testPacket) Seq() uint16 {
	return m.seq
}

func (m *testPacket) SetSeq(v uint16) {
	m.seq = v
}

func (m *testPacket) Type() fatchoy.PacketType {
	return m.typ
}

func (m *testPacket) SetType(v fatchoy.PacketType) {
	m.typ = v
}

func (m *testPacket) Flag() fatchoy.PacketFlag {
	return m.flag
}

func (m *testPacket) SetFlag(v fatchoy.PacketFlag) {
	m.flag = v
}

func (m *testPacket) Body() interface{} {
	return m.body
}

func (m *testPacket) Errno() int32 {
	if (m.flag & fatchoy.PFlagError) != 0 {
		return m.command
	}
	return 0
}

func (m *testPacket) SetErrno(ec int32) {
	m.flag |= fatchoy.PFlagError
	m.SetBody(int64(ec))
}

func (m *testPacket) Node() fatchoy.NodeID {
	return m.node
}

func (m *testPacket) SetNode(n fatchoy.NodeID) {
	m.node = n
}

func (m *testPacket) Refers() []fatchoy.NodeID {
	return m.refer
}

func (m *testPacket) SetRefers(v []fatchoy.NodeID) {
	m.refer = v
}

func (m *testPacket) AddRefers(v ...fatchoy.NodeID) {
	m.refer = append(m.refer, v...)
}

func (m *testPacket) Endpoint() fatchoy.MessageEndpoint {
	return m.endpoint
}

func (m *testPacket) SetEndpoint(v fatchoy.MessageEndpoint) {
	m.endpoint = v
}

func (m *testPacket) Clone() fatchoy.IPacket {
	return &testPacket{
		command:  m.command,
		seq:      m.seq,
		flag:     m.flag,
		typ:      m.typ,
		body:     m.body,
		refer:    m.refer,
		endpoint: m.endpoint,
	}
}

func (m *testPacket) SetBody(val interface{}) {
	switch val.(type) {
	case int, int8, int16, int32, int64:
	case uint, uint8, uint16, uint32, uint64:
	case float32, float64:
	case bool, nil, string, []byte, proto.Message:
	default:
		panic(fmt.Sprintf("cannot set body as %T", val))
	}
	if val != nil {
		var buf bytes.Buffer
		var enc = gob.NewEncoder(&buf)
		if err := enc.Encode(val); err != nil {
			panic(err)
		}
		m.body = buf.Bytes()
	} else {
		m.body = nil
	}
}

func (m *testPacket) BodyToInt() int64 {
	return int64(binary.LittleEndian.Uint64(m.body))
}

func (m *testPacket) BodyToFloat() float64 {
	return 0
}

func (m *testPacket) BodyToString() string {
	return string(m.body)
}

func (m *testPacket) BodyToBytes() []byte {
	return m.body
}

func (m *testPacket) DecodeTo(msg proto.Message) error {
	return proto.Unmarshal(m.body, msg)
}

func (m *testPacket) Decode() error {
	return nil
}

func (m testPacket) String() string {
	var checksum = md5Sum(m.body)
	return fmt.Sprintf("c:%d seq:%d 0x%x %s", m.Command(), m.Seq(), m.Flag(), checksum)
}

func (m *testPacket) Reply(command int32, s interface{}) error {
	panic("not implemented")
	return nil
}

func (m *testPacket) ReplyMsg(ack proto.Message) error {
	panic("not implemented")
	return nil
}

func (m *testPacket) Refuse(errno int32) error {
	panic("not implemented")
	return nil
}

func (m *testPacket) RefuseWith(command int32, errno int32) error {
	panic("not implemented")
	return nil
}
