// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"net"
	"sync"

	"qchen.fun/fatchoy/x/cipher"
	"qchen.fun/fatchoy/x/stats"
)

type EndpointFlag uint32

const (
	EndpointReader     EndpointFlag = 0x01 // 只开启reader
	EndpointWriter     EndpointFlag = 0x02 // 只开启writer
	EndpointReadWriter EndpointFlag = 0x03 // 开启reader和writer
)

type MessageEndpoint interface {
	NodeID() NodeID
	SetNodeID(NodeID)
	RemoteAddr() string

	// 发送消息
	SendPacket(IPacket) error

	// 关闭读/写
	Close() error
	ForceClose(error)
	IsClosing() bool

	SetUserData(interface{})
	UserData() interface{}
}

// 网络连接端点
type Endpoint interface {
	MessageEndpoint

	RawConn() net.Conn
	Stats() *stats.Stats

	Go(EndpointFlag)

	// 加密解密
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

func (e *EndpointMap) Get(node NodeID) Endpoint {
	e.RLock()
	v := e.endpoints[node]
	e.RUnlock()
	return v
}

func (e *EndpointMap) Add(node NodeID, endpoint Endpoint) {
	e.Lock()
	e.endpoints[node] = endpoint
	e.Unlock()
}

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

func (e *EndpointMap) Reset() {
	e.Lock()
	e.endpoints = make(map[NodeID]Endpoint)
	e.Unlock()
}

func (e *EndpointMap) Range(f func(Endpoint) bool) {
	e.Lock()
	defer e.Unlock()
	for _, endpoint := range e.endpoints {
		if !f(endpoint) {
			break
		}
	}
}

func (e *EndpointMap) List() []Endpoint {
	e.RLock()
	var endpoints = make([]Endpoint, 0, len(e.endpoints))
	for _, v := range e.endpoints {
		endpoints = append(endpoints, v)
	}
	e.RUnlock()
	return endpoints
}
