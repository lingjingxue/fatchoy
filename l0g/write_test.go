// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package l0g

import (
	"os"
	"testing"

	"qchen.fun/fatchoy/x/fsutil"
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
