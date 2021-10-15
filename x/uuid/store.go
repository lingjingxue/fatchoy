// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"errors"
)

var ErrIDOutOfRange = errors.New("ID out of range")

// DB超时
const OpTimeout = 300

// Storage表示一个存储组件，维持一个持续递增（不一定连续）的counter
type Storage interface {
	Incr() (int64, error)
	Close() error
}
