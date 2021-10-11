// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"io"

	"github.com/pkg/errors"
	"gopkg.in/qchencc/fatchoy.v1"
	"gopkg.in/qchencc/fatchoy.v1/x/cipher"
)

// 编码器
type v1Codec struct {
	maxPayloadBytes int32
}

var V1 = NewV1()

func NewV1() ICodec {
	return &v1Codec{
		maxPayloadBytes: 60 * 1024, // 60K
	}
}

// 内部不应该修改pkt的body
func (c *v1Codec) Marshal(w io.Writer, pkt fatchoy.IMessage, encryptor cipher.BlockCryptor) (int, error) {
	return marshalPacket(w, pkt, encryptor, VersionV1, int(c.maxPayloadBytes))
}

func (c *v1Codec) Unmarshal(r io.Reader, header *Header, pkt fatchoy.IMessage, decrypt cipher.BlockCryptor) (int, error) {
	return unmarshalPacket(r, header, pkt, decrypt, int(c.maxPayloadBytes))
}

func marshalPacket(w io.Writer, pkt fatchoy.IMessage, encryptor cipher.BlockCryptor, version, maxPayloadBytes int) (int, error) {
	payload, err := pkt.EncodeBodyToBytes()
	if err != nil {
		return 0, err
	}
	if len(payload) > 0 && encryptor != nil {
		payload = encryptor.Encrypt(payload)
		pkt.SetFlag(pkt.Flag() | fatchoy.PacketFlagEncrypted)
	}

	if n := len(payload); n >= maxPayloadBytes {
		err = errors.Errorf("message %v too large payload %d/%d", pkt.Command(), n, maxPayloadBytes)
		return 0, err
	}

	var bodyLen = len(payload)
	var head Header
	head.unmarshalFrom(pkt, bodyLen, version)
	head.SetupChecksum(payload)

	nbytes, err := w.Write(head[:])
	if err == nil && bodyLen > 0 {
		bodyLen, err = w.Write(payload)
		nbytes += bodyLen
	}
	return nbytes, err
}

func unmarshalPacket(r io.Reader, header *Header, pkt fatchoy.IMessage, decrypt cipher.BlockCryptor, maxPayloadBytes int) (int, error) {
	var bodyLen = header.Len()
	var flag = fatchoy.PacketFlag(header.Flag())
	pkt.SetFlag(flag)
	pkt.SetType(fatchoy.PacketType(header.Type()))
	pkt.SetSeq(header.Seq())
	pkt.SetCommand(header.Command())

	var checksum = header.Checksum()
	if bodyLen > maxPayloadBytes {
		err := errors.Errorf("packet %v payload size overflow %d/%d", pkt.Command(), bodyLen, maxPayloadBytes)
		return 0, err
	}

	var nbytes = HeaderSize
	if bodyLen == 0 {
		if crc := header.CalcChecksum(nil); crc != checksum {
			err := errors.Errorf("message %v header checksum mismatch %x != %x", pkt.Command(), checksum, crc)
			return 0, err
		}
		return nbytes, nil
	}

	var payload = make([]byte, bodyLen)
	if _, err := io.ReadFull(r, payload); err != nil {
		return 0, err
	}
	nbytes += bodyLen
	if crc := header.CalcChecksum(payload); checksum != crc {
		err := errors.Errorf("message %v checksum mismatch %x != %x", pkt.Command(), checksum, crc)
		return 0, err
	}
	if (flag & fatchoy.PacketFlagEncrypted) > 0 {
		if decrypt == nil {
			err := errors.Errorf("message %v must be decrypted", pkt.Command())
			return 0, err
		}
		payload = decrypt.Decrypt(payload)
		pkt.SetFlag(flag &^ fatchoy.PacketFlagEncrypted)
	}
	pkt.SetBodyBytes(payload)
	return nbytes, nil
}
