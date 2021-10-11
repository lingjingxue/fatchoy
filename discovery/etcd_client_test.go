// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package discovery

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"gopkg.in/qchencc/fatchoy.v1/log"
)

var (
	etcdHostAddr = "127.0.0.1:2379"
	etcdKeyspace = "/choyd-test"
	nodeId       = strconv.Itoa(rand.Int() % 100000)
)

func init() {
	rand.Seed(time.Now().UnixNano())
	log.Setup(log.NewConfig("debug"))
}

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
	node["ID"] = nodeId
	node["Type"] = "Bingo"
	var name = "bingo/" + nodeId
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	if err := client.PutNode(ctx, name, node, 0); err != nil {
		t.Fatalf("set node: %v\n", err)
	}
}

func TestEtcdClient_GetNode(t *testing.T) {
	var client = connectClient(t)
	defer client.Close()

	var name = "bingo/" + nodeId
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

	var name = "bingo/" + nodeId
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	found, err := client.IsNodeExist(ctx, name)
	if err != nil {
		t.Fatalf("is exist: %v\n", err)
	}
	t.Logf("node %s exist: %v\n", name, found)

	if err := client.DelKey(ctx, name); err != nil {
		t.Fatalf("delete node: %v\n", err)
	}
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
	node["id"] = nodeId
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
		var name = "bingo/" + nodeId
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
			fmt.Printf("RegisterNode re-register worker tick %d, in case of etcd server lost\n", ticks)
			if ticks >= 10 {
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
	ctx, cancel := context.WithTimeout(rootCtx, time.Second*30)
	defer cancel()
	var node = createNode()
	var name = "bingo/" + nodeId
	t.Logf("register and keepalive forever, only for 30s")
	if err := client.RegisterAndKeepAliveForever(ctx, name, node, 5); err != nil {
		t.Fatalf("register forever: %v", err)
	}
	select {
	case <-ctx.Done():
		break
	}
	t.Logf("done")
}

func TestEtcdClient_WatchDir(t *testing.T) {
	var client = connectClient(t)
	defer client.Close()

	var rootCtx = context.Background()
	ctx, cancel := context.WithTimeout(rootCtx, time.Second*60)
	defer cancel()

	var dir = "service"
	eventChan := client.WatchDir(ctx, dir)

	t.Logf("watch key %s/%s for 60s, you can add/delete some key by etcdctl", etcdKeyspace, dir)
	for {
		select {
		case event, ok := <-eventChan:
			if !ok {
				return
			}
			fmt.Printf("event: %v, key: %s, node: %v\n", event.Type, event.Key, event.Node)
		}
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

	// list all nodes, and insert to map
	nodes, err := client.ListDir(ctx, dir)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for _, node := range nodes {
		nodeMap.InsertNode(node)
	}

	t.Logf("watch key %s/%s for 60s, you can add/delete some key by etcdctl", etcdKeyspace, dir)
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
