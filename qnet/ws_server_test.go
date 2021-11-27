// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"qchen.fun/fatchoy"
	"qchen.fun/fatchoy/packet"
)

func startWsClient(t *testing.T, addr, path string, msgCount int, ready chan struct{}) {
	// wait until ready
	select {
	case <-ready:
	}

	wurl := url.URL{Scheme: "ws", Host: addr, Path: path}
	c, _, err := websocket.DefaultDialer.Dial(wurl.String(), nil)
	if err != nil {
		t.Fatal("ws dial:", err)
	}
	defer c.Close()

	var pkt = packet.Make()
	pkt.SetCommand(1234)
	pkt.SetSeq(100)
	pkt.SetBody("ping")

	data, err := json.Marshal(pkt)
	if err != nil {
		t.Fatalf("marshal %v", err)
	}
	if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Fatalf("write: %v", err)
	}

	for i := 0; i < msgCount; i++ {
		_, msg, err := c.ReadMessage()
		if err != nil {
			t.Fatalf("read %v", err)
		}
		var resp = packet.Make()
		if err := json.Unmarshal(msg, resp); err != nil {
			t.Fatalf("unmarshal resp: %v", err)
		}
		if s := resp.BodyToString(); s != "pong" {
			t.Fatalf("unexpected response %s", s)
		}
		//fmt.Printf("recv server msg: %s\n", string(msg))
		if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
}

func serveWs(incoming chan fatchoy.IPacket, server *WsServer) {
	var msgcnt int
	const totalMsgNum = maxPingPong * maxConnection
	var timer = time.NewTimer(time.Second * 10)
	for {
		select {
		case endpoint := <-server.BacklogChan():
			fmt.Printf("connection %s connected\n", endpoint.RemoteAddr())
			endpoint.Go(fatchoy.EndpointReadWriter)

		case err := <-server.ErrChan():
			var ne = err.(*Error)
			var endpoint = ne.Endpoint
			fmt.Printf("endpoint[%v] %v closed %v\n", endpoint.NodeID(), endpoint.RemoteAddr(), ne)

		case pkt := <-incoming:
			msgcnt++
			pkt.ReplyWith(pkt.Command(), "pong")
			//fmt.Printf("recv client message: %v\n", text)
			if msgcnt == totalMsgNum {
				return
			}

		case <-timer.C:
			return
		}
	}
}

func TestWebsocketServer(t *testing.T) {
	var addr = "localhost:10009"
	var path = "/ws-test"
	var incoming = make(chan fatchoy.IPacket, 1000)

	server := NewWebsocketServer(addr, path, incoming, 600)
	server.Go()
	var ready = make(chan struct{})
	for i := 0; i < maxConnection; i++ {
		go startWsClient(t, addr, path, maxPingPong, ready)
	}
	ready <- struct{}{}

	serveWs(incoming, server)
}
