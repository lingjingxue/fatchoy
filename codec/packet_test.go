// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"encoding/binary"

	"github.com/golang/protobuf/proto"
	"gopkg.in/qchencc/fatchoy"
)

type testPacket struct {
	command int32
	seq     int16
	flag    fatchoy.PacketFlag
	type_   fatchoy.PacketType
	body    []byte
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
	return m.type_
}

func (m *testPacket) SetType(v fatchoy.PacketType) {
	m.type_ = v
}

func (m *testPacket) Flag() fatchoy.PacketFlag {
	return m.flag
}

func (m *testPacket) SetFlag(v fatchoy.PacketFlag) {
	m.flag = v
}

func (m *testPacket) Errno() int32 {
	if (m.flag & fatchoy.PacketFlagError) != 0 {
		return m.command
	}
	return 0
}

func (m *testPacket) SetErrno(ec int32) {
	m.flag |= fatchoy.PacketFlagError
	m.SetBodyNumber(int64(ec))
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

func (m *testPacket) SetBodyNumber(n int64) {
	var buf = make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(n))
	m.body = buf
}

func (m *testPacket) BodyAsNumber() int64 {
	return int64(binary.LittleEndian.Uint64(m.body))
}

func (m *testPacket) SetBodyString(s string) {
	m.body = []byte(s)
}

func (m *testPacket) BodyAsString() string {
	return string(m.body)
}

func (m *testPacket) SetBodyBytes(b []byte) {
	m.body = b
}

func (m *testPacket) BodyAsBytes() []byte {
	return m.body
}

func (m *testPacket) SetBodyMsg(msg proto.Message) {
	data, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	m.body = data
}

func (m *testPacket) BodyAsMsg() proto.Message {
	return nil
}

func (m *testPacket) EncodeBodyToBytes() ([]byte, error) {
	return m.body, nil
}

func (m *testPacket) DecodeTo(msg proto.Message) error {
	return proto.Unmarshal(m.body, msg)
}
