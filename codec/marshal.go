// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"encoding/binary"
	"fmt"

	"gopkg.in/qchencc/fatchoy.v1"
	"gopkg.in/qchencc/fatchoy.v1/codes"
	"gopkg.in/qchencc/fatchoy.v1/x/cipher"
	"gopkg.in/qchencc/fatchoy.v1/x/fsutil"
)

func marshalPacketBody(pkt fatchoy.IPacket, threshold int, encryptor cipher.BlockCryptor) ([]byte, error) {
	var flag = pkt.Flag()
	var body = pkt.BodyToBytes()
	if threshold > 0 && len(body) > V2CompressThreshold {
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
	return body, nil
}

func unmarshalPacketBody(body []byte, decrypt cipher.BlockCryptor, pkt fatchoy.IPacket) error {
	var flag = pkt.Flag()
	if (flag & fatchoy.PFlagEncrypted) != 0 {
		if decrypt == nil {
			return fmt.Errorf("packet %v must be decrypted", pkt.Command())
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
			pkt.SetBody(val)
		} else {
			pkt.SetBody(int64(codes.TransportFailure))
		}
	} else {
		pkt.SetBody(body)
	}
	return nil
}
