// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package packet

import (
	"encoding/binary"

	"github.com/golang/protobuf/proto"
	"gopkg.in/qchencc/fatchoy.v1"
	"gopkg.in/qchencc/fatchoy.v1/log"
)

func (m *Packet) SetErrno(ec int32) {
	m.Flg |= fatchoy.PacketFlagError
	m.SetBodyNumber(int64(ec))
}

// 消息体是number
func (m *Packet) SetBodyNumber(n int64) {
	m.Body = n
}

func (m *Packet) BodyAsNumber() int64 {
	if n, ok := m.Body.(int64); ok {
		return n
	}
	return 0
}

// 消息体是string
func (m *Packet) SetBodyString(s string) {
	m.Body = s
}

func (m *Packet) BodyAsString() string {
	if s, ok := m.Body.(string); ok {
		return s
	}
	return ""
}

// 消息体是[]byte
func (m *Packet) SetBodyBytes(b []byte) {
	m.Body = b
}

func (m *Packet) BodyAsBytes() []byte {
	if b, ok := m.Body.([]byte); ok {
		return b
	}
	return nil
}

// 消息体是pbapi.Packet
func (m *Packet) SetBodyMsg(msg proto.Message) {
	m.Body = msg
}

func (m *Packet) BodyAsMsg() proto.Message {
	if v, ok := m.Body.(proto.Message); ok {
		return v
	}
	return nil
}

func (m *Packet) DecodeTo(msg proto.Message) error {
	return proto.Unmarshal(m.BodyAsBytes(), msg)
}

// 编码body到字节流
func (m *Packet) EncodeBodyToBytes() ([]byte, error) {
	switch v := m.Body.(type) {
	case int64:
		var sbuf [binary.MaxVarintLen64]byte
		var n = binary.PutVarint(sbuf[:], v)
		return sbuf[:n], nil
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	case proto.Message:
		return proto.Marshal(v)
	case nil:
		return nil, nil
	default:
		log.Panicf("message %d unsupported body type %T", m.Cmd, m.Body)
	}
	return nil, nil
}
