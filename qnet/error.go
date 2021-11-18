// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qnet

import (
	"errors"
	"fmt"

	"qchen.fun/fatchoy"
)

var (
	ErrConnIsClosing        = errors.New("connection is closing when sending")
	ErrConnOutboundOverflow = errors.New("connection outbound queue overflow")
	ErrConnForceClose       = errors.New("connection forced to close")
)

type Error struct {
	Err      error
	Endpoint fatchoy.Endpoint
}

func NewError(err error, endpoint fatchoy.Endpoint) *Error {
	return &Error{
		Err:      err,
		Endpoint: endpoint,
	}
}

func (e Error) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("node %v(%s) EOF", e.Endpoint.NodeID(), e.Endpoint.RemoteAddr())
	}
	return fmt.Sprintf("node %v(%s) %s", e.Endpoint.NodeID(), e.Endpoint.RemoteAddr(), e.Err.Error())
}
