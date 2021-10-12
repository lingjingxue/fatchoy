// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"gopkg.in/qchencc/fatchoy.v1"
	"gopkg.in/qchencc/fatchoy.v1/codec"
	"gopkg.in/qchencc/fatchoy.v1/log"
	"gopkg.in/qchencc/fatchoy.v1/packet"
	"gopkg.in/qchencc/fatchoy.v1/x/stats"
)

const (
	WSCONN_MAX_PAYLOAD = 16 * 1024 // 消息最大大小
)

var (
	WSConnReadTimeout = 100 * time.Second
)

// Websocket connection
type WsConn struct {
	StreamConn
	conn *websocket.Conn // websocket conn
}

func NewWsConn(parentCtx context.Context, node fatchoy.NodeID, codecVersion int, conn *websocket.Conn, errChan chan error,
	incoming chan<- fatchoy.IPacket, outsize int, stat *stats.Stats) *WsConn {
	wsconn := &WsConn{
		conn: conn,
	}
	wsconn.StreamConn.init(parentCtx, node, codecVersion, incoming, outsize, errChan, stat)
	wsconn.addr = conn.RemoteAddr().String()
	conn.SetReadLimit(WSCONN_MAX_PAYLOAD)
	conn.SetPingHandler(wsconn.handlePing)
	return wsconn
}

func (c *WsConn) RawConn() net.Conn {
	return c.conn.UnderlyingConn()
}

func (c *WsConn) Go(writer, reader bool) {
	if writer {
		c.wg.Add(1)
		go c.writePump()
	}
	if reader {
		c.wg.Add(1)
		go c.readLoop()
	}
}

func (c *WsConn) SendPacket(pkt fatchoy.IPacket) error {
	if c.IsClosing() {
		return ErrConnIsClosing
	}
	select {
	case c.outbound <- pkt:
		return nil
	default:
		return ErrConnOutboundOverflow
	}
}

func (c *WsConn) Close() error {
	if !atomic.CompareAndSwapInt32(&c.closing, 0, 1) {
		// log.Errorf("WsConn: connection %v is already closed", c.node)
		return nil
	}
	c.cancel()
	c.notifyErr(NewError(ErrConnForceClose, c))
	c.finally()
	return nil
}

func (c *WsConn) ForceClose(err error) {
	if !atomic.CompareAndSwapInt32(&c.closing, 0, 1) {
		// log.Errorf("WsConn: connection %v is already closed", c.node)
		return
	}
	c.cancel()
	c.notifyErr(NewError(err, c))
	go c.finally()
}

func (c *WsConn) finally() {
	c.wg.Wait()
	close(c.outbound)
	c.outbound = nil
	c.inbound = nil
	c.conn = nil
}

func (c *WsConn) writePacket(pkt fatchoy.IPacket) error {
	var buf bytes.Buffer
	var messageType int
	if (pkt.Type() & fatchoy.PacketTypeJSON) > 0 {
		var enc = json.NewEncoder(&buf)
		if err := enc.Encode(pkt); err != nil {
			return err
		}
		messageType = websocket.TextMessage
	} else {
		_, err := codec.Marshal(&buf, pkt, c.encrypt, c.codecVersion)
		if err != nil {
			return err
		}
		messageType = websocket.BinaryMessage
	}
	if err := c.conn.WriteMessage(messageType, buf.Bytes()); err != nil {
		return err
	}
	c.stats.Add(StatPacketsSent, 1)
	c.stats.Add(StatBytesSent, int64(buf.Len()))
	return nil
}

func (c *WsConn) writePump() {
	defer func() {
		c.wg.Done()
		log.Debugf("node %v writer exit", c.node)
	}()

	log.Debugf("node %v writer started at %v", c.node, c.addr)
	for {
		select {
		case pkt, ok := <-c.outbound:
			if !ok {
				return
			}
			if err := c.writePacket(pkt); err != nil {
				log.Errorf("send packet %d: %v", pkt.Command(), err)
			}

		case <-c.ctx.Done():
			return
		}
	}
}

func (c *WsConn) readLoop() {
	for {
		var pkt = packet.Make()
		if err := c.ReadPacket(pkt); err != nil {
			log.Errorf("%v read packet: %v", c.node, err)
			break
		}
		pkt.SetEndpoint(c)
		c.inbound <- pkt

		// check if we should exit
		if c.testShouldExit() {
			return
		}
	}
}

// Exported API
func (c *WsConn) ReadPacket(pkt fatchoy.IPacket) error {
	c.conn.SetReadDeadline(time.Now().Add(WSConnReadTimeout))
	msgType, data, err := c.conn.ReadMessage()
	if err != nil {
		return err
	}

	c.stats.Add(StatPacketsSent, int64(1))
	c.stats.Add(StatBytesSent, int64(len(data)))

	switch msgType {
	case websocket.TextMessage:
		if err := json.Unmarshal(data, pkt); err != nil {
			return err
		}
		pkt.SetType(pkt.Type() | fatchoy.PacketTypeJSON)

	case websocket.BinaryMessage:
		_, err = codec.Unmarshal(bytes.NewReader(data), pkt, c.decrypt)
		pkt.SetType(pkt.Type() | fatchoy.PacketTypeBinary)
		return err

	case websocket.PingMessage, websocket.PongMessage:

	default:
		return fmt.Errorf("unexpected websock message type %d", msgType)
	}
	return nil
}

func (c *WsConn) handlePing(data string) error {
	log.Debugf("ping message: %s", data)
	return nil
}
