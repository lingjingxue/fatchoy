// Copyright © 2021 ichenq@outlook.com. All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

package version

import (
	"fmt"
	"runtime"
	"strings"
)

// 一个版本号（如1.0.1）由`major.minor.patch`三部分组成
//
// `major`: 主版本号
// `minor`: 次版本号
// `patch`: 修订号
//
// 第一个初始开发版本使用`0.1.0`
// 第一个可以对外发布的版本使用`1.0.0`
//


var (
	_Version   = "0.1.0"
	_Branch    = "v1"
	_CommitRev = "???"
	_BuildTime = "???"
)

// 版本号
func Version() string {
	return _Version
}

// 代码分支
func Branch() string {
	return _Branch
}

// 提交版本
func CommitRev() string {
	return _CommitRev
}

// 打包时间
func BuildTime() string {
	return _BuildTime
}

func String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Version: %s\n", _Version)
	fmt.Fprintf(&sb, "Revision: %s-%s\n", _Branch, _CommitRev)
	fmt.Fprintf(&sb, "Built at: %s\n", _BuildTime)
	fmt.Fprintf(&sb, "Powered by: %s", runtime.Version())
	return sb.String()
}

func Print() {
	println(String())
}
