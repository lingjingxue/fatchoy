// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

import (
	"context"
)

type RunFunc func(context.Context) error

// Runner是一个可执行对象
type Runner interface {
	Run(context.Context) error
}

type Runnable struct {
	F RunFunc
}

func (r *Runnable) Run(ctx context.Context) error {
	return r.F(ctx)
}

func NewRunner(f RunFunc) Runner {
	return &Runnable{
		F: f,
	}
}
