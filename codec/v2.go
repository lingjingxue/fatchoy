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
	maxPayloadBytes int32
}

var V2 = NewV2()

func NewV2() ICodec {
	return &v2Codec{
		maxPayloadBytes: 1 << 22, // 4M
	}
}

func (c *v2Codec) Marshal(w io.Writer, pkt fatchoy.IPacket, encryptor cipher.BlockCryptor) (int, error) {
	return marshalPacket(w, pkt, encryptor, VersionV2, int(c.maxPayloadBytes))
}

func (c *v2Codec) Unmarshal(r io.Reader, header *Header, pkt fatchoy.IPacket, decrypt cipher.BlockCryptor) (int, error) {
	return unmarshalPacket(r, header, pkt, decrypt, int(c.maxPayloadBytes))
}
