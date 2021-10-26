// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

// +build !ignore

package qnet

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"gopkg.in/qchencc/fatchoy.v1"
	"gopkg.in/qchencc/fatchoy.v1/codec"
	"gopkg.in/qchencc/fatchoy.v1/packet"
)

func startClient(t *testing.T, addr, path string) {
	time.Sleep(500 * time.Millisecond)
	wurl := url.URL{Scheme: "ws", Host: addr, Path: path}
	c, _, err := websocket.DefaultDialer.Dial(wurl.String(), nil)
	if err != nil {
		t.Fatal("ws dial:", err)
	}
	defer c.Close()

	var pkt = packet.Make()
	pkt.SetCommand(1234)
	pkt.SetSeq(100)
	pkt.SetBodyString("ping")

	data, err := json.Marshal(pkt)
	if err != nil {
		t.Fatalf("marshal %v", err)
	}
	if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Fatalf("write: %v", err)
	}

	var nbytes, msgcnt int
	for i := 0; i < 10; i++ {
		_, msg, err := c.ReadMessage()
		if err != nil {
			t.Fatalf("read %v", err)
		}
		msgcnt += 1
		nbytes += len(msg)
		// fmt.Printf("recv server msg: %s\n", string(msg))
		if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	t.Logf("client recv %d messages, #%d bytes\n", msgcnt, nbytes)
}

func TestWebsocketServer(t *testing.T) {
	var addr = "localhost:9090"
	var path = "/example"
	var incoming = make(chan fatchoy.IPacket, 1000)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()
	server := NewWebsocketServer(ctx, addr, path, codec.VersionV2, incoming, 600)
	server.Go()

	go startClient(t, addr, path)

	var nbytes, msgcnt int
	for {
		select {
		case conn, ok := <-server.BacklogChan():
			if !ok {
				return
			}
			fmt.Printf("connection %s connected\n", conn.RemoteAddr())

		case err := <-server.ErrChan():
			var ne = err.(*Error)
			var endpoint = ne.Endpoint
			fmt.Printf("endpoint[%v] %v closed\n", endpoint.NodeID(), endpoint.RemoteAddr())
			return

		case pkt, ok := <-incoming:
			if !ok {
				return
			}
			msgcnt++
			text := pkt.BodyAsString()
			nbytes += len(text)
			pkt.ReplyString(pkt.Command(), "pong")
			// fmt.Printf("recv client message: %v\n", text))

		case <-ctx.Done():
			return
		}
	}
	t.Logf("server recv %d messages, #%d bytes\n", msgcnt, nbytes)
}
