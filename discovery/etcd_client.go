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
	"gopkg.in/qchencc/fatchoy.v1/qlog"
)

var (
	ErrEmptyLeasePointer   = errors.New("empty lease pointer")
	ErrNodeKeyAlreadyExist = errors.New("node key already exist")
	ErrNoKeyDeleted        = errors.New("no key deleted")
)

const (
	EventChanCapacity = 1000
	OpTimeout         = 3

	VerboseLv1 = 1
	VerboseLv2 = 2
)

// 基于etcd的服务发现
type Client struct {
	closing   int32            //
	verbose   int32            //
	endpoints []string         // etcd server address
	namespace string           // name space of key
	client    *clientv3.Client // etcd client
}

func NewClient(hostAddr, namespace string) *Client {
	d := &Client{
		endpoints: strings.Split(hostAddr, ","),
		namespace: namespace,
		verbose:   VerboseLv1,
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

func (c *Client) SetVerbose(v int32) {
	c.verbose = v
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
	if c.verbose >= VerboseLv1 {
		qlog.Infof("put key [%s] at rev %d", key, resp.Header.Revision)
	}
	return nil
}

// 删除一个key
func (c *Client) DelKey(ctx context.Context, name string) error {
	var key = c.formatKey(name)
	resp, err := c.client.Delete(ctx, key)
	if err != nil {
		return err
	}
	if resp.Deleted == 0 {
		return ErrNoKeyDeleted
	}
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
	_, err := c.client.Revoke(ctx, clientv3.LeaseID(leaseId))
	return err
}

// 注册一个节点信息，并返回一个ttl秒的lease
func (c *Client) RegisterNode(rootCtx context.Context, name string, value interface{}, ttl int) (int64, error) {
	ctx, cancel := context.WithTimeout(rootCtx, time.Second*OpTimeout)
	defer cancel()

	exist, err := c.IsNodeExist(ctx, name)
	if err != nil {
		return 0, err
	}
	if exist {
		return 0, ErrNodeKeyAlreadyExist
	}
	var leaseId int64
	if ttl <= 0 {
		ttl = 5
	}
	if leaseId, err = c.GrantLease(ctx, ttl); err != nil {
		return 0, err
	}
	if err = c.PutNode(ctx, name, value, leaseId); err != nil {
		return 0, err
	}
	return leaseId, nil
}

func revokeLeaseWithTimeout(c *Client, leaseId int64) {
	if c.verbose >= VerboseLv1 {
		qlog.Infof("try revoke lease %d", leaseId)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*OpTimeout)
	defer cancel()
	if err := c.RevokeLease(ctx, leaseId); err != nil {
		qlog.Warnf("revoke lease %x: %v", leaseId, err)
	} else {
		qlog.Infof("revoke lease %x done", leaseId)
	}
}

// lease保活，返回一个channel，当lease撤销时此channel被激活
func (c *Client) KeepAlive(ctx context.Context, leaseId int64) (chan struct{}, error) {
	kaChan, err := c.client.KeepAlive(ctx, clientv3.LeaseID(leaseId))
	if err != nil {
		return nil, err
	}
	var signal = make(chan struct{})
	var aliveKeeper = func() {
		defer func() {
			signal <- struct{}{} // notify signal
			revokeLeaseWithTimeout(c, leaseId)
		}()
		for {
			select {
			case ka, ok := <-kaChan:
				if !ok || ka == nil {
					qlog.Infof("lease %x is not alive", leaseId)
					return
				}
				if c.verbose >= VerboseLv2 {
					qlog.Infof("lease %d respond alive, ttl %d", ka.ID, ka.TTL)
				}

			case <-ctx.Done():
				return
			}
		}
	}
	go aliveKeeper()
	return signal, nil
}

// 注册一个节点，并永久保活
func (c *Client) RegisterAndKeepAliveForever(ctx context.Context, name string, value interface{}, ttl int) error {
	var leaseId int64
	var done chan struct{}
	var leaseAlive bool

	var doRegister = func() error {
		var err error
		if c.verbose >= VerboseLv1 {
			qlog.Infof("try register key: %s", name)
		}
		leaseAlive = false
		leaseId = 0

		leaseId, err = c.RegisterNode(ctx, name, value, ttl)
		if err != nil {
			return err
		}
		done, err = c.KeepAlive(ctx, leaseId)
		if err != nil {
			return err
		}
		leaseAlive = true
		if c.verbose >= VerboseLv1 {
			qlog.Infof("register key [%s] with lease %x done", name, leaseId)
		}
		return nil
	}

	if err := doRegister(); err != nil {
		return err
	}

	var aliveKeeper = func() {
		var ticker = time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if !leaseAlive {
					if err := doRegister(); err != nil {
						qlog.Infof("register or keepalive %s failed: %v", name, err)
					}
				}

			case <-done:
				if c.verbose >= VerboseLv1 {
					qlog.Infof("node %s lease(%d) is not alive, try register later", name, leaseId)
				}
				leaseAlive = false
				leaseId = 0

			case <-ctx.Done():
				return
			}
		}
	}
	go aliveKeeper()
	return nil
}

func propagateWatchEvent(eventChan chan<- *NodeEvent, ev *clientv3.Event) {
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
	if len(ev.Kv.Value) > 0 {
		if err := json.Unmarshal(ev.Kv.Value, &event.Node); err != nil {
			qlog.Errorf("unmarshal node %s: %v", event.Key, err)
			return
		}
	}

	select {
	case eventChan <- event:
	default:
		qlog.Warnf("watch event channel is full, new event lost: %v", event)
	}
}

// 订阅目录下的节点变化
func (c *Client) WatchDir(ctx context.Context, dir string) <-chan *NodeEvent {
	var key = c.formatKey(dir)
	watchCh := c.client.Watch(clientv3.WithRequireLeader(ctx), key, clientv3.WithPrefix())
	eventChan := make(chan *NodeEvent, EventChanCapacity)
	var watcher = func() {
		defer close(eventChan)
		for {
			select {
			case resp, ok := <-watchCh:
				if !ok {
					return
				}
				if resp.Err() != nil {
					qlog.Warnf("watch key %s canceled: %v", key, resp.Err())
					return
				}
				for _, ev := range resp.Events {
					propagateWatchEvent(eventChan, ev)
				}

			case <-ctx.Done():
				if err := c.client.Watcher.Close(); err != nil {
					qlog.Warnf("close watcher: %v", err)
				}
				return
			}
		}
	}
	go watcher()
	return eventChan
}

// 订阅目录下的所有节点变化, 并把节点变化更新到nodeMap
func (c *Client) WatchDirTo(ctx context.Context, dir string, nodeMap *NodeMap) {
	var evChan = c.WatchDir(ctx, dir)
	var prefix = c.formatKey(dir)
	var watcher = func() {
		for {
			select {
			case ev, ok := <-evChan:
				if !ok {
					return
				}
				updateNodeEvent(nodeMap, prefix, ev)
			}
		}
	}
	go watcher()
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
