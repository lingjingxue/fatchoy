// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package packet

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"qchen.fun/fatchoy/log"
)

// 将body转为[]byte，用于网络传输
func BodyToBytes(body interface{}) []byte {
	switch v := body.(type) {
	case string:
		return []byte(v)
	case []byte:
		return v
	case int32:
		return encodeFixedInt32(v)
	case int64:
		return encodeFixedUint64(uint64(v))
	case float64:
		return encodeFixedUint64(math.Float64bits(v))
	case proto.Message:
		if data, err := proto.Marshal(v); err != nil {
			log.Panicf("cannot marshal %T body: %v", v, err)
		} else {
			return data
		}
	default:
		log.Panicf("cannot convert %T to bytes", v)
	}
	return nil
}

// 将body转为string
func BodyToString(body interface{}) string {
	switch v := body.(type) {
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

func MessageToString(msg proto.Message) string {
	if b, err := protojson.Marshal(msg); err != nil {
		log.Errorf("marshal %T: %v", msg, err)
	} else {
		return string(b)
	}
	return ""
}

func Conv2Body(val interface{}) interface{} {
	switch v := val.(type) {
	case int:
		return int64(v)
	case uint:
		return int64(v)
	case int8:
		return int32(v)
	case uint8:
		return int32(v)
	case int16:
		return int32(v)
	case uint16:
		return int32(v)
	case uint32:
		return int32(v)
	case uint64:
		return int64(v)
	case float32:
		return float64(v)
	case nil:
		return nil
	case bool:
		if v {
			return int32(1)
		} else {
			return int32(0)
		}
	case int32, int64, float64, string, []byte, proto.Message:
		return val
	default:
		panic(fmt.Sprintf("cannot set body as %T", val))
	}
}

// 将body转为int64
func BodyToInt(body interface{}) int64 {
	switch v := body.(type) {
	case int32:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	case string:
		if n, err := strconv.ParseInt(v, 10, 64); err != nil {
			log.Panicf("cannot convert string body %s to int: %v", v, err)
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
			log.Panicf("cannot convert %d bytes to integer", len(v))
		}
	default:
		log.Panicf("cannot convert %T to integer", v)
	}
	return 0
}

// 将body转为float4
func BodyToFloat(body interface{}) float64 {
	switch v := body.(type) {
	case int64:
		return float64(v)
	case float64:
		return v
	case string:
		if f, err := strconv.ParseFloat(v, 64); err != nil {
			log.Panicf("cannot convert string body %s to float: %v", v, err)
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
			log.Panicf("cannot convert %d bytes to float", len(v))
		}
	default:
		log.Panicf("cannot convert %T to float", v)
	}
	return 0
}

func encodeFixedInt32(n int32) []byte {
	var tmp [4]byte
	binary.LittleEndian.PutUint32(tmp[:], uint32(n))
	return tmp[:]
}

func encodeFixedUint64(n uint64) []byte {
	var tmp [8]byte
	binary.LittleEndian.PutUint64(tmp[:], n)
	return tmp[:]
}
