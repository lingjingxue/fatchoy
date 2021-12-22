// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"encoding/binary"
	"hash/crc32"

	"qchen.fun/fatchoy"
)

const (
	V2EncoderVersion  = 1
	V2HeaderSize      = 24            // 包头大小
	V2MaxPayloadBytes = (1 << 24) - 1 // 16M
	V2MaxReferCount   = 1000
)

// V2协议主要用于server内部的通信，一条消息大致分为下面这3段：
//  `header + refer(变长） + body`
// refer指的是节点的session号，个数由header中的`nref`指定，主要用于协议转发和广播等等
//
// V2协议头，len包含header和body
//       ----------------------------------------------
// field | len | flag | #ref | seq | node | cmd | crc |
//       ----------------------------------------------
// bytes |  4  |  2   |   2  |  4  |  4   |  4  |  4  |

type V2Header []byte

// 长度包含头部和body
func (h V2Header) Len() uint32 {
	return binary.BigEndian.Uint32(h[:4])
}

// 类型
func (h V2Header) Type() uint8 {
	return h[4]
}

// 标记位
func (h V2Header) Flag() uint8 {
	return h[5]
}

// 标记位
func (h V2Header) RefCount() uint16 {
	return binary.BigEndian.Uint16(h[6:])
}

// session内的唯一序号
func (h V2Header) Seq() uint32 {
	return binary.BigEndian.Uint32(h[8:])
}

// 目标节点
func (h V2Header) Node() fatchoy.NodeID {
	return fatchoy.NodeID(binary.BigEndian.Uint32(h[12:]))
}

func (h V2Header) Command() int32 {
	return int32(binary.BigEndian.Uint32(h[16:]))
}

// CRC校验码
func (h V2Header) Checksum() uint32 {
	return binary.BigEndian.Uint32(h[V2HeaderSize-4:])
}

// 校验码包含head和body
func (h V2Header) CalcChecksum(refer, payload []byte) uint32 {
	var hasher = crc32.NewIEEE()
	hasher.Write(h[:V2HeaderSize-4])
	if len(refer) > 0 {
		hasher.Write(refer)
	}
	if len(payload) > 0 {
		hasher.Write(payload)
	}
	return hasher.Sum32()
}

func (h V2Header) SetChecksum(crc uint32) {
	binary.BigEndian.PutUint32(h[V2HeaderSize-4:], crc)
}

func (h V2Header) Pack(pkt fatchoy.IPacket, refcnt uint16, size uint32) {
	binary.BigEndian.PutUint32(h[:4], size)
	binary.BigEndian.PutUint16(h[4:], uint16(pkt.Flags()))
	binary.BigEndian.PutUint16(h[6:], refcnt)
	binary.BigEndian.PutUint32(h[8:], pkt.Seq())
	binary.BigEndian.PutUint32(h[12:], uint32(pkt.Node()))
	binary.BigEndian.PutUint32(h[16:], uint32(pkt.Command()))
}

func (h V2Header) MD5Sum() string {
	return md5Sum(h[:])
}
