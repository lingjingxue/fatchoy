package packet

import (
	"google.golang.org/protobuf/proto"
	"qchen.fun/fatchoy"
)

type FakePacket struct {
	PacketBase
}

func MakeFake() *FakePacket {
	return &FakePacket{}
}

func (m *FakePacket) Clone() fatchoy.IPacket {
	var clone = MakeFake()
	clone.Cmd = m.Cmd
	clone.Seqno = m.Seqno
	clone.Flag = m.Flag
	clone.Refer = m.Refer
	clone.Body_ = m.Body_
	return clone
}

// body的类型仅支持int64/float64/string/bytes/proto.Message
func (m *FakePacket) ReplyAny(body interface{}) error {
	var pkt = MakeFake()
	pkt.Cmd = m.Cmd
	pkt.Flag = m.Flag
	pkt.Seqno = m.Seqno
	pkt.Refer = m.Refer
	pkt.Body_ = body
	return m.endpoint.SendPacket(pkt)
}

// 响应proto消息内容
func (m *FakePacket) Reply(ack proto.Message) error {
	var pkt = MakeFake()
	pkt.Cmd = m.Cmd
	pkt.Flag = m.Flag
	pkt.Seqno = m.Seqno
	pkt.Refer = m.Refer
	pkt.Body_ = ack
	return m.endpoint.SendPacket(pkt)
}

// 返回一个错误码消息
func (m *FakePacket) Refuse(errno int32) error {
	var pkt = MakeFake()
	pkt.Cmd = m.Cmd
	pkt.Flag = m.Flag | fatchoy.PFlagError
	pkt.Seqno = m.Seqno
	pkt.Refer = m.Refer
	pkt.SetErrno(errno)
	return m.endpoint.SendPacket(pkt)
}
