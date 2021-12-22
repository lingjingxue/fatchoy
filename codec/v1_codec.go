// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"fmt"
	"io"

	"qchen.fun/fatchoy"
	"qchen.fun/fatchoy/x/cipher"
)

// V1格式编码
type codecV1 struct {
	threshold int
}

func NewV1Encoder(threshold int) Encoder {
	if threshold <= 0 {
		threshold = 4096 // 默认压缩阈值，4K
	}
	return &codecV1{
		threshold: threshold,
	}
}

func init() {
	Register(NewV1Encoder(4096))
}

func (c *codecV1) Name() string {
	return "V1"
}

func (c *codecV1) Version() int {
	return V1EncoderVersion
}

// 把`pkt`编码到`w`，内部除了flag不应该修改pkt的其它字段
func (c *codecV1) WritePacket(w io.Writer, encrypt cipher.BlockCryptor, pkt fatchoy.IPacket) (int, error) {
	body, err := marshalPacketBody(pkt, c.threshold, encrypt)
	if err != nil {
		return 0, err
	}
	var nbytes = V1HeaderSize + len(body)
	if nbytes > V1MaxPayloadBytes {
		return 0, fmt.Errorf("packet %d payload size %d overflow", pkt.Command(), nbytes)
	}
	var headbuf = make([]byte, V1HeaderSize)
	var head = V1Header(headbuf)
	head.Pack(pkt, uint16(nbytes))
	var checksum = head.CalcChecksum(body)
	head.SetChecksum(checksum)

	if _, err := w.Write(headbuf); err != nil {
		return 0, err
	}
	if _, err := w.Write(body); err != nil {
		return 0, err
	}
	return nbytes, nil
}

// 按V1协议格式读取head和body
func (codecV1) ReadHeadBody(r io.Reader) ([]byte, []byte, error) {
	var buf [V1HeaderSize]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return nil, nil, err
	}
	var head = V1Header(buf[:])
	var length = head.Len()
	if length > V1MaxPayloadBytes {
		return nil, nil, fmt.Errorf("payload size %d overflow", length)
	}
	var payload = make([]byte, length-V1HeaderSize)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, nil, err
	}
	return head, payload, nil
}

// 解码消息到`pkt`
func (codecV1) UnmarshalPacket(header, body []byte, decrypt cipher.BlockCryptor, pkt fatchoy.IPacket) error {
	var head = V1Header(header)
	pkt.SetFlags(fatchoy.PacketFlag(head.Flag()))
	pkt.SetSeq(head.Seq())
	pkt.SetCommand(head.Command())

	var checksum = head.Checksum()
	if crc := head.CalcChecksum(body); crc != checksum {
		return fmt.Errorf("packet %v checksum mismatch %x != %x", pkt.Command(), checksum, crc)
	}
	if len(body) > 0 {
		return unmarshalPacketBody(body, decrypt, pkt)
	}
	return nil
}

// 从`r`里读取消息到`pkt`
func (c *codecV1) ReadPacket(r io.Reader, decrypt cipher.BlockCryptor, pkt fatchoy.IPacket) error {
	head, body, err := c.ReadHeadBody(r)
	if err != nil {
		return err
	}
	if err := c.UnmarshalPacket(head, body, decrypt, pkt); err != nil {
		return err
	}
	return nil
}
