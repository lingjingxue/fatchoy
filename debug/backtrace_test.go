// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package debug

import (
	"bytes"
	"testing"
)

func TestBacktrace(t *testing.T) {
	defer CatchPanic()
	var buf bytes.Buffer
	Backtrace("", &buf)
	t.Logf("%s", buf.String())
}
