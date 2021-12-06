// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"time"

	"qchen.fun/fatchoy"
	"qchen.fun/fatchoy/codec"
	"qchen.fun/fatchoy/packet"
	"qchen.fun/fatchoy/x/cipher"

	"google.golang.org/protobuf/proto"
)

var (
	RequestReadTimeout = 100 // 默认read超时时间，100s
)

// 读取2字节开头的message
func ReadLenMessage(conn net.Conn, msg proto.Message) error {
	var deadline = time.Now().Add(time.Duration(RequestReadTimeout) * time.Second)
	conn.SetReadDeadline(deadline)
	body, err := codec.ReadLenData(conn)
	if err != nil {
		return err
	}
	return proto.Unmarshal(body, msg)
}

// 写入2字节开头的message
func WriteLenMessage(conn net.Conn, msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if _, err := codec.WriteLenData(&buf, data); err != nil {
		return err
	}
	_, err = conn.Write(buf.Bytes())
	return err
}

// 写入req并且等待读取ack
func RequestLenMessage(conn net.Conn, req, ack proto.Message) error {
	if err := WriteLenMessage(conn, req); err != nil {
		return err
	}
	if err := ReadLenMessage(conn, ack); err != nil {
		return err
	}
	return nil
}

// 读取一条消息
func ReadProtoMessage(conn net.Conn, enc codec.Encoder, decrypt cipher.BlockCryptor, pkt fatchoy.IPacket, msg proto.Message) error {
	var deadline = time.Now().Add(time.Duration(RequestReadTimeout) * time.Second)
	conn.SetReadDeadline(deadline)
	if err := enc.ReadPacket(conn, decrypt, pkt); err != nil {
		return err
	}
	if ec := pkt.Errno(); ec > 0 {
		return fmt.Errorf("message %v has error: %d", pkt.Command(), ec)
	}
	if err := pkt.DecodeTo(msg); err != nil {
		return err
	}
	return nil
}

// send一条protobuf消息
func SendProtoMessage(conn io.Writer, enc codec.Encoder, encrypt cipher.BlockCryptor, command int32, msg proto.Message) error {
	var pkt = packet.New(command, 0, 0, msg)
	_, err := enc.WritePacket(conn, encrypt, pkt)
	return err
}

// send并且立即等待recv(不加密)
func RequestProtoMessage(conn net.Conn, enc codec.Encoder, command int32, req, resp proto.Message) error {
	if err := SendProtoMessage(conn, enc, nil, command, req); err != nil {
		return err
	}
	var pkt = packet.Make()
	if err := ReadProtoMessage(conn, enc, nil, pkt, resp); err != nil {
		return err
	}
	return nil
}
