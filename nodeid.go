// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"fmt"
	"log"
	"strconv"

	"qchen.fun/fatchoy/collections"
)

const (
	NodeServiceShift = 16
	NodeTypeShift    = 31
	NodeServiceMask  = 0x00FF0000
	NodeInstanceMask = 0x0000FFFF
)

// 节点ID
// 一个32位整数表示的节点号，用以标识一个service（最高位为0），或者一个客户端session(最高位为1)
// 如果是服务编号：8位服务编号，16位服务实例编号
//
//	服务实例二进制布局
// 		--------------------------------------
// 		|  reserved |  service  |  instance  |
// 		--------------------------------------
// 		32          24         16            0
type NodeID uint32

// 根据服务号和实例号创建一个节点ID
func MakeNodeID(service uint8, instance uint16) NodeID {
	return NodeID((uint32(service) << NodeServiceShift) | uint32(instance))
}

// 解析16进制字符串的节点ID
func MustParseNodeID(s string) NodeID {
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		log.Panicf("MustParseNodeID: %v", err)
	}
	return NodeID(n)
}

// 是否service节点
func (n NodeID) IsTypeBackend() bool {
	return (uint32(n) & uint32(1<<NodeTypeShift)) == 0
}

// service节点的service类型
func (n NodeID) Service() uint8 {
	return uint8(n >> NodeServiceShift)
}

// service节点的实例编号
func (n NodeID) Instance() uint16 {
	return uint16(n)
}

func (n NodeID) String() string {
	return fmt.Sprintf("%02x%04x", int8(n.Service()), n.Instance())
}

// 没有重复ID的有序集合
type NodeIDSet = collections.OrderedIDSet
