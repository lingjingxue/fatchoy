// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

// +build !ignore

package qnet

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"gopkg.in/qchencc/fatchoy.v1"
	"gopkg.in/qchencc/fatchoy.v1/codec"
	"gopkg.in/qchencc/fatchoy.v1/packet"
	"gopkg.in/qchencc/fatchoy.v1/x/strutil"
)

const (
	maxConnection = 100
	maxPingpong   = 1000
)

func handleConn(conn net.Conn) {
	var count = 0
	var ctx = context.Background()
	tconn := NewTcpConn(ctx, 0, codec.VersionV2, conn, nil, nil, 1000, nil)
	tconn.Go(true, false)
	defer tconn.Close()
	for {
		conn.SetReadDeadline(time.Now().Add(time.Minute))
		var pkt = packet.Make()
		if _, err := codec.Unmarshal(conn, pkt, nil); err != nil {
			fmt.Printf("Decode: %v\n", err)
			break
		}

		// fmt.Printf("%d srecv: %s\n", file.Fd(), pkt.Body)
		pkt.SetBodyString(fmt.Sprintf("pong %d", pkt.Command()))
		tconn.SendPacket(pkt)
		//fmt.Printf("message %d OK\n", count)
		count++
		if count == maxPingpong {
			break
		}
	}
	stats := tconn.Stats()
	fmt.Printf("sent %d packets, %s\n", stats.Get(StatPacketsSent), strutil.PrettyBytes(stats.Get(StatBytesSent)))
}

func startMyServer(t *testing.T, ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			//t.Logf("Listener: Accept %v", err)
			return
		}
		go handleConn(conn)
	}
}

func tconnReadLoop(errchan chan error, inbound chan fatchoy.IPacket) {
	for {
		select {
		case pkt, ok := <-inbound:
			if !ok {
				return
			}
			pkt.SetCommand(pkt.Command() + 1)
			pkt.ReplyString(pkt.Command(), fmt.Sprintf("ping %d", pkt.Command()))

		case <-errchan:
			return
		}
	}
}

func TestExampleTcpConn(t *testing.T) {
	TConnReadTimeout = 30

	var testTcpAddress = "localhost:10002"

	ln, err := net.Listen("tcp", testTcpAddress)
	if err != nil {
		t.Fatalf("Listen %v", err)
	}
	defer ln.Close()

	go startMyServer(t, ln)

	conn, err := net.Dial("tcp", testTcpAddress)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	//file, _ := conn.File()
	inbound := make(chan fatchoy.IPacket, 1000)
	errchan := make(chan error, 4)
	tconn := NewTcpConn(context.Background(), 0, codec.VersionV2, conn, errchan, inbound, 1000, nil)
	tconn.SetNodeID(fatchoy.NodeID(0x12345))
	tconn.Go(true, true)
	defer tconn.Close()
	stats := tconn.Stats()
	var pkt = packet.Make()
	pkt.SetCommand(1)
	pkt.SetBodyString("ping")
	tconn.SendPacket(pkt)
	tconnReadLoop(errchan, inbound)
	fmt.Printf("recv %d packets, %s\n", stats.Get(StatPacketsRecv), strutil.PrettyBytes(stats.Get(StatBytesRecv)))
}
