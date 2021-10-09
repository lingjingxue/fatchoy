// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fatchoy

// 定义应用层服务接口
type Service interface {
	Type() int8
	Name() string

	NodeID() NodeID
	SetNodeID(id NodeID)

	// 初始化、启动和关闭
	Init(*ServiceContext) error
	Startup() error
}
