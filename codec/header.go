// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"hash/crc32"

	"gopkg.in/qchencc/fatchoy"
)

const (
	VersionV1 = 1
	VersionV2 = 2

	HeaderSize        = 16      // 包头大小
	PayloadBytesLimit = 1 << 24 // 3字节限制
)

//  header wire format
//       ---------------------------------------------
// field | ver | len | flag | type | seq | cmd | crc |
//       ---------------------------------------------
// bytes |  1  |  3  |   1  |   1  |  2  |  4  |  4  |

// 协议头，little endian表示
type Header [HeaderSize]byte

func (h *Header) Version() uint8 {
	return h[0]
}

func (h *Header) SetVersion(v uint8) {
	h[0] = v
}

// 包体长度
func (h *Header) Len() int {
	var n = uint32(h[1]) | uint32(h[2])<<8 | uint32(h[3])<<16 // little endian
	return int(n)
}

// 标记位
func (h *Header) Flag() uint8 {
	return h[4]
}

// 标记位
func (h *Header) Type() uint8 {
	return h[5]
}

// session内的唯一序号
func (h *Header) Seq() int16 {
	return int16(binary.LittleEndian.Uint16(h[6:]))
}

func (h *Header) Command() int32 {
	return int32(binary.LittleEndian.Uint32(h[8:]))
}

// CRC校验码
func (h *Header) Checksum() uint32 {
	return binary.LittleEndian.Uint32(h[12:])
}

// 校验码包含head和body
func (h *Header) CalcChecksum(payload []byte) uint32 {
	var hasher = crc32.NewIEEE()
	hasher.Write(h[:12])
	if len(payload) > 0 {
		hasher.Write(payload)
	}
	return hasher.Sum32()
}

func (h *Header) SetupChecksum(payload []byte) {
	var crc = h.CalcChecksum(payload)
	binary.LittleEndian.PutUint32(h[12:], crc)
}

func (h *Header) unmarshalFrom(pkt fatchoy.IMessage, bodySize, ver int) {
	var n = uint32(bodySize)
	h[0] = byte(ver)
	h[1] = byte(n)
	h[2] = byte(n >> 8)
	h[3] = byte(n >> 16)
	h[4] = byte(pkt.Flag())
	h[5] = byte(pkt.Type() | fatchoy.PacketTypeBinary)
	binary.LittleEndian.PutUint16(h[6:], uint16(pkt.Seq()))
	binary.LittleEndian.PutUint32(h[8:], uint32(pkt.Command()))
}

func (h *Header) MD5Sum() string {
	return Md5Sum(h[:])
}

func Md5Sum(data []byte) string {
	var hash = md5.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}
