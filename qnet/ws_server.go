// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"gopkg.in/qchencc/fatchoy.v1"
	"gopkg.in/qchencc/fatchoy.v1/codec"
	"gopkg.in/qchencc/fatchoy.v1/log"
	"gopkg.in/qchencc/fatchoy.v1/x/stats"
)

// Websocket server
type WsServer struct {
	ctx      context.Context
	cancel   context.CancelFunc
	server   *http.Server
	upgrader *websocket.Upgrader  //
	pending  chan *WsConn         //
	errChan  chan error           //
	inbound  chan fatchoy.IPacket // incoming message queue
	version  codec.Version        // codec version
	outsize  int                  // outgoing queue size
}

func NewWebsocketServer(parentCtx context.Context, addr, path string, version codec.Version, inbound chan fatchoy.IPacket, outsize int) *WsServer {
	ctx, cancel := context.WithCancel(parentCtx)
	mux := http.NewServeMux()
	var server = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       45 * time.Second,
		MaxHeaderBytes:    4096,
		BaseContext:       func(listener net.Listener) context.Context { return ctx },
	}
	ws := &WsServer{
		ctx:     ctx,
		cancel:  cancel,
		server:  server,
		inbound: inbound,
		version: version,
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
		log.Errorf("WebSocket upgrade %s, %v", r.RemoteAddr, err)
		return
	}
	wsconn := NewWsConn(r.Context(), 0, s.version, conn, s.errChan, s.inbound, s.outsize, stats.New(NumStat))
	log.Infof("websocket connection %s established", wsconn.RemoteAddr())
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
			log.Errorf("ListenAndServe: %v", err)
		}
	}()
}

func (s *WsServer) Shutdown() {
	s.server.Shutdown(s.ctx)
	s.cancel()
	close(s.pending)
	close(s.errChan)
	s.errChan = nil
	s.pending = nil
	s.inbound = nil
	s.server = nil
}
