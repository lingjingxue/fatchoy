// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"fmt"

	"gopkg.in/qchencc/fatchoy.v1"
	"gopkg.in/qchencc/fatchoy.v1/x/cipher"
)

var V1CompressThreshold = 4096 // 默认压缩阈值，4K

// 内部除了flag不应该修改pkt的其它字段
func MarshalV1(pkt fatchoy.IPacket, encryptor cipher.BlockCryptor) ([]byte, error) {
	body, err := marshalPacketBody(pkt, V1CompressThreshold, encryptor)
	if err != nil {
		return nil, err
	}
	var nbytes = V1HeaderSize + len(body)
	if nbytes > V1MaxPayloadBytes {
		return nil, fmt.Errorf("packet %d payload size %d overflow", pkt.Command(), nbytes)
	}
	var buf = make([]byte, nbytes)
	copy(buf[V1HeaderSize:], body)

	var head = V1Header(buf[:V1HeaderSize])
	head.Pack(pkt, uint16(nbytes))
	var checksum = head.CalcChecksum(buf[V1HeaderSize:])
	head.SetChecksum(checksum)
	return buf, nil
}

// 解码消息到pkt
func UnmarshalV1(head V1Header, payload []byte, decrypt cipher.BlockCryptor, pkt fatchoy.IPacket) error {
	pkt.SetFlag(fatchoy.PacketFlag(head.Flag()))
	pkt.SetSeq(head.Seq())
	pkt.SetCommand(head.Command())

	var checksum = head.Checksum()
	if crc := head.CalcChecksum(payload); crc != checksum {
		return fmt.Errorf("packet %v checksum mismatch %x != %x", pkt.Command(), checksum, crc)
	}
	if len(payload) == 0 {
		return nil
	}
	return unmarshalPacketBody(payload, decrypt, pkt)
}
