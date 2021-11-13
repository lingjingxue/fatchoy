// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"encoding/binary"
	"fmt"
	"math"

	"gopkg.in/qchencc/fatchoy.v1/codes"
	"gopkg.in/qchencc/fatchoy.v1/x/fsutil"

	"gopkg.in/qchencc/fatchoy.v1"
	"gopkg.in/qchencc/fatchoy.v1/x/cipher"
)

var CompressThreshold = 4096 // 默认压缩阈值，4K

// 内部不应该修改pkt的body
func MarshalV1(pkt fatchoy.IPacket, encryptor cipher.BlockCryptor) ([]byte, error) {
	var flag = pkt.Flag()
	var refer = pkt.Refer()
	if n := len(refer); n > math.MaxUint8 {
		return nil, fmt.Errorf("refer count #%d overflow", n)
	}
	var body = pkt.BodyToBytes()
	if CompressThreshold > 0 && len(body) > CompressThreshold {
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

	var nn = HeaderSize + len(refer)*4
	var nbytes = nn + len(body)
	if nbytes > MaxPayloadBytes {
		return nil, fmt.Errorf("payload size %d overflow", nbytes)
	}
	var buf = make([]byte, nbytes)
	copy(buf[nn:], body)

	if len(refer) > 0 {
		var i = HeaderSize
		for _, ref := range refer {
			binary.BigEndian.PutUint32(buf[i:], ref)
			i += 4
		}
	}
	var head = Header(buf[:HeaderSize])
	head.Pack(pkt, uint8(len(refer)), uint32(nbytes))
	var checksum = head.CalcChecksum(buf[HeaderSize:])
	head.SetChecksum(checksum)
	return buf, nil
}

func UnmarshalV1(header Header, payload []byte, decrypt cipher.BlockCryptor, pkt fatchoy.IPacket) error {
	var flag = fatchoy.PacketFlag(header.Flag())
	pkt.SetType(fatchoy.PacketType(header.Type()))
	pkt.SetSeq(header.Seq())
	pkt.SetCommand(header.Command())

	var checksum = header.Checksum()
	if crc := header.CalcChecksum(payload); crc != checksum {
		return fmt.Errorf("message %v checksum mismatch %x != %x", pkt.Command(), checksum, crc)
	}

	if len(payload) == 0 {
		return nil
	}
	var refcnt = header.RefCount()
	if refcnt > 0 {
		if len(payload) < int(refcnt)*4 {
			return fmt.Errorf("message %d refer count mismatch %d != %d", pkt.Command(), len(payload)/4, refcnt)
		}
		var pos = 0
		var refers = make([]uint32, 0, refcnt)
		for i := 0; i < int(refcnt); i++ {
			var refer = binary.BigEndian.Uint32(payload[pos:])
			pos += 4
			refers = append(refers, refer)
		}
		pkt.SetRefer(refers)
	}
	var body = payload[refcnt*4:]
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
