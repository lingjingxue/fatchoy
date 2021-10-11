// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"bufio"
	"context"
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/qchencc/fatchoy"
	"gopkg.in/qchencc/fatchoy/codec"
	"gopkg.in/qchencc/fatchoy/log"
	"gopkg.in/qchencc/fatchoy/packet"
	"gopkg.in/qchencc/fatchoy/x/stats"
)

var (
	TConnReadTimeout = 200
)

// TCP connection
type TcpConn struct {
	StreamConn
	conn   net.Conn      // TCP connection object
	reader *bufio.Reader // buffered reader
	writer *bufio.Writer // buffered writer
}

func NewTcpConn(parentCtx context.Context, node fatchoy.NodeID, codecVer int, conn net.Conn, errChan chan error,
	incoming chan<- fatchoy.IMessage, outsize int, stats *stats.Stats) *TcpConn {
	tconn := &TcpConn{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}
	tconn.StreamConn.init(parentCtx, node, codecVer, incoming, outsize, errChan, stats)
	tconn.addr = conn.RemoteAddr().String()
	return tconn
}

func (t *TcpConn) RawConn() net.Conn {
	return t.conn
}

func (t *TcpConn) OutboundQueue() chan fatchoy.IMessage {
	return t.outbound
}

func (t *TcpConn) Go(writer, reader bool) {
	if writer {
		t.wg.Add(1)
		go t.writePump()
	}
	if reader {
		t.wg.Add(1)
		go t.readPump()
	}
}

func (t *TcpConn) SendPacket(pkt fatchoy.IMessage) error {
	if t.IsClosing() {
		return ErrConnIsClosing
	}
	select {
	case t.outbound <- pkt:
		return nil
	default:
		log.Errorf("TcpConn: message %v ignored due to queue overflow", pkt.Command())
		return errors.WithStack(ErrConnOutboundOverflow)
	}
}

func (t *TcpConn) Close() error {
	if !atomic.CompareAndSwapInt32(&t.closing, 0, 1) {
		// log.Errorf("TcpConn: connection %v is already closed", t.node)
		return nil
	}
	if tconn, ok := t.conn.(*net.TCPConn); ok {
		tconn.CloseRead()
	}
	t.cancel()
	t.notifyErr(NewError(ErrConnForceClose, t))
	t.finally() // 阻塞等待投递剩余的消息
	return nil
}

func (t *TcpConn) ForceClose(err error) {
	if !atomic.CompareAndSwapInt32(&t.closing, 0, 1) {
		// log.Errorf("TcpConn: connection %v is already closed", t.node)
		return
	}
	if tconn, ok := t.conn.(*net.TCPConn); ok {
		tconn.CloseRead()
	}
	t.cancel()
	t.notifyErr(NewError(err, t))
	go t.finally() // 不阻塞等待
}

func (t *TcpConn) finally() {
	t.wg.Wait()
	if tconn, ok := t.conn.(*net.TCPConn); ok {
		tconn.CloseWrite()
	} else {
		t.conn.Close()
	}

	close(t.outbound)
	t.outbound = nil
	t.inbound = nil
	t.errChan = nil
	t.conn = nil
	t.reader = nil
}

func (t *TcpConn) flush() {
	for i := 0; i < len(t.outbound); i++ {
		select {
		case pkt, ok := <-t.outbound:
			if !ok {
				break
			}
			if err := t.writePacket(pkt); err != nil {
				log.Errorf("write message %v: %v", pkt.Command(), err)
			}

		default:
			return
		}
	}
}

func (t *TcpConn) writePacket(pkt fatchoy.IMessage) error {
	n, err := codec.Marshal(t.writer, pkt, t.encrypt, t.codecVer)
	if err != nil {
		return err
	}
	if err := t.writer.Flush(); err != nil {
		return err
	}
	t.stats.Add(StatPacketsSent, 1)
	t.stats.Add(StatBytesSent, int64(n))
	return nil
}

func (t *TcpConn) writePump() {
	defer func() {
		t.flush()
		t.wg.Done()
		log.Debugf("TcpConn: node %v writer stopped", t.node)
	}()

	log.Debugf("TcpConn: node %v(%v) writer started", t.node, t.addr)

	for {
		select {
		case pkt, ok := <-t.outbound:
			if !ok {
				return
			}
			if err := t.writePacket(pkt); err != nil {
				log.Errorf("write message %v: %v", pkt.Command, err)
			}

		case <-t.ctx.Done():
			return
		}
	}
}

func (t *TcpConn) readPacket() (fatchoy.IMessage, error) {
	var deadline = time.Now().Add(time.Duration(TConnReadTimeout) * time.Second)
	t.conn.SetReadDeadline(deadline)
	var pkt = packet.Make()
	nbytes, err := codec.Unmarshal(t.reader, pkt, t.decrypt)
	if err != nil {
		if err != io.EOF {
			log.Errorf("read message from node %v: %v", t.node, err)
		}
		return pkt, err
	}
	t.stats.Add(StatPacketsRecv, 1)
	t.stats.Add(StatBytesRecv, int64(nbytes))
	pkt.SetEndpoint(t)
	return pkt, nil
}

func (t *TcpConn) readPump() {
	defer func() {
		t.wg.Done()
		log.Debugf("TcpConn: node %v reader stopped", t.node)
	}()
	log.Debugf("TcpConn: node %v(%v) reader started", t.node, t.addr)
	for {
		pkt, err := t.readPacket()
		if err != nil {
			t.ForceClose(err) // I/O超时或者发生错误，强制关闭连接
			return
		}
		t.inbound <- pkt // 如果channel满了，这里会阻塞

		// test if we should exit
		if t.testShouldExit() {
			return
		}
	}
}
