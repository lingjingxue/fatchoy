// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"

	"go.etcd.io/etcd/clientv3"
	"gopkg.in/qchencc/fatchoy/log"
)

var (
	ErrEmptyLeasePointer   = errors.New("empty lease pointer")
	ErrNodeKeyAlreadyExist = errors.New("node key already exist")
)

const (
	EventChanCapacity = 1000
)

// 基于etcd的服务发现
type Client struct {
	closing   int32            //
	endpoints []string         // etcd server address
	namespace string           // name space of key
	client    *clientv3.Client // etcd client
}

func NewClient(hostAddr, namespace string) *Client {
	d := &Client{
		endpoints: strings.Split(hostAddr, ","),
		namespace: namespace,
	}
	return d
}

func (c *Client) Init() error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   c.endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}
	c.client = cli
	return nil
}

func (c *Client) Close() {
	if !atomic.CompareAndSwapInt32(&c.closing, 0, 1) {
		return
	}
	if c.client != nil {
		c.client.Close()
		c.client = nil
	}
}

func (c *Client) IsClosing() bool {
	return atomic.LoadInt32(&c.closing) == 1
}

func (c *Client) formatKey(name string) string {
	return fmt.Sprintf("%s/%s", c.namespace, name)
}

// 节点是否存在
func (c *Client) IsNodeExist(ctx context.Context, name string) (bool, error) {
	var key = c.formatKey(name)
	resp, err := c.client.Get(ctx, key, clientv3.WithCountOnly())
	if err != nil {
		return false, err
	}
	return resp.Count > 0, nil
}

// 获取节点信息
func (c *Client) GetNode(ctx context.Context, name string) (Node, error) {
	var key = c.formatKey(name)
	resp, err := c.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if resp.Count == 0 {
		return nil, nil
	}
	var node Node
	if err := json.Unmarshal(resp.Kvs[0].Value, &node); err != nil {
		return nil, err
	}
	return node, nil
}

// 设置节点信息
func (c *Client) PutNode(ctx context.Context, name string, value interface{}, leaseId int64) error {
	var key = c.formatKey(name)
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var resp *clientv3.PutResponse
	if leaseId <= 0 {
		resp, err = c.client.Put(ctx, key, bytesAsString(data))
	} else {
		resp, err = c.client.Put(ctx, key, bytesAsString(data), clientv3.WithLease(clientv3.LeaseID(leaseId)))
	}
	if err != nil {
		return err
	}
	log.Debugf("put key [%s] at rev %d", key, resp.Header.Revision)
	return nil
}

// 列出目录下的所有节点
func (c *Client) ListDir(ctx context.Context, dir string) ([]Node, error) {
	var key = c.formatKey(dir)
	resp, err := c.client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if resp.Count == 0 {
		return nil, nil
	}
	var nodes = make([]Node, 0, resp.Count)
	for _, kv := range resp.Kvs {
		var node Node
		if err := json.Unmarshal(kv.Value, &node); err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// 申请一个lease
func (c *Client) GrantLease(ctx context.Context, ttl int) (int64, error) {
	lease, err := c.client.Grant(ctx, int64(ttl))
	if err != nil {
		return 0, err
	}
	if lease == nil {
		return 0, ErrEmptyLeasePointer
	}
	return int64(lease.ID), nil
}

func (c *Client) GetLeaseTTL(ctx context.Context, leaseId int64) (int, error) {
	resp, err := c.client.TimeToLive(ctx, clientv3.LeaseID(leaseId))
	if err != nil {
		return 0, nil
	}
	return int(resp.TTL), nil
}

// 撤销一个lease
func (c *Client) RevokeLease(ctx context.Context, leaseId int64) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	_, err := c.client.Revoke(ctx, clientv3.LeaseID(leaseId))
	return err
}

// 注册一个节点信息，并返回一个ttl秒的lease
func (c *Client) RegisterNode(rootCtx context.Context, name string, value interface{}, ttl int) (int64, error) {
	exist, err := c.IsNodeExist(rootCtx, name)
	if err != nil {
		return 0, err
	}
	if exist {
		return 0, ErrNodeKeyAlreadyExist
	}

	ctx, cancel := context.WithTimeout(rootCtx, time.Second*3)
	defer cancel()

	var leaseId int64
	if ttl > 0 {
		if leaseId, err = c.GrantLease(ctx, ttl); err != nil {
			return 0, err
		}
	}
	if err = c.PutNode(ctx, name, value, leaseId); err != nil {
		return 0, err
	}
	return leaseId, nil
}

func doRevoke(c *Client, ctx context.Context, leaseId int64) {
	log.Debugf("context canceled, try revoke lease %d", leaseId)
	revokeCtx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()
	if err := c.RevokeLease(revokeCtx, leaseId); err != nil {
		log.Debugf("revoke lease %x: %v", leaseId, err)
	} else {
		log.Debugf("revoke lease %x done", leaseId)
	}
}

// lease保活，返回一个channel，当lease撤销时此channel被激活
func (c *Client) KeepAlive(ctx context.Context, leaseId int64) (chan struct{}, error) {
	kaChan, err := c.client.KeepAlive(ctx, clientv3.LeaseID(leaseId))
	if err != nil {
		return nil, err
	}
	var signal = make(chan struct{})
	go func() {
		defer func() {
			signal <- struct{}{} // notify signal
		}()
		for {
			select {
			case ka, ok := <-kaChan:
				if !ok || ka == nil {
					log.Debugf("lease %x is not alive", leaseId)
					return
				}
				// log.Debugf("lease %d respond alive, ttl %d", ka.ID, ka.TTL)

			case <-ctx.Done():
				doRevoke(c, context.Background(), leaseId)
				return
			}
		}
	}()
	return signal, nil
}

// 注册一个节点，并永久保活
func (c *Client) RegisterAndKeepAliveForever(rootCtx context.Context, name string, value interface{}, ttl int) error {
	var leaseId int64
	var signal chan struct{}
	var leaseAlive bool

	var doJob = func() error {
		var err error
		log.Debugf("try to register key: %s", name)
		leaseAlive = false
		leaseId, err = c.RegisterNode(rootCtx, name, value, ttl)
		if err != nil {
			return err
		}
		signal, err = c.KeepAlive(rootCtx, leaseId)
		if err != nil {
			return err
		}
		leaseAlive = true
		log.Debugf("register key [%s] with lease %x done", name, leaseId)
		return nil
	}

	if err := doJob(); err != nil {
		return err
	}

	go func() {
		var ticker = time.NewTicker(time.Second * 3)
		defer ticker.Stop()
		for {
			select {
			case <-rootCtx.Done():
				return

			case <-ticker.C:
				if !leaseAlive {
					if err := doJob(); err != nil {
						log.Warnf("register and keepalive %s: %v", name, err)
					}
				}

			case <-signal:
				leaseAlive = false
				log.Debugf("lease %x of node %s is not alive, try register later", leaseId, name)
			}
		}
	}()
	return nil
}

func propagateWatchEvent(eventChan chan<- *NodeEvent, resp *clientv3.WatchResponse) {
	for _, ev := range resp.Events {
		var event = &NodeEvent{
			Type: EventUnknown,
			Key:  string(ev.Kv.Key),
		}
		switch ev.Type {
		case 0: // PUT
			if ev.IsCreate() {
				event.Type = EventCreate
			} else {
				event.Type = EventUpdate
			}
		case 1: // DELETE
			event.Type = EventDelete
		}
		if ev.Kv.Value != nil {
			if err := json.Unmarshal(ev.Kv.Value, &event.Node); err != nil {
				log.Debugf("unmarshal node %s: %v", event.Key, err)
				continue
			}
		}
		eventChan <- event
	}
}

// 订阅目录下的节点变化
func (c *Client) WatchDir(ctx context.Context, dir string) <-chan *NodeEvent {
	var key = c.formatKey(dir)
	watchCh := c.client.Watch(clientv3.WithRequireLeader(ctx), key, clientv3.WithPrefix())
	eventChan := make(chan *NodeEvent, EventChanCapacity)
	go func() {
		defer close(eventChan)
		for {
			select {
			case resp, ok := <-watchCh:
				if !ok {
					return
				}
				if resp.Err() != nil {
					log.Warnf("watch of key %s canceled: %v", key, resp.Err())
					return
				}
				propagateWatchEvent(eventChan, &resp)

			case <-ctx.Done():
				if err := c.client.Watcher.Close(); err != nil {
					log.Warnf("close watcher: %v", err)
				}
				return
			}
		}
	}()
	return eventChan
}

// 订阅目录下的所有节点变化, 并把节点变化更新到nodeMap
func (c *Client) WatchDirTo(ctx context.Context, dir string, nodeMap *NodeMap) {
	var evChan = c.WatchDir(ctx, dir)
	var prefix = c.formatKey(dir)
	go func() {
		for {
			select {
			case ev, ok := <-evChan:
				if !ok {
					return
				}
				updateNodeEvent(nodeMap, prefix, ev)
			}
		}
	}()
}

func updateNodeEvent(nodeMap *NodeMap, rootDir string, ev *NodeEvent) {
	switch ev.Type {
	case EventCreate:
		nodeMap.InsertNode(ev.Node)
	case EventUpdate:
		nodeMap.InsertNode(ev.Node) // 插入前会先检查是否有重复
	case EventDelete:
		nodeType, id := parseNodeTypeAndID(rootDir, ev.Key)
		if nodeType != "" && id > 0 {
			nodeMap.DeleteNode(nodeType, id)
		}
	}
}

func parseNodeTypeAndID(root, key string) (string, uint16) {
	idx := strings.Index(key, root)
	if idx < 0 {
		return "", 0
	}
	key = key[len(root)+1:] // root + '/' + key
	idx = strings.Index(key, "/")
	if idx <= 0 {
		return "", 0
	}
	var nodeType = key[:idx]
	var strId = key[idx+1:]
	id, _ := strconv.Atoi(strId)
	return nodeType, uint16(id)
}

func bytesAsString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
