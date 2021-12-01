// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"qchen.fun/fatchoy"
	"qchen.fun/fatchoy/codec"
	"qchen.fun/fatchoy/qlog"
	"qchen.fun/fatchoy/x/stats"
)

// Websocket server
type WsServer struct {
	server   *http.Server
	upgrader *websocket.Upgrader  //
	pending  chan *WsConn         // pending connections
	enc      codec.Encoder        // message encode/decode
	errChan  chan error           // error signal
	inbound  chan fatchoy.IPacket // incoming message queue
	outsize  int                  // outgoing queue size
}

func NewWebsocketServer(addr, path string, inbound chan fatchoy.IPacket, outsize int) *WsServer {
	mux := http.NewServeMux()
	var server = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       45 * time.Second,
		MaxHeaderBytes:    4096,
	}
	ws := &WsServer{
		server:  server,
		inbound: inbound,
		outsize: outsize,
		errChan: make(chan error, 32),
		pending: make(chan *WsConn, 128),
		upgrader: &websocket.Upgrader{
			HandshakeTimeout: 10 * time.Second,
			CheckOrigin:      func(r *http.Request) bool { return true }, // allow CORS
		},
	}
	mux.HandleFunc(path, ws.onRequest)
	return ws
}

func (s *WsServer) onRequest(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		qlog.Errorf("WebSocket upgrade %s, %v", r.RemoteAddr, err)
		return
	}
	wsconn := NewWsConn(0, conn, s.enc, s.errChan, s.inbound, s.outsize, stats.New(NumStat))
	qlog.Infof("websocket connection %s established", wsconn.RemoteAddr())
	defer wsconn.Close()
	wsconn.Go(fatchoy.EndpointWriter)
	wsconn.readLoop()
}

func (s *WsServer) BacklogChan() chan *WsConn {
	return s.pending
}

func (s *WsServer) ErrChan() chan error {
	return s.errChan
}

func (s *WsServer) Go() {
	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			qlog.Errorf("ListenAndServe: %v", err)
		}
	}()
}

func (s *WsServer) Shutdown() {
	s.server.Shutdown(context.Background())
	close(s.pending)
	close(s.errChan)
	s.errChan = nil
	s.pending = nil
	s.inbound = nil
	s.server = nil
}
