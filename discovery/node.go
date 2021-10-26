// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package discovery

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	NODE_KEY_ID        = "ID"
	NODE_KEY_TYPE      = "Type"
	NODE_KEY_INTERFACE = "Interface"
	NODE_KEY_PID       = "PID"
	NODE_KEY_HOST      = "Host"
)

type NodeEventType int

const (
	EventUnknown NodeEventType = 0
	EventCreate  NodeEventType = 1
	EventUpdate  NodeEventType = 2
	EventDelete  NodeEventType = 3
)

// 节点事件
type NodeEvent struct {
	Type NodeEventType
	Key  string
	Node Node
}

type INode interface {
	ID() int16
	Type() string
	Interface() string
}

// 一个节点信息
type Node map[string]string

func NewNode(nodeType string, id uint16) Node {
	node := map[string]string{
		NODE_KEY_TYPE: nodeType,
		NODE_KEY_ID:   strconv.Itoa(int(id)),
		NODE_KEY_PID:  strconv.Itoa(os.Getpid()),
	}
	if hostname, err := os.Hostname(); err == nil {
		node[NODE_KEY_HOST] = hostname
	}
	return node
}

// 节点类型
func (n Node) Type() string {
	return n.Get(NODE_KEY_TYPE)
}

// 节点ID
func (n Node) ID() uint16 {
	return uint16(n.GetInt(NODE_KEY_ID))
}

// 节点接口地址
func (n Node) Interface() string {
	return n.Get(NODE_KEY_INTERFACE)
}

func (n Node) GetInt(key string) int {
	s := n.Get(key)
	i, _ := strconv.Atoi(s)
	return i
}

func (n Node) Get(key string) string {
	return n[key]
}

func (n Node) Set(key, v string) {
	n[key] = v
}

func (n Node) String() string {
	var sb strings.Builder
	sb.WriteByte('[')
	for k, v := range n {
		fmt.Fprintf(&sb, "%s: %v ", k, v)
	}
	sb.WriteByte(']')
	return sb.String()
}

// 节点列表
type NodeSet []Node

// 按服务类型区分的节点信息
type NodeMap struct {
	guard sync.RWMutex
	nodes map[string]NodeSet
}

func NewNodeMap() *NodeMap {
	return &NodeMap{
		nodes: make(map[string]NodeSet),
	}
}

// 所有节点数量
func (m *NodeMap) Count() int {
	m.guard.RLock()
	var count = 0
	for _, nodes := range m.nodes {
		count += len(nodes)
	}
	m.guard.RUnlock()
	return count
}

func (m *NodeMap) GetKeys() []string {
	m.guard.RLock()
	var names = make([]string, 0, len(m.nodes))
	for name := range m.nodes {
		names = append(names, name)
	}
	m.guard.RUnlock()
	return names
}

// 所有本类型的节点，不要修改返回值
func (m *NodeMap) GetNodes(nodeType string) NodeSet {
	m.guard.RLock()
	v := m.nodes[nodeType]
	m.guard.RUnlock()
	return v
}

// 添加一个节点
func (m *NodeMap) InsertNode(node Node) {
	m.guard.Lock()
	defer m.guard.Unlock()

	var stype = node.Type()
	slice := m.nodes[stype]
	for i, v := range slice {
		if v.ID() == node.ID() {
			slice[i] = node
			return
		}
	}
	m.nodes[stype] = append(slice, node)
}

func (m *NodeMap) Clear() {
	m.guard.Lock()
	m.nodes = make(map[string]NodeSet)
	m.guard.Unlock()
}

// 删除某一类型的所有节点
func (m *NodeMap) DeleteNodes(nodeType string) {
	m.guard.Lock()
	m.nodes[nodeType] = nil
	m.guard.Unlock()
}

// 删除一个节点
func (m *NodeMap) DeleteNode(nodeType string, id uint16) {
	m.guard.Lock()
	defer m.guard.Unlock()

	slice := m.nodes[nodeType]
	var idx = -1
	for i, v := range slice {
		if v.ID() == id {
			idx = i
			break
		}
	}
	if idx >= 0 {
		var last = len(slice) - 1
		slice[last], slice[idx] = slice[idx], slice[last]
		slice[last] = nil
		m.nodes[nodeType] = slice[:last]
		if len(m.nodes[nodeType]) == 0 {
			delete(m.nodes, nodeType)
		}
	}
}

func (m *NodeMap) String() string {
	var sb strings.Builder
	for name, set := range m.nodes {
		fmt.Fprintf(&sb, "%s: %v,\n", name, set)
	}
	return sb.String()
}
