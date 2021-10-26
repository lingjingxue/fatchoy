// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"io"

	"gopkg.in/qchencc/fatchoy.v1"
	"gopkg.in/qchencc/fatchoy.v1/x/cipher"
)

type V2Codec struct {
	maxUpstreamPacketBytes   int // 上行包最大大小
	maxDownstreamPacketBytes int // 下行包最大大小
}

var V2 = NewV2()

func NewV2() *V2Codec {
	var upLimit = GetEnvInt("V2_UP_PKT_BYTES", PacketBytesLimit)
	var downLimit = GetEnvInt("V2_DOWN_PKT_BYTES", PacketBytesLimit)
	if upLimit <= 0 || upLimit > PacketBytesLimit {
		upLimit = PacketBytesLimit
	}
	if downLimit <= 0 || downLimit > PacketBytesLimit {
		downLimit = PacketBytesLimit
	}
	return &V2Codec{
		maxUpstreamPacketBytes:   upLimit,
		maxDownstreamPacketBytes: downLimit,
	}
}

func (c *V2Codec) Marshal(w io.Writer, pkt fatchoy.IPacket, encryptor cipher.BlockCryptor) (int, error) {
	return marshalPacket(VersionV2, w, pkt, encryptor, c.maxDownstreamPacketBytes)
}

func (c *V2Codec) Unmarshal(r io.Reader, head *Header, pkt fatchoy.IPacket, decrypt cipher.BlockCryptor) (int, error) {
	return unmarshalPacket(r, head, pkt, decrypt, c.maxUpstreamPacketBytes)
}
