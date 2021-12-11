// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"net"
	"sync"

	"qchen.fun/fatchoy/x/cipher"
	"qchen.fun/fatchoy/x/stats"
)

// 开启reader/writer标记
type EndpointFlag uint32

const (
	EndpointReader     EndpointFlag = 0x01 // 只开启reader
	EndpointWriter     EndpointFlag = 0x02 // 只开启writer
	EndpointReadWriter EndpointFlag = 0x03 // 开启reader和writer
)

// 绑定到消息上的endpoint
type MessageEndpoint interface {
	// 节点ID
	NodeID() NodeID
	SetNodeID(NodeID)

	// 远端地址
	RemoteAddr() string

	// 发送消息
	SendPacket(IPacket) error

	// 关闭读/写
	Close() error
	ForceClose(error)
	IsRunning() bool

	// 绑定自定义数据
	SetUserData(interface{})
	UserData() interface{}
}

// 网络端点
type Endpoint interface {
	MessageEndpoint

	// 原始连接对象
	RawConn() net.Conn

	// 发送/接收计数数据
	Stats() *stats.Stats

	// 开启read/write线程
	Go(EndpointFlag)

	// 设置加解密
	SetEncryptPair(cipher.BlockCryptor, cipher.BlockCryptor)
}

// 线程安全的endpoint map
type EndpointMap struct {
	sync.RWMutex
	endpoints map[NodeID]Endpoint
}

func NewEndpointMap() *EndpointMap {
	return &EndpointMap{
		endpoints: make(map[NodeID]Endpoint),
	}
}

// 查找一个endpoint
func (e *EndpointMap) Get(node NodeID) Endpoint {
	e.RLock()
	v := e.endpoints[node]
	e.RUnlock()
	return v
}

// 添加一个endpoint
func (e *EndpointMap) Add(node NodeID, endpoint Endpoint) {
	e.Lock()
	e.endpoints[node] = endpoint
	e.Unlock()
}

// 删除一个endpoint
func (e *EndpointMap) Delete(node NodeID) bool {
	e.Lock()
	delete(e.endpoints, node)
	e.Unlock()
	return false
}

func (e *EndpointMap) Size() int {
	e.Lock()
	n := len(e.endpoints)
	e.Unlock()
	return n
}

// 重置
func (e *EndpointMap) Reset() {
	e.Lock()
	e.endpoints = make(map[NodeID]Endpoint)
	e.Unlock()
}

// 遍历
func (e *EndpointMap) Range(f func(Endpoint) bool) {
	e.Lock()
	defer e.Unlock()
	for _, endpoint := range e.endpoints {
		if !f(endpoint) {
			break
		}
	}
}

// 返回切片拷贝
func (e *EndpointMap) List() []Endpoint {
	e.RLock()
	var endpoints = make([]Endpoint, 0, len(e.endpoints))
	for _, v := range e.endpoints {
		endpoints = append(endpoints, v)
	}
	e.RUnlock()
	return endpoints
}
