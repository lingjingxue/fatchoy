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

type Version uint8

const (
	VersionV1 Version = 1
)

const (
	HeaderSize       = 16               // 包头大小(包含长度）
	PacketBytesLimit = (1 << 24) - 1    // 3字节限制
	MaxPayloadBytes  = 16 * 1024 * 1024 // 16M
)

//  协议头，little endian表示，len包含header和body
//       ---------------------------------------
// field | len | flag | type | seq | cmd | crc |
//       ---------------------------------------
// bytes |  4  |   1  |   1  |  2  |  4  |  4  |

// 不包含length
type Header []byte

func NewHeader() Header {
	return make([]byte, HeaderSize-4)
}

// 标记位
func (h Header) Flag() uint8 {
	return h[0]
}

// 标记位
func (h Header) Type() uint8 {
	return h[1]
}

// session内的唯一序号
func (h Header) Seq() int16 {
	return int16(binary.LittleEndian.Uint16(h[2:]))
}

func (h Header) Command() int32 {
	return int32(binary.LittleEndian.Uint32(h[4:]))
}

// CRC校验码
func (h Header) Checksum() uint32 {
	return binary.LittleEndian.Uint32(h[8:])
}

// 校验码包含head和body
func (h Header) CalcChecksum(payload []byte) uint32 {
	var hasher = crc32.NewIEEE()
	hasher.Write(h[:8])
	if len(payload) > 0 {
		hasher.Write(payload)
	}
	return hasher.Sum32()
}

func (h Header) SetChecksum(crc uint32) {
	binary.LittleEndian.PutUint32(h[8:], crc)
}

func (h Header) Pack(pkt fatchoy.IPacket) {
	h[0] = byte(pkt.Flag())
	h[1] = byte(pkt.Type())
	binary.LittleEndian.PutUint16(h[2:], uint16(pkt.Seq()))
	binary.LittleEndian.PutUint32(h[4:], uint32(pkt.Command()))
}

func (h Header) MD5Sum() string {
	return Md5Sum(h[:])
}

func Md5Sum(data []byte) string {
	var hash = md5.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}
