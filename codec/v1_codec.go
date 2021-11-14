// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"fmt"

	"gopkg.in/qchencc/fatchoy.v1"
	"gopkg.in/qchencc/fatchoy.v1/x/cipher"
	"gopkg.in/qchencc/fatchoy.v1/x/fsutil"
)

var V1CompressThreshold = 4096 // 默认压缩阈值，4K

// 内部除了flag不应该修改pkt的其它字段
func MarshalV1(pkt fatchoy.IPacket, encryptor cipher.BlockCryptor) ([]byte, error) {
	var flag = pkt.Flag()
	var body = pkt.BodyToBytes()
	if V1CompressThreshold > 0 && len(body) > V2CompressThreshold {
		if data, err := fsutil.CompressBytes(body); err != nil {
			return nil, fmt.Errorf("compress packet %v: %w", pkt.Command(), err)
		} else {
			body = data
			flag |= fatchoy.PFlagCompressed
		}
	}
	if len(body) > 0 && encryptor != nil {
		body = encryptor.Encrypt(body)
		flag |= fatchoy.PFlagEncrypted
	}
	pkt.SetFlag(flag)

	var nbytes = V1HeaderSize + len(body)
	if nbytes > MaxPayloadBytes {
		return nil, fmt.Errorf("payload size %d overflow", nbytes)
	}
	var buf = make([]byte, nbytes)
	copy(buf[V1HeaderSize:], body)

	var head = V1Header(buf[:V1HeaderSize])
	head.Pack(pkt, uint16(nbytes))
	var checksum = head.CalcChecksum(buf[V1HeaderSize:])
	head.SetChecksum(checksum)
	return buf, nil
}

// 解码消息到pkt
func UnmarshalV1(header V1Header, payload []byte, decrypt cipher.BlockCryptor, pkt fatchoy.IPacket) error {
	var flag = fatchoy.PacketFlag(header.Flag())
	pkt.SetSeq(header.Seq())
	pkt.SetCommand(header.Command())

	var checksum = header.Checksum()
	if crc := header.CalcChecksum(payload); crc != checksum {
		return fmt.Errorf("message %v checksum mismatch %x != %x", pkt.Command(), checksum, crc)
	}
	if len(payload) == 0 {
		return nil
	}
	var body = payload[:]
	if (flag & fatchoy.PFlagEncrypted) != 0 {
		if decrypt == nil {
			return fmt.Errorf("message %v must be decrypted", pkt.Command())
		}
		body = decrypt.Decrypt(body)
		flag = flag &^ fatchoy.PFlagEncrypted
	}
	if (flag & fatchoy.PFlagCompressed) != 0 {
		if uncompressed, err := fsutil.UncompressBytes(body); err != nil {
			return fmt.Errorf("decompress packet %d: %w", pkt.Command(), err)
		} else {
			body = uncompressed
			flag = flag &^ fatchoy.PFlagCompressed
		}
	}
	pkt.SetFlag(flag)
	pkt.SetBodyBytes(body)
	return nil
}
