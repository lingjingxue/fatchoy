// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"fmt"
	"io"

	"gopkg.in/qchencc/fatchoy.v1"
	"gopkg.in/qchencc/fatchoy.v1/x/cipher"
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
	return 0, fmt.Errorf("codec version %d unrecognized", ver)
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
		return 0, fmt.Errorf("codec version %d unrecognized", ver)
	}
}
