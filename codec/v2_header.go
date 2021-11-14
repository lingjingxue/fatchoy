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
	VersionV2        = 2
	V2HeaderSize     = 16               // 包头大小(包含长度）
	PacketBytesLimit = (1 << 24) - 1    // 3字节限制
	MaxPayloadBytes  = 16 * 1024 * 1024 // 16M
)

//  协议头，len包含header和body
//       ------------------------------------------------
// field | len | flag | type | refcnt | seq | cmd | crc |
//       ------------------------------------------------
// bytes |  3  |   1  |   1  |   1    |  2  |  4  |  4  |

type V2Header []byte

func (h V2Header) Len() uint32 {
	return uint32(h[2]) | uint32(h[1])<<8 | uint32(h[0])<<16 // big endian
}

// 标记位
func (h V2Header) Flag() uint8 {
	return h[3]
}

// 类型
func (h V2Header) Type() uint8 {
	return h[4]
}

// 标记位
func (h V2Header) RefCount() uint8 {
	return h[5]
}

// session内的唯一序号
func (h V2Header) Seq() int16 {
	return int16(binary.BigEndian.Uint16(h[6:]))
}

func (h V2Header) Command() int32 {
	return int32(binary.BigEndian.Uint32(h[8:]))
}

// CRC校验码
func (h V2Header) Checksum() uint32 {
	return binary.BigEndian.Uint32(h[12:])
}

// 校验码包含head和body
func (h V2Header) CalcChecksum(payload []byte) uint32 {
	var hasher = crc32.NewIEEE()
	hasher.Write(h[:12])
	if len(payload) > 0 {
		hasher.Write(payload)
	}
	return hasher.Sum32()
}

func (h V2Header) SetChecksum(crc uint32) {
	binary.BigEndian.PutUint32(h[12:], crc)
}

func (h V2Header) Pack(pkt fatchoy.IPacket, refcnt uint8, size uint32) {
	h[0] = byte(size >> 16)
	h[1] = byte(size >> 8)
	h[2] = byte(size)
	h[3] = byte(pkt.Flag())
	h[4] = byte(pkt.Type())
	h[5] = refcnt
	binary.BigEndian.PutUint16(h[6:], uint16(pkt.Seq()))
	binary.BigEndian.PutUint32(h[8:], uint32(pkt.Command()))
}

func (h V2Header) MD5Sum() string {
	return md5Sum(h[:])
}

func md5Sum(data []byte) string {
	var hash = md5.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}
