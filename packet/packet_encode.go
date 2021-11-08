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

func (m *Packet) BodyToNumber() int64 {
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

// 根据pkt的Flag标志位，对body进行压缩
func Encode(pkt fatchoy.IPacket, threshold int) error {
	payload := pkt.BodyToBytes()
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
	payload := pkt.BodyToBytes()
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
