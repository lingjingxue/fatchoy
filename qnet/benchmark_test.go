// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"qchen.fun/fatchoy"
	"qchen.fun/fatchoy/codec"
	"qchen.fun/fatchoy/packet"
	"qchen.fun/fatchoy/x/datetime"
)

const (
	benchConnCount     = 100
	totalBenchMsgCount = 1000000
)

func serveBench(t *testing.T, ctx context.Context, addr string, ready chan struct{}) {
	var incoming = make(chan fatchoy.IPacket, totalBenchMsgCount)
	var listener = NewTcpServer(context.Background(), incoming, totalBenchMsgCount)
	if err := listener.Listen(addr); err != nil {
		t.Fatalf("Listen: %s %v", addr, err)
	}

	ready <- struct{}{} // server listen ready
	var autoId int32 = 1

	for {
		select {
		case endpoint := <-listener.BacklogChan():
			endpoint.SetNodeID(fatchoy.NodeID(autoId))
			endpoint.Go(fatchoy.EndpointReadWriter)
			autoId++

		case err := <-listener.ErrorChan():
			// handle connection error
			var ne = err.(*Error)
			var endpoint = ne.Endpoint
			if !endpoint.IsClosing() {
				endpoint.Close()
			}

		case pkt := <-incoming:
			pkt.ReplyWith(pkt.Command(), "pong") //返回pong

		case <-ctx.Done():
			// handle shutdown
			return
		}
	}
}

func startBenchClient(t *testing.T, address string, msgCount int, ready chan struct{}, respChan chan int) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("Dial %s: %v", address, err)
	}

	// wait until ready
	select {
	case <-ready:
	}

	for i := 0; i < msgCount; i++ {
		var buf bytes.Buffer
		var pkt = packet.New(int32(i), 0, 0, "ping")
		if _, err := codec.MarshalV2(&buf, pkt, nil); err != nil {
			t.Fatalf("Encode: %v", err)
		}
		if _, err := buf.WriteTo(conn); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}
	for i := 0; i < msgCount; i++ {
		var resp = packet.Make()
		if err := codec.ReadPacketV2(conn, nil, resp); err != nil {
			t.Fatalf("Decode: %v", err)
		}
		respChan <- 1
	}
}

func TestQPSBenchmark(t *testing.T) {
	var address = "localhost:10001"
	const eachConnectMsgCount = totalBenchMsgCount / benchConnCount

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var ready = make(chan struct{})
	go serveBench(t, ctx, address, ready)
	<-ready // listener ready

	var respChan = make(chan int, totalBenchMsgCount)
	for i := 0; i < benchConnCount; i++ {
		go startBenchClient(t, address, eachConnectMsgCount, ready, respChan)
	}

	ready <- struct{}{} // all client ready
	var timer = time.NewTimer(time.Second * 10)
	fmt.Printf("start QPS benchmark %s\n", datetime.FormatTime(time.Now()))
	var startTime = time.Now()

	var cnt = 0
	for {
		select {
		case <-timer.C:
			goto LabelDone
		case <-respChan:
			cnt++
			if cnt == totalBenchMsgCount {
				goto LabelDone
			}
		default:
		}
	}

LabelDone:
	var now = time.Now()
	cancel()
	fmt.Printf("QPS benchmark finished %s\n", datetime.FormatTime(now))
	var elapsed = now.Sub(startTime)
	var qps = float64(totalBenchMsgCount) / (float64(elapsed) / float64(time.Second))
	fmt.Printf("Send %d message with %d clients cost %v, QPS: %f\n", totalBenchMsgCount, benchConnCount, elapsed, qps)

	fmt.Printf("Benchmark finished\n")

	// Output:
	// 	QPS: 100206.900193
}
