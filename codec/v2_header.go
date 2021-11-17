// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"hash/crc32"

	"gopkg.in/qchencc/fatchoy.v1"
)

const (
	VersionV2         = 2
	V2HeaderSize      = 20              // 包头大小
	V2MaxPayloadBytes = 8 * 1024 * 1024 // 8M
)

// V2协议主要用于server内部的通信，一条消息大致分为下面这3段：
//  `header + refer(变长） + body`
// refer指的是节点的session号，个数由header中的`nref`指定，主要用于协议转发和广播等等
//
// V2协议头，len包含header和body
//       -----------------------------------------------------
// field | len | type | flag | #ref | seq | node | cmd | crc |
//       -----------------------------------------------------
// bytes |  3  |   1  |   1  |   1  |  2  |   4  |   4  |  4 |

type V2Header []byte

// 长度包含头部和body
func (h V2Header) Len() uint32 {
	return uint32(h[2]) | uint32(h[1])<<8 | uint32(h[0])<<16 // big endian
}

// 类型
func (h V2Header) Type() uint8 {
	return h[3]
}

// 标记位
func (h V2Header) Flag() uint8 {
	return h[4]
}

// 标记位
func (h V2Header) RefCount() uint8 {
	return h[5]
}

// session内的唯一序号
func (h V2Header) Seq() uint16 {
	return binary.BigEndian.Uint16(h[6:])
}

func (h V2Header) Node() fatchoy.NodeID {
	return fatchoy.NodeID(binary.BigEndian.Uint32(h[8:]))
}

func (h V2Header) Command() int32 {
	return int32(binary.BigEndian.Uint32(h[12:]))
}

// CRC校验码
func (h V2Header) Checksum() uint32 {
	return binary.BigEndian.Uint32(h[16:])
}

// 校验码包含head和body
func (h V2Header) CalcChecksum(payload []byte) uint32 {
	var hasher = crc32.NewIEEE()
	hasher.Write(h[:V2HeaderSize-4])
	if len(payload) > 0 {
		hasher.Write(payload)
	}
	return hasher.Sum32()
}

func (h V2Header) SetChecksum(crc uint32) {
	binary.BigEndian.PutUint32(h[V2HeaderSize-4:], crc)
}

func (h V2Header) Pack(pkt fatchoy.IPacket, nRef uint8, size uint32) {
	// big endian
	h[0] = byte(size >> 16)
	h[1] = byte(size >> 8)
	h[2] = byte(size)

	h[3] = byte(pkt.Type())
	h[4] = byte(pkt.Flag())
	h[5] = nRef
	binary.BigEndian.PutUint16(h[6:], pkt.Seq())
	binary.BigEndian.PutUint32(h[8:], uint32(pkt.Node()))
	binary.BigEndian.PutUint32(h[12:], uint32(pkt.Command()))
}

func (h V2Header) MD5Sum() string {
	return md5Sum(h[:])
}

func md5Sum(data []byte) string {
	var hash = md5.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}
