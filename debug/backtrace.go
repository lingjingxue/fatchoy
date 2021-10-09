// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package debug

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

const timestampLayout = "2006-01-02 15:04:05.999"

// code taken from https://github.com/pkg/error with modification

// Frame represents a program counter inside a stack frame.
// For historical reasons if Frame is interpreted as a uintptr
// its value represents the program counter + 1.
type Frame uintptr

// PC returns the program counter for this frame;
// multiple frames may have the same PC value.
func (f Frame) PC() uintptr { return uintptr(f) - 1 }

// stack represents a stack of program counters.
type Stack struct {
	pcs []uintptr
}

func (s *Stack) String() string {
	var sb strings.Builder
	for i, v := range s.pcs {
		pc := Frame(v).PC()
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			break
		}
		file, line := fn.FileLine(pc)
		fnName := fn.Name()
		fmt.Fprintf(&sb, "% 3d. %s() %s:%d\n", i+1, fnName, file, line)
		if fnName == "main.main" {
			break
		}
	}
	return sb.String()
}

// 获取当前调用堆栈
func GetCallerStack(stack *Stack, skip int) {
	if stack.pcs == nil {
		stack.pcs = make([]uintptr, 32) // 32 depth is enough
	}
	n := runtime.Callers(skip, stack.pcs[:])
	stack.pcs = stack.pcs[0:n]
}

func Backtrace(val interface{}, f *os.File) {
	if f == nil {
		f = os.Stderr
	}
	var stack Stack
	GetCallerStack(&stack, 2)

	var buf bytes.Buffer
	var now = time.Now()
	fmt.Fprintf(&buf, "Traceback[%s] (most recent call last):\n", now.Format(timestampLayout))
	fmt.Fprintf(&buf, "%v %v\n", stack, val)
	buf.WriteTo(f)
}

func CatchPanic() {
	if v := recover(); v != nil {
		Backtrace(v, os.Stderr)
	}
}
