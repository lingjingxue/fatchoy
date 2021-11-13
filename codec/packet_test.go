// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"encoding/binary"
	"fmt"

	"github.com/golang/protobuf/proto"
	"gopkg.in/qchencc/fatchoy.v1"
)

type testPacket struct {
	command  int32
	seq      int16
	flag     fatchoy.PacketFlag
	typ    fatchoy.PacketType
	body     []byte
	refer []uint32
	endpoint fatchoy.MessageEndpoint
}

func (m *testPacket) Command() int32 {
	return m.command
}

func (m *testPacket) SetCommand(v int32) {
	m.command = v
}

func (m *testPacket) Seq() int16 {
	return m.seq
}

func (m *testPacket) SetSeq(v int16) {
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

func (m *testPacket) IBody() interface{} {
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
	m.SetBodyInt(int64(ec))
}

func (m *testPacket) Refer() []uint32 {
	return m.refer
}

func (m *testPacket) SetRefer(v []uint32) {
	m.refer = v
}

func (m *testPacket) Endpoint() fatchoy.MessageEndpoint {
	return m.endpoint
}

func (m *testPacket) SetEndpoint(v fatchoy.MessageEndpoint) {
	m.endpoint = v
}

func (m *testPacket) Body() interface{} {
	return m.body
}

func (m *testPacket) SetBodyInt(n int64) {
	var buf = make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(n))
	m.body = buf
}

func (m *testPacket) BodyToInt() int64 {
	return int64(binary.LittleEndian.Uint64(m.body))
}

func (m *testPacket) SetBodyString(s string) {
	m.body = []byte(s)
}

func (m *testPacket) BodyToString() string {
	return string(m.body)
}

func (m *testPacket) SetBodyBytes(b []byte) {
	m.body = b
}

func (m *testPacket) BodyToBytes() []byte {
	return m.body
}

func (m *testPacket) SetBodyMsg(msg proto.Message) {
	data, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	m.body = data
}

func (m *testPacket) EncodeBodyToBytes() ([]byte, error) {
	return m.body, nil
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

func (m *testPacket) ReplyWith(command int32, ack proto.Message) error {
	panic("not implemented")
	return nil
}

func (m *testPacket) Reply(ack proto.Message) error {
	panic("not implemented")
	return nil
}

func (m *testPacket) ReplyString(command int32, s string) error {
	panic("not implemented")
	return nil
}

func (m *testPacket) ReplyBytes(command int32, b []byte) error {
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
