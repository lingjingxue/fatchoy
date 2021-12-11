// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"bufio"
	"io"
	"net"
	"time"

	"qchen.fun/fatchoy"
	"qchen.fun/fatchoy/codec"
	"qchen.fun/fatchoy/log"
	"qchen.fun/fatchoy/packet"
	"qchen.fun/fatchoy/x/stats"
)

var (
	TConnReadTimeout = 200
)

// TCP connection
type TcpConn struct {
	StreamConn
	conn   net.Conn // TCP connection object
	reader io.Reader
	writer *bufio.Writer
}

func NewTcpConn(node fatchoy.NodeID, conn net.Conn, enc codec.Encoder, errChan chan error,
	incoming chan<- fatchoy.IPacket, outsize int, stats *stats.Stats) *TcpConn {
	tconn := &TcpConn{
		conn:   conn,
		writer: bufio.NewWriter(conn),
		reader: bufio.NewReader(conn),
	}
	tconn.StreamConn.Init(node, enc, incoming, outsize, errChan, stats)
	tconn.addr = conn.RemoteAddr().String()
	return tconn
}

func (t *TcpConn) RawConn() net.Conn {
	return t.conn
}

func (t *TcpConn) OutboundQueue() chan fatchoy.IPacket {
	return t.outbound
}

func (t *TcpConn) Go(flag fatchoy.EndpointFlag) {
	t.Set(fatchoy.StateRunning)
	if (flag & fatchoy.EndpointWriter) > 0 {
		t.wg.Add(1)
		go t.writePump()
	}
	if (flag & fatchoy.EndpointReader) > 0 {
		t.wg.Add(1)
		go t.readPump()
	}
}

func (t *TcpConn) SendPacket(pkt fatchoy.IPacket) error {
	if !t.IsRunning() {
		return ErrConnIsClosing
	}
	select {
	case t.outbound <- pkt:
		return nil
	default:
		return ErrConnOutboundOverflow
	}
}

func (t *TcpConn) Close() error {
	if !t.CAS(fatchoy.StateRunning, fatchoy.StateShutdown) {
		// log.Errorf("TcpConn: connection %v is already closed", t.node)
		return nil
	}
	if tconn, ok := t.conn.(*net.TCPConn); ok {
		tconn.CloseRead()
	}
	close(t.done)
	t.notifyErr(NewError(ErrConnForceClose, t))
	t.finally() // 阻塞等待投递剩余的消息
	return nil
}

func (t *TcpConn) ForceClose(err error) {
	if !t.CAS(fatchoy.StateRunning, fatchoy.StateShutdown) {
		// log.Errorf("TcpConn: connection %v is already closed", t.node)
		return
	}
	if tconn, ok := t.conn.(*net.TCPConn); ok {
		tconn.CloseRead()
	}
	close(t.done)
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
	t.Set(fatchoy.StateTerminated)
	close(t.outbound)
	t.outbound = nil
	t.inbound = nil
	t.errChan = nil
	t.conn = nil
}

func (t *TcpConn) flush() {
	for i := 0; i < len(t.outbound); i++ {
		select {
		case pkt, ok := <-t.outbound:
			if !ok {
				break
			}
			if err := t.write(pkt); err != nil {
				log.Errorf("%v marshal message %v: %v", t.node, pkt.Command(), err)
			}

		default:
			return
		}
	}
}

func (t *TcpConn) write(pkt fatchoy.IPacket) error {
	nbytes, err := t.enc.WritePacket(t.writer, t.encrypt, pkt)
	if err != nil {
		return err
	}
	if err := t.writer.Flush(); err != nil {
		return err
	}
	t.stats.Add(StatPacketsSent, 1)
	t.stats.Add(StatBytesSent, int64(nbytes))
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
			if err := t.write(pkt); err != nil {
				log.Errorf("%v write message %v: %v", t.node, pkt.Command(), err)
			}

		case <-t.done:
			return
		}
	}
}

func (t *TcpConn) readPacket() (fatchoy.IPacket, error) {
	var deadline = time.Now().Add(time.Duration(TConnReadTimeout) * time.Second)
	t.conn.SetReadDeadline(deadline)
	head, body, err := t.enc.ReadHeadBody(t.reader)
	if err != nil {
		return nil, err
	}
	var pkt = packet.Make()
	if err := t.enc.UnmarshalPacket(head, body, t.decrypt, pkt); err != nil {
		return nil, err
	}
	var nbytes = len(head) + len(body)
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
			if err != io.EOF {
				log.Errorf("%v read packet %v", t.node, err)
			}
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
