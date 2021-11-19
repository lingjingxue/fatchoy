// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"encoding/binary"
	"fmt"
	"math"

	"qchen.fun/fatchoy"
	"qchen.fun/fatchoy/x/cipher"
)

var V2CompressThreshold = 4096 // 默认压缩阈值，4K

// 内部除了flag不应该修改pkt的其它字段
func MarshalV2(pkt fatchoy.IPacket, encryptor cipher.BlockCryptor) ([]byte, error) {
	var refers = pkt.Refers()
	if n := len(refers); n > math.MaxUint8 {
		return nil, fmt.Errorf("packet %d refer count #%d overflow", pkt.Command(), n)
	}
	body, err := marshalPacketBody(pkt, V2CompressThreshold, encryptor)
	if err != nil {
		return nil, err
	}

	var nn = V2HeaderSize + len(refers)*4
	var nbytes = nn + len(body)
	if nbytes > V2MaxPayloadBytes {
		return nil, fmt.Errorf("packet %d payload size %d overflow", pkt.Command(), nbytes)
	}
	var buf = make([]byte, nbytes)
	copy(buf[nn:], body)

	if len(refers) > 0 {
		var i = V2HeaderSize
		for _, node := range refers {
			binary.BigEndian.PutUint32(buf[i:], uint32(node))
			i += 4
		}
	}
	var head = V2Header(buf[:V2HeaderSize])
	head.Pack(pkt, uint8(len(refers)), uint32(nbytes))
	var checksum = head.CalcChecksum(buf[V2HeaderSize:])
	head.SetChecksum(checksum)
	return buf, nil
}

// 解码消息到pkt
func UnmarshalV2(head V2Header, payload []byte, decrypt cipher.BlockCryptor, pkt fatchoy.IPacket) error {
	pkt.SetFlag(fatchoy.PacketFlag(head.Flag()))
	pkt.SetType(fatchoy.PacketType(head.Type()))
	pkt.SetSeq(head.Seq())
	pkt.SetCommand(head.Command())
	pkt.SetNode(head.Node())

	var checksum = head.Checksum()
	if crc := head.CalcChecksum(payload); crc != checksum {
		return fmt.Errorf("packet %v checksum mismatch %x != %x", pkt.Command(), checksum, crc)
	}
	var pos = 0
	if refcnt := head.RefCount(); refcnt > 0 {
		if len(payload) < int(refcnt)*4 {
			return fmt.Errorf("packet %d refer count mismatch %d != %d", pkt.Command(), len(payload)/4, refcnt)
		}
		var refers = make([]fatchoy.NodeID, 0, refcnt)
		for i := 0; i < int(refcnt); i++ {
			var refer = binary.BigEndian.Uint32(payload[pos:])
			pos += 4
			refers = append(refers, fatchoy.NodeID(refer))
		}
		pkt.SetRefers(refers)
	}
	var body = payload[pos:]
	if len(body) > 0 {
		return unmarshalPacketBody(body, decrypt, pkt)
	}
	return nil
}
