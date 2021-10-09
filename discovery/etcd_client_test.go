// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package discovery

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"
)

var (
	etcdHostAddr = "127.0.0.1:2379"
	etcdKeyspace = "/choyd"
)

func connectClient(t *testing.T) *Client {
	var client = NewClient(etcdHostAddr, etcdKeyspace)
	if err := client.Init(); err != nil {
		t.Fatalf("connect server: %v", err)
	}
	return client
}

func TestEtcdClient_PutNode(t *testing.T) {
	var client = connectClient(t)
	defer client.Close()

	var node = make(Node)
	node["ID"] = "123"
	node["Type"] = "Bingo"
	var name = "bingo/123"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	if err := client.PutNode(ctx, name, node, 0); err != nil {
		t.Fatalf("set node: %v\n", err)
	}
}

func TestEtcdClient_GetNode(t *testing.T) {
	var client = connectClient(t)
	defer client.Close()

	var name = "bingo/123"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	node, err := client.GetNode(ctx, name)
	if err != nil {
		t.Fatalf("set node: %v\n", err)
	}
	t.Logf("node %s: %v\n", name, node)
}

func TestEtcdClient_IsNodeExist(t *testing.T) {
	var client = connectClient(t)
	defer client.Close()

	var name = "bingo/123"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	found, err := client.IsNodeExist(ctx, name)
	if err != nil {
		t.Fatalf("is exist: %v\n", err)
	}
	t.Logf("node %s exist: %v\n", name, found)
}

func TestEtcdClient_ListDir(t *testing.T) {
	var client = connectClient(t)
	defer client.Close()

	var dir = "service"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	nodes, err := client.ListDir(ctx, dir)
	if err != nil {
		t.Fatalf("list dir %s: %v\n", dir, err)
	}
	t.Logf("%d nodes in dir %s", len(nodes), dir)
	for _, node := range nodes {
		t.Logf("  node: %v", node)
	}
}

func createNode() Node {
	var node = make(Node)
	node["id"] = "123"
	node["type"] = "Bingo"
	node["pid"] = strconv.Itoa(os.Getpid())
	node["ts"] = strconv.Itoa(int(time.Now().Unix()))
	return node
}

func TestEtcdClient_RegisterNode(t *testing.T) {
	var client = connectClient(t)
	defer client.Close()

	var rootCtx = context.Background()
	var leaseId int64
	var err error
	var signal chan struct{}

	var registerDone = false

	var job = func() {
		var node = createNode()
		var name = "bingo/123"
		t.Logf("try to register %s", name)
		leaseId, err = client.RegisterNode(rootCtx, name, node, 5)
		if err != nil {
			t.Logf("register: %v\n", err)
		} else {
			signal, err = client.KeepAlive(rootCtx, leaseId)
			if err != nil {
				t.Logf("keepalive: %v", err)
			} else {
				registerDone = true
				t.Logf("register %s with lease %d done", name, leaseId)
			}
		}
	}

	job()

	var ticker = time.NewTicker(time.Second * 3)
	defer ticker.Stop()
	var ticks = 0
	for {
		select {
		case <-ticker.C:
			ticks++
			println("tick", ticks)
			if ticks >= 20 {
				return
			}
			if !registerDone {
				job()
			}

		case <-signal:
			registerDone = false
			fmt.Printf("lease %d is dead, try re-register later\n", leaseId)
		}
	}
}

func TestEtcdClient_RegisterAndKeepAliveForever(t *testing.T) {
	var client = connectClient(t)
	defer client.Close()

	var rootCtx = context.Background()
	ctx, cancel := context.WithTimeout(rootCtx, time.Second*60)
	defer cancel()
	var node = createNode()
	var name = "bingo/123"
	if err := client.RegisterAndKeepAliveForever(ctx, name, node, 5); err != nil {
		t.Fatalf("register forever: %v", err)
	}
	select {
	case <-ctx.Done():
		break
	}
}

func TestEtcdClient_WatchDir(t *testing.T) {
	var client = connectClient(t)
	defer client.Close()

	var rootCtx = context.Background()
	ctx, cancel := context.WithTimeout(rootCtx, time.Second*600)
	defer cancel()
	eventChan := client.WatchDir(ctx, "service")

	var pollChan = func() bool {
		select {
		case event, ok := <-eventChan:
			if !ok {
				return false
			}
			fmt.Printf("event: %v, key: %s, node: %v\n", event.Type, event.Key, event.Node)
		}
		return true
	}

	for pollChan() {
	}
}

func TestEtcdClient_WatchDirTo(t *testing.T) {
	var client = connectClient(t)
	defer client.Close()

	var rootCtx = context.Background()
	ctx, cancel := context.WithTimeout(rootCtx, time.Second*600)
	defer cancel()

	var nodeMap = NewNodeMap()
	var dir = "service"
	nodes, err := client.ListDir(ctx, dir)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for _, node := range nodes {
		nodeMap.InsertNode(node)
	}

	client.WatchDirTo(ctx, dir, nodeMap)

	var ticker = time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			fmt.Printf("now we have %d nodes\n", nodeMap.Count())
			for _, name := range nodeMap.GetKeys() {
				var nn = nodeMap.GetNodes(name)
				fmt.Printf("  %s count %d\n", name, len(nn))
			}
		}
	}
}
