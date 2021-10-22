// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"io"

	"gopkg.in/qchencc/fatchoy.v1"
	"gopkg.in/qchencc/fatchoy.v1/x/cipher"
)

type v2Codec struct {
	maxUpstreamPacketBytes   int // 上行包最大大小
	maxDownstreamPacketBytes int // 下行包最大大小
}

var V2 = NewV2()

func NewV2() ICodec {
	var upLimit = GetEnvInt("V2_UP_PKT_BYTES", PacketBytesLimit)
	var downLimit = GetEnvInt("V2_DOWN_PKT_BYTES", PacketBytesLimit)
	if upLimit <= 0 || upLimit > PacketBytesLimit {
		upLimit = PacketBytesLimit
	}
	if downLimit <= 0 || downLimit > PacketBytesLimit {
		downLimit = PacketBytesLimit
	}
	return &v2Codec{
		maxUpstreamPacketBytes:   upLimit,
		maxDownstreamPacketBytes: downLimit,
	}
}

func (c *v2Codec) Marshal(w io.Writer, pkt fatchoy.IPacket, encryptor cipher.BlockCryptor) (int, error) {
	return marshalPacket(w, pkt, encryptor, VersionV2, c.maxDownstreamPacketBytes)
}

func (c *v2Codec) Unmarshal(r io.Reader, header *Header, pkt fatchoy.IPacket, decrypt cipher.BlockCryptor) (int, error) {
	return unmarshalPacket(r, header, pkt, decrypt, c.maxUpstreamPacketBytes)
}
