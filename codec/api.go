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

const (
	MaxPayloadBytes = (1 << 16) - 1
)

// read 2-bytes length-prefixed data
func ReadLenData(r io.Reader) ([]byte, error) {
	var tmp [2]byte
	if _, err := io.ReadFull(r, tmp[:]); err != nil {
		return nil, err
	}
	var length = binary.BigEndian.Uint16(tmp[:])
	if length > MaxPayloadBytes {
		return nil, fmt.Errorf("payload size %d overflow", length)
	}
	var buf = make([]byte, length-2)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// write 2-bytes length-prefixed data
func WriteLenData(w io.Writer, data []byte) (int, error) {
	var length = len(data) + 2
	if length > MaxPayloadBytes {
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
