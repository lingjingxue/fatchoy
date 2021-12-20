// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"encoding/binary"
	"fmt"
	"io"

	"qchen.fun/fatchoy"
	"qchen.fun/fatchoy/x/cipher"
)

// V2格式编码
type codecV2 struct {
	threshold int
}

func NewV2Encoder(threshold int) Encoder {
	if threshold <= 0 {
		threshold = 8192 // 默认压缩阈值，8K
	}
	return &codecV2{
		threshold: threshold,
	}
}

func init() {
	Register(NewV2Encoder(8192))
}

func (c *codecV2) Name() string {
	return "V2"
}

func (c *codecV2) Version() int {
	return V2EncoderVersion
}

// 把`pkt`编码到`w`，内部除了flag不应该修改pkt的其它字段
func (c *codecV2) WritePacket(w io.Writer, encrypt cipher.BlockCryptor, pkt fatchoy.IPacket) (int, error) {
	var refers = pkt.Refers()
	if n := len(refers); n > V2MaxReferCount {
		return 0, fmt.Errorf("packet %d refer count overflow %d/%d", pkt.Command(), n, V2MaxReferCount)
	}
	body, err := marshalPacketBody(pkt, c.threshold, encrypt)
	if err != nil {
		return 0, err
	}

	var nheader = V2HeaderSize + len(refers)*4
	var nbytes = nheader + len(body)
	if nbytes > V2MaxPayloadBytes {
		return 0, fmt.Errorf("packet %d payload size %d overflow", pkt.Command(), nbytes)
	}
	var headbuf = make([]byte, nheader)
	if len(refers) > 0 {
		var i = V2HeaderSize
		for _, node := range refers {
			binary.BigEndian.PutUint32(headbuf[i:], uint32(node))
			i += 4
		}
	}
	var head = V2Header(headbuf[:nheader])
	head.Pack(pkt, uint16(len(refers)), uint32(nbytes))
	var checksum = head.CalcChecksum(headbuf[V2HeaderSize:], body)
	head.SetChecksum(checksum)

	if _, err := w.Write(headbuf); err != nil {
		return 0, err
	}
	if _, err := w.Write(body); err != nil {
		return 0, err
	}
	return nbytes, nil
}

// 按V2协议格式读取head和body
func (codecV2) ReadHeadBody(r io.Reader) ([]byte, []byte, error) {
	var buf [V2HeaderSize]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return nil, nil, err
	}
	var head = V2Header(buf[:])
	var length = head.Len()
	if length > V2MaxPayloadBytes {
		return nil, nil, fmt.Errorf("payload size %d overflow", length)
	}
	var payload = make([]byte, length-V2HeaderSize)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, nil, err
	}
	return head, payload, nil
}

// 解码消息到`pkt`
func (codecV2) UnmarshalPacket(header, body []byte, decrypt cipher.BlockCryptor, pkt fatchoy.IPacket) error {
	var head = V2Header(header)
	pkt.SetFlag(fatchoy.PacketFlag(head.Flag()))
	pkt.SetSeq(head.Seq())
	pkt.SetCommand(head.Command())
	pkt.SetNode(head.Node())

	var checksum = head.Checksum()
	if crc := head.CalcChecksum(nil, body); crc != checksum {
		return fmt.Errorf("packet %v checksum mismatch %x != %x", pkt.Command(), checksum, crc)
	}
	var pos = 0
	var refcnt = head.RefCount()
	if refcnt > V2MaxReferCount {
		return fmt.Errorf("packet %d refer count overflow %d/%d", pkt.Command(), refcnt, V2MaxReferCount)
	}
	if refcnt > 0 {
		if len(body) < int(refcnt)*4 {
			return fmt.Errorf("packet %d refer count mismatch %d != %d", pkt.Command(), len(body)/4, refcnt)
		}
		var refers = make([]fatchoy.NodeID, 0, refcnt)
		for i := 0; i < int(refcnt); i++ {
			var refer = binary.BigEndian.Uint32(body[pos:])
			pos += 4
			refers = append(refers, fatchoy.NodeID(refer))
		}
		pkt.SetRefers(refers)
	}
	body = body[pos:]
	if len(body) > 0 {
		return unmarshalPacketBody(body, decrypt, pkt)
	}
	return nil
}

// 从`r`里读取消息到`pkt`
func (c *codecV2) ReadPacket(r io.Reader, decrypt cipher.BlockCryptor, pkt fatchoy.IPacket) error {
	head, body, err := c.ReadHeadBody(r)
	if err != nil {
		return err
	}
	if err := c.UnmarshalPacket(head, body, decrypt, pkt); err != nil {
		return err
	}
	return nil
}
