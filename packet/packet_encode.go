// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package packet

import (
	"encoding/binary"

	"github.com/golang/protobuf/proto"
	"gopkg.in/qchencc/fatchoy.v1"
	"gopkg.in/qchencc/fatchoy.v1/codes"
	"gopkg.in/qchencc/fatchoy.v1/log"
	"gopkg.in/qchencc/fatchoy.v1/x/fsutil"
)

// 如果消息表示一个错误码，设置PacketFlagError标记，并且body为错误码数值
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

// 根据pkt的Flag标志位，对body进行压缩
func Encode(pkt fatchoy.IPacket, threshold int) error {
	payload, err := pkt.EncodeBodyToBytes()
	if err != nil {
		return err
	}
	if payload == nil {
		return nil
	}
	if n := len(payload); threshold > 0 && n > threshold {
		if data, err := fsutil.CompressBytes(payload); err != nil {
			log.Errorf("compress packet %v with %d bytes: %v", pkt.Command(), n, err)
			return err
		} else {
			payload = data
			pkt.SetFlag(pkt.Flag() | fatchoy.PacketFlagCompressed)
		}
	}
	pkt.SetBodyBytes(payload)
	return nil
}

// 根据pkt的Flag标志位，对body进行解压缩
func Decode(pkt fatchoy.IPacket) error {
	payload := pkt.BodyAsBytes()
	if payload == nil {
		return nil
	}
	var flag = pkt.Flag()
	if (flag & fatchoy.PacketFlagCompressed) > 0 {
		if uncompressed, err := fsutil.UncompressBytes(payload); err != nil {
			log.Errorf("decompress packet %v(%d bytes): %v", pkt.Command(), len(payload), err)
			return err
		} else {
			payload = uncompressed
			pkt.SetFlag(flag &^ fatchoy.PacketFlagCompressed)
		}
	}
	// 如果有FlagError，则body是数值错误码
	if (flag & fatchoy.PacketFlagError) != 0 {
		val, n := binary.Varint(payload)
		if n > 0 {
			pkt.SetBodyNumber(val)
		} else {
			pkt.SetBodyNumber(int64(codes.TransportFailure))
		}
	} else {
		pkt.SetBodyBytes(payload)
	}
	return nil
}
