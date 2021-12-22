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
	"qchen.fun/fatchoy/log"
	"qchen.fun/fatchoy/x/cipher"
	"qchen.fun/fatchoy/x/fsutil"
)

// 把packet序列化为字节流，有压缩和加密
func marshalPacketBody(pkt fatchoy.IPacket, threshold int, encryptor cipher.BlockCryptor) ([]byte, error) {
	var flag = pkt.Flags()
	var body = pkt.EncodeToBytes()
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
	pkt.SetFlags(flag)
	return body, nil
}

// 把字节流反序列化为packet，有解密和解压
func unmarshalPacketBody(body []byte, decrypt cipher.BlockCryptor, pkt fatchoy.IPacket) error {
	var flag = pkt.Flags()
	if flag.Has(fatchoy.PFlagEncrypted) {
		if decrypt == nil {
			return fmt.Errorf("packet %v must be decrypted", pkt.Command())
		}
		body = decrypt.Decrypt(body)
		flag = flag.Clear(fatchoy.PFlagEncrypted)
	}
	if flag.Has(fatchoy.PFlagCompressed)  {
		if uncompressed, err := fsutil.UncompressBytes(body); err != nil {
			return fmt.Errorf("decompress packet %d: %w", pkt.Command(), err)
		} else {
			body = uncompressed
			flag = flag.Clear(fatchoy.PFlagCompressed)
		}
	}
	pkt.SetFlags(flag)
	// 如果有FlagError，则body是数值错误码
	if flag.Has(fatchoy.PFlagError)  {
		if len(body) == 4 {
			pkt.SetBody(int32(binary.LittleEndian.Uint32(body)))
		} else {
			log.Errorf("packet %d errno invalid length %d", len(body))
		}
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
