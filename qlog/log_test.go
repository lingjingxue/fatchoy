// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qlog

import (
	"testing"
)

func TestSetup(t *testing.T) {
	Setup(NewConfig("debug"))
}