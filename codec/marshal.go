// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"qchen.fun/fatchoy"
	"qchen.fun/fatchoy/x/cipher"
	"qchen.fun/fatchoy/x/fsutil"
)

// 把packet序列化为字节流，有压缩和加密
func marshalPacketBody(pkt fatchoy.IPacket, threshold int, encryptor cipher.BlockCryptor) ([]byte, error) {
	var flag = pkt.Flag()
	var body = pkt.BodyToBytes()
	if threshold > 0 && len(body) > threshold {
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

// 把字节流反序列化为packet，有解密和解压
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
		x, _ := binary.Varint(body) // TODO: deal varint error
		pkt.SetBody(x)
	} else {
		pkt.SetBody(body)
	}
	return nil
}

func md5Sum(data []byte) string {
	var hash = md5.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}
