// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"encoding/binary"
	"fmt"
	"io"

	"gopkg.in/qchencc/fatchoy.v1"
	"gopkg.in/qchencc/fatchoy.v1/x/cipher"
)

// read length-prefixed data
func ReadLenData(r io.Reader) ([]byte, error) {
	var tmp [4]byte
	if _, err := io.ReadFull(r, tmp[:]); err != nil {
		return nil, err
	}
	var length = binary.BigEndian.Uint32(tmp[:])
	if length > MaxPayloadBytes {
		return nil, fmt.Errorf("payload size %d overflow", length)
	}
	var buf = make([]byte, length-4)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// write length-prefixed data
func WriteLenData(w io.Writer, data []byte) (int, error) {
	var length = uint32(len(data))
	if length > MaxPayloadBytes {
		return 0, fmt.Errorf("payload size %d overflow", length)
	}
	var tmp [4]byte
	binary.BigEndian.PutUint32(tmp[:], length+4)
	if _, err := w.Write(tmp[:]); err != nil {
		return 0, err
	}
	n, err := w.Write(data)
	if err != nil {
		return 0, err
	}
	return n + 4, nil
}

// 使用从r读取消息到pkt，并按需使用decrypt解密，返回读取长度和错误
func ReadV1(r io.Reader) (Header, []byte, error) {
	var headbuf [HeaderSize]byte
	if _, err := io.ReadFull(r, headbuf[:]); err != nil {
		return nil, nil, err
	}
	var head = Header(headbuf[:])
	var length = head.Len()
	if length > MaxPayloadBytes {
		return nil, nil, fmt.Errorf("payload size %d overflow", length)
	}
	var payload = make([]byte, length-HeaderSize)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, nil, err
	}
	return head, payload, nil
}

// 读取一个packet
func ReadPacket(r io.Reader, decrypt cipher.BlockCryptor, pkt fatchoy.IPacket) error {
	head, body, err := ReadV1(r)
	if err != nil {
		return err
	}
	if err := UnmarshalV1(head, body, decrypt, pkt); err != nil {
		return err
	}
	return nil
}

// 写入一个packet
func WritePacket(w io.Writer, encrypt cipher.BlockCryptor, pkt fatchoy.IPacket) error {
	buf, err := MarshalV1(pkt, encrypt)
	if err != nil {
		return err
	}
	if _, err := w.Write(buf); err != nil {
		return err
	}
	return nil
}
