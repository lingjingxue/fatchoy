// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

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
)

//不断发送ping接收pong
func startRawClient(t *testing.T, id int, address string, msgCount int) {
	//t.Logf("client %d start connect %s", id, address)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("Dial %s: %v", address, err)
	}
	defer conn.Close()

	var pkt = packet.Make()
	for i := 1; i <= msgCount; i++ {
		pkt.SetCommand(int32(i))
		pkt.SetSeq(int16(i))
		pkt.SetBodyString("ping")
		buf, err := codec.MarshalV2(pkt, nil)
		if err != nil {
			t.Fatalf("Encode: %v", err)
		}
		if _, err := conn.Write(buf); err != nil {
			t.Fatalf("Write: %v", err)
		}
		var resp = packet.Make()
		if err := codec.ReadPacketV2(conn, nil, resp); err != nil {
			t.Fatalf("Decode: %v", err)
		}
		if resp.Seq() != pkt.Seq() {
			t.Fatalf("session mismatch, %d != %d", resp.Seq(), pkt.Seq())
		}
		s := resp.BodyToString()
		if s != "pong" {
			t.Fatalf("invalid response message: %s", s)
		}
	}
	//fmt.Printf("Connection %v done\n", conn.RemoteAddr())
}

func startServeRawClient(t *testing.T, ctx context.Context, cancel context.CancelFunc, address string, ready chan struct{}) {
	var incoming = make(chan fatchoy.IPacket, 1000)
	var server = NewTcpServer(context.Background(), incoming, 100)
	if err := server.Listen(address); err != nil {
		t.Fatalf("Listen: %s %v", address, err)
	}

	ready <- struct{}{} // listen ready

	var autoId uint32
	var recvNum = 0

	const totalMsgNum = maxPingPong * maxConnection

	for {
		select {
		case endpoint := <-server.BacklogChan():
			//var addr = endpoint.RemoteAddr()
			//fmt.Printf("endpoint %v connected\n", addr)
			autoId++
			endpoint.SetNodeID(fatchoy.NodeID(autoId))
			endpoint.Go(fatchoy.EndpointReadWriter)

		case err := <-server.ErrorChan():
			// handle connection error
			var ne = err.(*Error)
			var endpoint = ne.Endpoint
			// fmt.Printf("endpoint[%v] %v closed\n", endpoint.Node(), endpoint.RemoteAddr())
			if !endpoint.IsClosing() {
				endpoint.Close()
			}

		case pkt := <-incoming:
			//println("recv", pkt.BodyToString())
			pkt.ReplyString(pkt.Command(), "pong") //返回pong

			// all message recv, close server
			recvNum++
			if recvNum > 0 && recvNum%100 == 0 {
				//fmt.Printf("recv messages: %d/%d\n", recvNum, totalMsgNum)
			}
			if recvNum == totalMsgNum {
				fmt.Printf("all messages recv OK, shutdown\n")
				cancel()
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

func TestExampleServerUsage(t *testing.T) {
	var addr = "localhost:10004"
	var ready = make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go startServeRawClient(t, ctx, cancel, addr, ready)

	<-ready // wait listen ready
	t.Logf("server listen OK")

	// start client connections
	for i := 0; i < maxConnection; i++ {
		time.Sleep(10 * time.Millisecond)
		go startRawClient(t, i+1, addr, maxPingPong)
	}

	var timer = time.NewTimer(time.Minute) // this case should pass no more than 1 minute
	select {
	case <-timer.C:
		cancel()
		fmt.Printf("timeout to end\n")
	case <-ctx.Done():
	}
}
