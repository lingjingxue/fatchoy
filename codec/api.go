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

// read 3bytes length-prefixed data
func ReadLenData(r io.Reader) ([]byte, error) {
	var tmp [3]byte
	if _, err := io.ReadFull(r, tmp[:]); err != nil {
		return nil, err
	}
	var length = uint32(tmp[2]) | uint32(tmp[1])<<8 | uint32(tmp[0])<<16 // big endian
	if length > MaxPayloadBytes {
		return nil, fmt.Errorf("payload size %d overflow", length)
	}
	var buf = make([]byte, length-3)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// write 3bytes length-prefixed data
func WriteLenData(w io.Writer, data []byte) (int, error) {
	var length = uint32(len(data)) + 3
	if length > MaxPayloadBytes {
		return 0, fmt.Errorf("payload size %d overflow", length)
	}
	// big endian
	var tmp [3]byte
	tmp[0] = byte(length >> 16)
	tmp[1] = byte(length >> 8)
	tmp[2] = byte(length)

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
func ReadV1(r io.Reader) (V1Header, []byte, error) {
	var headbuf [V1HeaderSize]byte
	if _, err := io.ReadFull(r, headbuf[:]); err != nil {
		return nil, nil, err
	}
	var head = V1Header(headbuf[:])
	var length = head.Len()
	if length > MaxPayloadBytes {
		return nil, nil, fmt.Errorf("payload size %d overflow", length)
	}
	var payload = make([]byte, length-V1HeaderSize)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, nil, err
	}
	return head, payload, nil
}

// 使用从r读取消息到pkt，并按需使用decrypt解密，返回读取长度和错误
func ReadV2(r io.Reader) (V2Header, []byte, error) {
	var headbuf [V2HeaderSize]byte
	if _, err := io.ReadFull(r, headbuf[:]); err != nil {
		return nil, nil, err
	}
	var head = V2Header(headbuf[:])
	var length = head.Len()
	if length > MaxPayloadBytes {
		return nil, nil, fmt.Errorf("payload size %d overflow", length)
	}
	var payload = make([]byte, length-V2HeaderSize)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, nil, err
	}
	return head, payload, nil
}

// 读取一个packet
func ReadPacketV1(r io.Reader, decrypt cipher.BlockCryptor, pkt fatchoy.IPacket) error {
	head, body, err := ReadV1(r)
	if err != nil {
		return err
	}
	if err := UnmarshalV1(head, body, decrypt, pkt); err != nil {
		return err
	}
	return nil
}

// 读取一个packet
func ReadPacketV2(r io.Reader, decrypt cipher.BlockCryptor, pkt fatchoy.IPacket) error {
	head, body, err := ReadV2(r)
	if err != nil {
		return err
	}
	if err := UnmarshalV2(head, body, decrypt, pkt); err != nil {
		return err
	}
	return nil
}

// 写入一个packet
func WritePacketV1(w io.Writer, encrypt cipher.BlockCryptor, pkt fatchoy.IPacket) error {
	buf, err := MarshalV1(pkt, encrypt)
	if err != nil {
		return err
	}
	if _, err := w.Write(buf); err != nil {
		return err
	}
	return nil
}

// 写入一个packet
func WritePacketV2(w io.Writer, encrypt cipher.BlockCryptor, pkt fatchoy.IPacket) error {
	buf, err := MarshalV2(pkt, encrypt)
	if err != nil {
		return err
	}
	if _, err := w.Write(buf); err != nil {
		return err
	}
	return nil
}
