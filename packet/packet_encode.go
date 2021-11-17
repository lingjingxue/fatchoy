// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package packet

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"gopkg.in/qchencc/fatchoy.v1/qlog"
)

func (m *Packet) Body() interface{} {
	return m.Body_
}

func (m *Packet) SetBody(val interface{}) {
	switch v := val.(type) {
	case int:
		m.Body_ = int64(v)
	case uint:
		m.Body_ = int64(v)
	case int8:
		m.Body_ = int64(v)
	case int16:
		m.Body_ = int64(v)
	case int32:
		m.Body_ = int64(v)
	case uint8:
		m.Body_ = int64(v)
	case uint16:
		m.Body_ = int64(v)
	case uint32:
		m.Body_ = int64(v)
	case uint64:
		m.Body_ = int64(v)
	case float32:
		m.Body_ = float64(v)
	case nil:
		m.Body_ = nil
	case bool:
		if v {
			m.Body_ = int64(1)
		} else {
			m.Body_ = int64(0)
		}
	case int64, float64, string, []byte, proto.Message:
		m.Body_ = val
	default:
		panic(fmt.Sprintf("cannot set body as %T", val))
	}
}

// 将body转为int64
func (m *Packet) BodyToInt() int64 {
	switch v := m.Body_.(type) {
	case int64:
		return v
	case float64:
		return int64(v)
	case string:
		if n, err := strconv.ParseInt(v, 10, 64); err != nil {
			panic(fmt.Sprintf("cannot convert packet %d body to int: %v", m.Cmd, err))
		} else {
			return n
		}
	case []byte:
		switch len(v) {
		case 0:
			return 0
		case 1:
			return int64(v[0])
		case 2:
			return int64(binary.LittleEndian.Uint16(v))
		case 4:
			return int64(binary.LittleEndian.Uint32(v))
		case 8:
			return int64(binary.LittleEndian.Uint64(v))
		default:
			panic(fmt.Sprintf("cannot convert %d bytes to integer", len(v)))
		}
	default:
		panic(fmt.Sprintf("cannot convert %T to integer", v))
	}
	return 0
}

// 将body转为float4
func (m *Packet) BodyToFloat() float64 {
	switch v := m.Body_.(type) {
	case int64:
		return float64(v)
	case float64:
		return v
	case string:
		if f, err := strconv.ParseFloat(v, 64); err != nil {
			panic(fmt.Sprintf("cannot convert packet %d body to float: %v", m.Cmd, err))
		} else {
			return f
		}
	case []byte:
		switch len(v) {
		case 4:
			b := binary.LittleEndian.Uint32(v)
			return float64(math.Float32frombits(b))
		case 8:
			b := binary.LittleEndian.Uint64(v)
			return math.Float64frombits(b)
		default:
			panic(fmt.Sprintf("cannot convert %d bytes to float", len(v)))
		}
	default:
		panic(fmt.Sprintf("cannot convert %T to float", v))
	}
}

// 将body转为string
func (m *Packet) BodyToString() string {
	switch v := m.Body_.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case int64:
		return strconv.FormatInt(v, 64)
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64)
	case proto.Message:
		return MessageToString(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func encodeInt64(n int64) []byte {
	var tmp [binary.MaxVarintLen64]byte
	i := binary.PutVarint(tmp[:], n)
	return tmp[:i]
}

func encodeUint64(n uint64) []byte {
	var tmp [binary.MaxVarintLen64]byte
	i := binary.PutUvarint(tmp[:], n)
	return tmp[:i]
}

// 将body转为[]byte，用于网络传输
func (m *Packet) BodyToBytes() []byte {
	switch v := m.Body_.(type) {
	case string:
		return []byte(v)
	case []byte:
		return v
	case int64:
		return encodeInt64(v)
	case float64:
		return encodeUint64(math.Float64bits(v))
	case proto.Message:
		if data, err := proto.Marshal(v); err != nil {
			panic(fmt.Sprintf("cannot marshal packet %d body: %v", m.Cmd, err))
		} else {
			return data
		}
	default:
		panic(fmt.Sprintf("cannot convert %T to bytes", v))
	}
	return nil
}

func (m *Packet) DecodeTo(msg proto.Message) error {
	var data = m.BodyToBytes()
	return proto.Unmarshal(data, msg)
}

// 自动解析
func (m *Packet) Decode() error {
	var name = GetMessageNameByID(m.Cmd)
	if name == "" {
		return fmt.Errorf("cannot create message of %d", m.Cmd)
	}
	var msg = CreateMessageByName(name)
	if msg == nil {
		return fmt.Errorf("cannot create message %s", name)
	}
	var data = m.BodyToBytes()
	if err := proto.Unmarshal(data, msg); err != nil {
		return fmt.Errorf("cannot unmarshal message %d: %w", m.Cmd, err)
	}
	m.Body_ = msg
	return nil
}

func MessageToString(msg proto.Message) string {
	var m jsonpb.Marshaler
	if s, err := m.MarshalToString(msg); err != nil {
		qlog.Errorf("marshal %T: %v", msg, err)
	} else {
		return s
	}
	return msg.String()
}
