// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
	"gopkg.in/qchencc/fatchoy.v1"
	"gopkg.in/qchencc/fatchoy.v1/codes"
	"gopkg.in/qchencc/fatchoy.v1/log"
	"gopkg.in/qchencc/fatchoy.v1/x/cipher"
	"gopkg.in/qchencc/fatchoy.v1/x/fsutil"
)

// 消息编解码
type ICodec interface {
	Marshal(w io.Writer, pkt fatchoy.IMessage, encrypt cipher.BlockCryptor) (int, error)
	Unmarshal(r io.Reader, head *Header, pkt fatchoy.IMessage, decrypt cipher.BlockCryptor) (int, error)
}

// 消息编解码，同样一个codec会在多个goroutine执行，需要多线程安全
// 把pkt按需用encrypt加密后编码到w里，，返回编码长度和err
func Marshal(w io.Writer, pkt fatchoy.IMessage, encrypt cipher.BlockCryptor, ver int) (int, error) {
	switch ver {
	case VersionV1:
		return V1.Marshal(w, pkt, encrypt)
	case VersionV2:
		return V2.Marshal(w, pkt, encrypt)
	}
	return 0, errors.Errorf("codec version %d unrecognized", ver)
}

// 使用从r读取消息到pkt，并按需使用decrypt解密，返回读取长度和错误
func Unmarshal(r io.Reader, pkt fatchoy.IMessage, decrypt cipher.BlockCryptor) (int, error) {
	var header Header
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return 0, err
	}
	var ver = header.Version()
	switch ver {
	case VersionV1:
		return V1.Unmarshal(r, &header, pkt, decrypt)
	case VersionV2:
		return V2.Unmarshal(r, &header, pkt, decrypt)
	default:
		return 0, errors.Errorf("codec version %d unrecognized", ver)
	}
}

// 根据pkt的Flag标志位，对body进行压缩
func CompressPacket(pkt fatchoy.IMessage, threshold int) error {
	payload, err := pkt.EncodeBodyToBytes()
	if err != nil {
		return err
	}
	if payload == nil {
		return nil
	}
	if n := len(payload); threshold > 0 && n > threshold {
		if data, err := fsutil.CompressBytes(payload); err != nil {
			log.Errorf("compress packet %v with %d bytes: %v", pkt.Command, n, err)
			return err
		} else {
			payload = data
			pkt.SetFlag(pkt.Flag() | fatchoy.PacketFlagCompressed)
		}
	}
	pkt.SetBodyBytes(payload)
	return nil
}

// 根据pkt的Flag标志位，对body进行解压缩
func UncompressPacket(pkt fatchoy.IMessage) error {
	payload := pkt.BodyAsBytes()
	if payload == nil {
		return nil
	}
	var flag = pkt.Flag()
	if (flag & fatchoy.PacketFlagCompressed) > 0 {
		if uncompressed, err := fsutil.UncompressBytes(payload); err != nil {
			log.Errorf("decompress packet %v(%d bytes): %v", pkt.Command, len(payload), err)
			return err
		} else {
			payload = uncompressed
			pkt.SetFlag(flag &^ fatchoy.PacketFlagCompressed)
		}
	}
	// 如果有FlagError，则body是数值错误码
	if (flag & fatchoy.PacketFlagError) != 0 {
		val, n := binary.Varint(payload)
		if n > 0 {
			pkt.SetBodyNumber(val)
		} else {
			pkt.SetBodyNumber(int64(codes.TransportFailure))
		}
	} else {
		pkt.SetBodyBytes(payload)
	}
	return nil
}
