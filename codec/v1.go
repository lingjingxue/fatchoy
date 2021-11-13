// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"encoding/binary"
	"fmt"
	"io"

	"gopkg.in/qchencc/fatchoy.v1/codes"
	"gopkg.in/qchencc/fatchoy.v1/x/fsutil"

	"gopkg.in/qchencc/fatchoy.v1"
	"gopkg.in/qchencc/fatchoy.v1/x/cipher"
)

var CompressThreshold = 4096 // 默认压缩阈值，4K

// 内部不应该修改pkt的body
func MarshalV1(w io.Writer, pkt fatchoy.IPacket, encryptor cipher.BlockCryptor) (int, error) {
	var flag = pkt.Flag()
	var payload = pkt.BodyToBytes()
	if CompressThreshold > 0 && len(payload) > CompressThreshold {
		if data, err := fsutil.CompressBytes(payload); err != nil {
			return 0, fmt.Errorf("compress packet %v: %w", pkt.Command(), err)
		} else {
			payload = data
			flag |= fatchoy.PFlagCompressed
		}
	}
	if len(payload) > 0 && encryptor != nil {
		payload = encryptor.Encrypt(payload)
		flag |= fatchoy.PFlagEncrypted
	}

	var nbytes = HeaderSize + len(payload)
	if nbytes > MaxPayloadBytes {
		return 0, fmt.Errorf("payload size %d overflow", nbytes)
	}

	pkt.SetFlag(flag)

	var head = NewHeader()
	head.Pack(pkt, uint32(nbytes))
	head.SetChecksum(head.CalcChecksum(payload))

	if _, err := w.Write(head); err != nil {
		return 0, err
	}
	if len(payload) > 0 {
		if _, err := w.Write(payload); err != nil {
			return 0, err
		}
	}
	return nbytes, nil
}

func UnmarshalV1(header Header, body []byte, decrypt cipher.BlockCryptor, pkt fatchoy.IPacket) error {
	var flag = fatchoy.PacketFlag(header.Flag())
	pkt.SetType(fatchoy.PacketType(header.Type()))
	pkt.SetSeq(header.Seq())
	pkt.SetCommand(header.Command())

	var checksum = header.Checksum()
	if crc := header.CalcChecksum(body); crc != checksum {
		return fmt.Errorf("message %v checksum mismatch %x != %x", pkt.Command(), checksum, crc)
	}

	if len(body) == 0 {
		return nil
	}

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
	// 如果有FlagError，则body是数值错误码
	if (flag & fatchoy.PFlagError) != 0 {
		val, n := binary.Varint(body)
		if n > 0 {
			pkt.SetBodyInt(val)
		} else {
			pkt.SetBodyInt(int64(codes.TransportFailure))
		}
	} else {
		pkt.SetBodyBytes(body)
	}
	return nil
}
