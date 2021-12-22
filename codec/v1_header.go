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
	V1EncoderVersion  = 1
	V1HeaderSize      = 16            // 包头大小
	V1MaxPayloadBytes = (1 << 16) - 1 // 64K
)

// V1协议主要用于client和gateway之间的通信，设计目标主要是简单、稳定、易实现
//
//  V1协议头，len包含header和body
//       ---------------------------------
// field | len | flag |  seq | cmd | crc |
//       ---------------------------------
// bytes |  2  |   2  |   4  |  4  |  4  |

type V1Header []byte

// 长度包含头部和body
func (h V1Header) Len() uint16 {
	return binary.BigEndian.Uint16(h)
}

// 标记位
func (h V1Header) Flag() uint16 {
	return binary.BigEndian.Uint16(h[2:])
}

// session内的唯一序号
func (h V1Header) Seq() uint32 {
	return binary.BigEndian.Uint32(h[4:])
}

func (h V1Header) Command() int32 {
	return int32(binary.BigEndian.Uint32(h[8:]))
}

// CRC校验码
func (h V1Header) Checksum() uint32 {
	return binary.BigEndian.Uint32(h[V1HeaderSize-4:])
}

// 校验码包含head和body
func (h V1Header) CalcChecksum(payload []byte) uint32 {
	var hasher = crc32.NewIEEE()
	hasher.Write(h[:V1HeaderSize-4])
	if len(payload) > 0 {
		hasher.Write(payload)
	}
	return hasher.Sum32()
}

func (h V1Header) SetChecksum(crc uint32) {
	binary.BigEndian.PutUint32(h[V1HeaderSize-4:], crc)
}

func (h V1Header) Pack(pkt fatchoy.IPacket, size uint16) {
	binary.BigEndian.PutUint16(h, size)
	binary.BigEndian.PutUint16(h[2:], uint16(pkt.Flags()))
	binary.BigEndian.PutUint32(h[4:], pkt.Seq())
	binary.BigEndian.PutUint32(h[8:], uint32(pkt.Command()))
}

func (h V1Header) MD5Sum() string {
	return md5Sum(h[:])
}
