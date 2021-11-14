// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"encoding/binary"
	"gopkg.in/qchencc/fatchoy.v1"
	"hash/crc32"
)

const (
	VersionV1    = 1
	V1HeaderSize = 14 // 包头大小(包含长度）
)

//  协议头，len包含header和body
//       ---------------------------------
// field | len | flag |  seq | cmd | crc |
//       ---------------------------------
// bytes |  3  |   1  |   2  |  4  |  4  |

type V1Header []byte

func (h V1Header) Len() uint32 {
	return uint32(h[2]) | uint32(h[1])<<8 | uint32(h[0])<<16 // big endian
}

// 标记位
func (h V1Header) Flag() uint8 {
	return h[3]
}

// session内的唯一序号
func (h V1Header) Seq() int16 {
	return int16(binary.BigEndian.Uint16(h[4:]))
}

func (h V1Header) Command() int32 {
	return int32(binary.BigEndian.Uint32(h[6:]))
}

// CRC校验码
func (h V1Header) Checksum() uint32 {
	return binary.BigEndian.Uint32(h[10:])
}

// 校验码包含head和body
func (h V1Header) CalcChecksum(payload []byte) uint32 {
	var hasher = crc32.NewIEEE()
	hasher.Write(h[:10])
	if len(payload) > 0 {
		hasher.Write(payload)
	}
	return hasher.Sum32()
}

func (h V1Header) SetChecksum(crc uint32) {
	binary.BigEndian.PutUint32(h[10:], crc)
}

func (h V1Header) Pack(pkt fatchoy.IPacket, size uint32) {
	h[0] = byte(size >> 16)
	h[1] = byte(size >> 8)
	h[2] = byte(size)
	h[3] = byte(pkt.Flag())
	binary.BigEndian.PutUint16(h[4:], uint16(pkt.Seq()))
	binary.BigEndian.PutUint32(h[6:], uint32(pkt.Command()))
}
