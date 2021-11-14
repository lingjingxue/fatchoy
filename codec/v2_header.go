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
	VersionV2    = 2
	V2HeaderSize = 16 // 包头大小(包含长度）
)

//  协议头，len包含header和body
//       -----------------------------------------------
// field | len | type | flag | refer | seq | cmd | crc |
//       -----------------------------------------------
// bytes |  2  |   1  |   1  |   1   |  2  |  4  |  4  |

type V2Header []byte

func (h V2Header) Len() uint16 {
	return binary.BigEndian.Uint16(h)
}

// 类型
func (h V2Header) Type() uint8 {
	return h[2]
}

// 标记位
func (h V2Header) Flag() uint8 {
	return h[3]
}

// 标记位
func (h V2Header) RefCount() uint16 {
	return binary.BigEndian.Uint16(h[4:])
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

func (h V2Header) Pack(pkt fatchoy.IPacket, refcnt, size uint16) {
	binary.BigEndian.PutUint16(h[:], size)
	h[2] = byte(pkt.Type())
	h[3] = byte(pkt.Flag())
	binary.BigEndian.PutUint16(h[4:], refcnt)
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
