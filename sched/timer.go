// Copyright © 2020 qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

type Timer interface {
	RunAt(ts int64, r Runner) int
	RunAfter(interval int, r Runner) int
	RunEvery(interval int, r Runner) int

	Cancel(id int) bool
}
