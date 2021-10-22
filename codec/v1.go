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
	maxUpstreamPacketBytes   int // 上行包最大大小
	maxDownstreamPacketBytes int // 下行包最大大小
}

var V1 = NewV1()

func NewV1() ICodec {
	var upLimit = GetEnvInt("V1_UP_PKT_BYTES", 60*1024)       // 60K
	var downLimit = GetEnvInt("V1_DOWN_PKT_BYTES", 1024*1024) // 1M
	if upLimit <= 0 || upLimit > PacketBytesLimit {
		upLimit = PacketBytesLimit
	}
	if downLimit <= 0 || downLimit > PacketBytesLimit {
		downLimit = PacketBytesLimit
	}
	return &v1Codec{
		maxUpstreamPacketBytes:   upLimit,
		maxDownstreamPacketBytes: downLimit,
	}
}

// 内部不应该修改pkt的body
func (c *v1Codec) Marshal(w io.Writer, pkt fatchoy.IPacket, encryptor cipher.BlockCryptor) (int, error) {
	return marshalPacket(w, pkt, encryptor, VersionV1, c.maxDownstreamPacketBytes)
}

func (c *v1Codec) Unmarshal(r io.Reader, header *Header, pkt fatchoy.IPacket, decrypt cipher.BlockCryptor) (int, error) {
	return unmarshalPacket(r, header, pkt, decrypt, c.maxUpstreamPacketBytes)
}

func marshalPacket(w io.Writer, pkt fatchoy.IPacket, encryptor cipher.BlockCryptor, version, maxPacketBytes int) (int, error) {
	payload, err := pkt.EncodeBodyToBytes()
	if err != nil {
		return 0, err
	}
	if len(payload) > 0 && encryptor != nil {
		payload = encryptor.Encrypt(payload)
		pkt.SetFlag(pkt.Flag() | fatchoy.PacketFlagEncrypted)
	}

	if n := len(payload); n >= maxPacketBytes {
		err = errors.Errorf("message %v too large payload %d/%d", pkt.Command(), n, maxPacketBytes)
		return 0, err
	}

	var bodyLen = len(payload)
	var head Header
	head.Pack(pkt, bodyLen, version)
	head.SetupChecksum(payload)

	nbytes, err := w.Write(head[:])
	if err == nil && bodyLen > 0 {
		bodyLen, err = w.Write(payload)
		nbytes += bodyLen
	}
	return nbytes, err
}

func unmarshalPacket(r io.Reader, header *Header, pkt fatchoy.IPacket, decrypt cipher.BlockCryptor, maxPacketBytes int) (int, error) {
	var pktLen = header.Len()
	if pktLen < HeaderSize {
		return 0, errors.Errorf("packet length %d out of range", pktLen)
	}
	var flag = fatchoy.PacketFlag(header.Flag())
	pkt.SetFlag(flag)
	pkt.SetType(fatchoy.PacketType(header.Type()))
	pkt.SetSeq(header.Seq())
	pkt.SetCommand(header.Command())

	var checksum = header.Checksum()
	if pktLen > maxPacketBytes {
		err := errors.Errorf("packet %v payload size overflow %d/%d", pkt.Command(), pktLen, maxPacketBytes)
		return 0, err
	}

	if pktLen == HeaderSize {
		if crc := header.CalcChecksum(nil); crc != checksum {
			err := errors.Errorf("message %v header checksum mismatch %x != %x", pkt.Command(), checksum, crc)
			return 0, err
		}
		return pktLen, nil
	}

	var payload = make([]byte, pktLen-HeaderSize)
	if _, err := io.ReadFull(r, payload); err != nil {
		return 0, err
	}
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
	return pktLen, nil
}
