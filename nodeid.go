// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

import (
	"fmt"
	"log"
	"strconv"

	"gopkg.in/qchencc/fatchoy/x/collections"
)

const (
	NodeServiceShift = 16
	NodeServiceMask  = 0xFF00FFFF
	NodeInstanceMask = 0xFFFF0000
	NodeTypeShift    = 31
	NodeTypeClient   = NodeID(1 << NodeTypeShift)
)

// 一个32位整数表示的节点号，用以标识一个service（最高位为0），或者一个客户端session(最高位为1)
// 如果是服务编号：8位服务编号，16位服务实例编号
//
//	服务实例二进制布局
// 		--------------------------------------------
// 		| type | reserved |  service  |  instance  |
// 		--------------------------------------------
// 		32    31          24         16            0
//

// 节点号
type NodeID uint32

func MakeNodeID(service uint8, instance uint16) NodeID {
	return NodeID((uint32(service) << NodeServiceShift) | uint32(instance))
}

func MustParseNodeID(s string) NodeID {
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		log.Panicf("MustParseNode: %v", err)
	}
	return NodeID(n)
}

// 是否backend instance
func (n NodeID) IsBackend() bool {
	return (uint32(n) & uint32(NodeTypeClient)) == 0
}

// 服务类型编号
func (n NodeID) Service() uint8 {
	return uint8(n >> NodeServiceShift)
}

// 实例编号
func (n NodeID) Instance() uint16 {
	return uint16(n)
}

func (n NodeID) String() string {
	return fmt.Sprintf("%02x%04x", int8(n.Service()), n.Instance())
}

// 没有重复ID的有序集合
type NodeIDSet = collections.OrderedIDSet