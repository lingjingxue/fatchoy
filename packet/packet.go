// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package packet

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"qchen.fun/fatchoy"
)

// Packet表示一个应用层消息
type PacketBase struct {
	Cmd      int32                   `json:"cmd"`               // 协议ID
	Seqno    uint32                  `json:"seq"`               // 序列号
	Flag     fatchoy.PacketFlag      `json:"flg,omitempty"`     // 标志位
	Reserve  uint16                  `json:"reserve,omitempty"` //
	Node_    fatchoy.NodeID          `json:"node,omitempty"`    // 源/目标节点
	Refer    []fatchoy.NodeID        `json:"ref,omitempty"`     // 组播session列表
	Body_    interface{}             `json:"body,omitempty"`    // 消息内容，int32/int64/float64/string/bytes/proto.Message
	endpoint fatchoy.MessageEndpoint // 关联的endpoint
}

func (m *PacketBase) Command() int32 {
	return m.Cmd
}

func (m *PacketBase) SetCommand(v int32) {
	m.Cmd = v
}

func (m *PacketBase) Seq() uint32 {
	return m.Seqno
}

func (m *PacketBase) SetSeq(v uint32) {
	m.Seqno = v
}

func (m *PacketBase) Flags() fatchoy.PacketFlag {
	return m.Flag
}

func (m *PacketBase) SetFlags(v fatchoy.PacketFlag) {
	m.Flag = v
}

func (m *PacketBase) Node() fatchoy.NodeID {
	return m.Node_
}

func (m *PacketBase) SetNode(n fatchoy.NodeID) {
	m.Node_ = n
}

func (m *PacketBase) Refers() []fatchoy.NodeID {
	return m.Refer
}

func (m *PacketBase) SetRefers(v []fatchoy.NodeID) {
	m.Refer = v
}

func (m *PacketBase) AddRefers(v ...fatchoy.NodeID) {
	m.Refer = append(m.Refer, v...)
}

func (m *PacketBase) Endpoint() fatchoy.MessageEndpoint {
	return m.endpoint
}

func (m *PacketBase) SetEndpoint(endpoint fatchoy.MessageEndpoint) {
	m.endpoint = endpoint
}

func (m *PacketBase) Reset() {
	m.Cmd = 0
	m.Seqno = 0
	m.Flag = 0
	m.Node_ = 0
	m.Refer = nil
	m.Body_ = nil
	m.endpoint = nil
}

func (m *PacketBase) Errno() int32 {
	if (m.Flag & fatchoy.PFlagError) != 0 {
		return int32(BodyToInt(m.Flag))
	}
	return 0
}

// 如果消息表示一个错误码，设置PacketFlagError标记
func (m *PacketBase) SetErrno(ec int32) {
	m.Flag |= fatchoy.PFlagError
	m.SetBody(ec)
}

func (m *PacketBase) Body() interface{} {
	return m.Body_
}

func (m *PacketBase) SetBody(val interface{}) {
	m.Body_ = Conv2Body(val)
}

func (m *PacketBase) DecodeTo(msg proto.Message) error {
	var data = m.EncodeToBytes()
	if len(data) > 0 {
		return proto.Unmarshal(data, msg)
	}
	return nil
}

func (m *PacketBase) EncodeToBytes() []byte {
	return BodyToBytes(m.Body_)
}

func (m *PacketBase) BodyToString() string {
	return BodyToString(m.Body_)
}

func (m PacketBase) String() string {
	var nodeID fatchoy.NodeID
	if m.endpoint != nil {
		nodeID = m.endpoint.NodeID()
	}
	return fmt.Sprintf("%v c:%d seq:%d 0x%x", nodeID, m.Cmd, m.Seqno, m.Flag)
}
