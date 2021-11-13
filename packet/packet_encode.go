// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package packet

import (
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/golang/protobuf/proto"

	"gopkg.in/qchencc/fatchoy.v1"
)

// 如果消息表示一个错误码，设置PacketFlagError标记，并且body为错误码数值
func (m *Packet) SetErrno(ec int32) {
	m.Flg |= fatchoy.PFlagError
	m.SetBodyInt(int64(ec))
}

// 消息体是number
func (m *Packet) SetBodyInt(n int64) {
	m.Body = n
}

func (m *Packet) BodyToInt() int64 {
	switch v := m.Body.(type) {
	case int64:
		return v
	case string:
		n, _ := strconv.ParseInt(v, 10, 64)
		return n
	case []byte:
		s := string(v)
		n, _ := strconv.ParseInt(s, 10, 64)
		return n
	default:
		panic(fmt.Sprintf("cannot convert %T to number", v))
	}
	return 0
}

// 消息体是string
func (m *Packet) SetBodyString(s string) {
	m.Body = s
}

func (m *Packet) BodyToString() string {
	switch v := m.Body.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case int64:
		return strconv.FormatInt(v, 64)
	case proto.Message:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// 消息体是[]byte
func (m *Packet) SetBodyBytes(b []byte) {
	m.Body = b
}

func (m *Packet) BodyToBytes() []byte {
	switch v := m.Body.(type) {
	case string:
		return []byte(v)
	case []byte:
		return v
	case int64:
		var tmp [binary.MaxVarintLen64]byte
		var n = binary.PutVarint(tmp[:], v)
		return tmp[:n]
	case proto.Message:
		if data, err := proto.Marshal(v); err != nil {
			panic(err)
		} else {
			return data
		}
	default:
		panic(fmt.Sprintf("cannot convert %T to bytes", v))
	}
	return nil
}

// 消息体是pbapi.Packet
func (m *Packet) SetBodyMsg(msg proto.Message) {
	m.Body = msg
}

func (m *Packet) DecodeTo(msg proto.Message) error {
	return proto.Unmarshal(m.Body.([]byte), msg)
}

// 自动解析
func (m *Packet) Decode() error {
	var msg = CreateMessageByID(m.Cmd)
	if msg == nil {
		return fmt.Errorf("cannot create message of ID %d", m.Cmd)
	}
	var data = m.BodyToBytes()
	if err := proto.Unmarshal(data, msg); err != nil {
		return fmt.Errorf("cannot unmarshal message %d: %w", m.Cmd, err)
	}
	m.Body = msg
	return nil
}
