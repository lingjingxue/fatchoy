// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"qchen.fun/fatchoy"
	"qchen.fun/fatchoy/x/cipher"
)

// 消息编/解码接口
type Encoder interface {
	Name() string
	Version() int

	// 把`pkt`编码到`w`，内部除了flag不应该修改pkt的其它字段
	WritePacket(w io.Writer, encrypt cipher.BlockCryptor, pkt fatchoy.IPacket) (int, error)

	// 按协议格式读取head和body
	ReadHeadBody(r io.Reader) ([]byte, []byte, error)

	// 根据head和body解码消息到`pkt`
	UnmarshalPacket(header, body []byte, decrypt cipher.BlockCryptor, pkt fatchoy.IPacket) error

	// 从`r`里读取消息到`pkt`
	ReadPacket(r io.Reader, decrypt cipher.BlockCryptor, pkt fatchoy.IPacket) error
}

// 读取2字节开头的数据
func ReadLenData(r io.Reader) ([]byte, error) {
	var tmp [2]byte
	if _, err := io.ReadFull(r, tmp[:]); err != nil {
		return nil, err
	}
	var length = binary.BigEndian.Uint16(tmp[:])
	var buf = make([]byte, length-2)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// 写入2字节开头的数据
func WriteLenData(w io.Writer, data []byte) (int, error) {
	var length = len(data) + 2
	if length >= math.MaxUint16 {
		return 0, fmt.Errorf("payload size %d overflow", length)
	}
	var tmp [2]byte
	binary.BigEndian.PutUint16(tmp[:], uint16(length))
	if _, err := w.Write(tmp[:]); err != nil {
		return 0, err
	}
	n, err := w.Write(data)
	if err != nil {
		return 0, err
	}
	return n + 4, nil
}
