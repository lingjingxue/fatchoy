// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"gopkg.in/qchencc/fatchoy"
	"gopkg.in/qchencc/fatchoy/codec"
	"gopkg.in/qchencc/fatchoy/log"
	"gopkg.in/qchencc/fatchoy/packet"
	"gopkg.in/qchencc/fatchoy/x/stats"
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

func NewWsConn(parentCtx context.Context, node fatchoy.NodeID, codecVer int, conn *websocket.Conn, errChan chan error,
	incoming chan<- fatchoy.IMessage, outsize int, stat *stats.Stats) *WsConn {
	wsconn := &WsConn{
		conn: conn,
	}
	wsconn.StreamConn.init(parentCtx, node, codecVer, incoming, outsize, errChan, stat)
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
}

func (c *WsConn) SendPacket(pkt fatchoy.IMessage) error {
	if c.IsClosing() {
		return ErrConnIsClosing
	}
	select {
	case c.outbound <- pkt:
		return nil
	default:
		log.Errorf("message %v ignored due to queue overflow", pkt.Command())
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
	c.finally(ErrConnForceClose)
	return nil
}

func (c *WsConn) ForceClose(err error) {
	if !atomic.CompareAndSwapInt32(&c.closing, 0, 1) {
		// log.Errorf("WsConn: connection %v is already closed", c.node)
		return
	}
	c.cancel()
	c.notifyErr(NewError(err, c))
	go c.finally(err)
}

func (c *WsConn) finally(err error) {
	c.wg.Wait()
	if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
		log.Errorf("WsConn: write close message, %v", err)
	}
	if err := c.conn.Close(); err != nil {
		log.Errorf("WsConn: close connection %v, %v", c.node, err)
	}
	close(c.outbound)
	c.outbound = nil
	c.inbound = nil
	c.conn = nil
}

func (c *WsConn) sendPacket(pkt fatchoy.IMessage) error {
	return c.sendBinary(pkt)
}

func (c *WsConn) writePacket(pkt fatchoy.IMessage) error {
	var buf bytes.Buffer
	n, err := codec.Marshal(&buf, pkt, c.encrypt, c.codecVer)
	if err != nil {
		log.Errorf("encode message %v: %v", pkt.Command, err)
		return err
	}
	if err := c.conn.WriteMessage(websocket.BinaryMessage, buf.Bytes()); err != nil {
		log.Errorf("WsConn: send message %v, %v", pkt.Command, err)
		return err
	}
	c.stats.Add(StatPacketsSent, 1)
	c.stats.Add(StatBytesSent, int64(n))
	return nil
}

func (c *WsConn) sendBinary(pkt fatchoy.IMessage) error {
	var buf bytes.Buffer
	n, err := codec.Marshal(&buf, pkt, c.encrypt, c.codecVer)
	if err != nil {
		log.Errorf("encode message %v: %v", pkt.Command, err)
		return err
	}
	if err := c.conn.WriteMessage(websocket.BinaryMessage, buf.Bytes()); err != nil {
		log.Errorf("WsConn: send message %d, %v", pkt.Command, err)
		return err
	}
	c.stats.Add(StatPacketsSent, 1)
	c.stats.Add(StatBytesSent, int64(n))
	return nil
}

func (c *WsConn) writePump() {
	defer c.wg.Done()
	defer log.Debugf("node %v writer exit", c.node)
	log.Debugf("node %v writer started at %v", c.node, c.addr)
	for {
		select {
		case pkt, ok := <-c.outbound:
			if !ok {
				return
			}
			c.sendPacket(pkt)

		case <-c.ctx.Done():
			return
		}
	}
}

func (c *WsConn) readLoop() {
	for {
		var pkt = packet.Make()
		if err := c.ReadPacket(pkt); err != nil {
			log.Errorf("read message: %v", err)
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
func (c *WsConn) ReadPacket(pkt fatchoy.IMessage) error {
	c.conn.SetReadDeadline(fatchoy.Now().Add(WSConnReadTimeout))
	msgType, data, err := c.conn.ReadMessage()
	if err != nil {
		return err
	}

	c.stats.Add(StatPacketsSent, int64(1))
	c.stats.Add(StatBytesSent, int64(len(data)))

	switch msgType {
	case websocket.TextMessage:
		// log.Debugf("recv message: %s", data)
		return json.Unmarshal(data, pkt)

	case websocket.BinaryMessage:
		_, err = codec.Unmarshal(bytes.NewReader(data), pkt, c.decrypt)
		return err

	case websocket.PingMessage, websocket.PongMessage:

	default:
		return errors.Errorf("unexpected websock message type %d", msgType)
	}
	return nil
}

func (c *WsConn) handlePing(data string) error {
	log.Infof("ping message: %s", data)
	return nil
}
