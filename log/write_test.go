// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package log

import (
	"os"
	"testing"

	"gopkg.in/qchencc/fatchoy.v1/x/fsutil"
)

func TestWriteFileLog(t *testing.T) {
	filename := "file.log"
	WriteFileLog(filename, "hello world")
	os.Remove(filename)
}

func TestServerErrorLog(t *testing.T) {
	ServerErrorLog("example.com")
}

func TestNewFileSync(t *testing.T) {
	var w = NewFileSync("app-log", fsutil.WriterSync)
	w.Write([]byte("hello"))
	w.Sync()
}
