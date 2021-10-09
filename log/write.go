// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package log

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/ssh/terminal"

	"gopkg.in/qchencc/fatchoy/x/fsutil"
)

const timestampLayout = "2006-01-02 15:04:05.999"

func WriteFileLog(filename, format string, a ...interface{}) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, format, a...)
	return err
}

func AppFileErrorLog(format string, a ...interface{}) error {
	appname := filepath.Base(os.Args[0])
	if i := strings.LastIndex(appname, "."); i > 0 {
		appname = appname[:i]
	}
	if i := strings.LastIndex(appname, "_"); i > 0 {
		appname = appname[i+1:]
	}
	var filename = fmt.Sprintf("logs/%s_error.log", appname)
	return WriteFileLog(filename, format, a...)
}

// 用于在初始化服务时的错误日志
func ServerErrorLog(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	fmt.Fprintf(os.Stderr, "\n")
	var isTerminal = terminal.IsTerminal(int(os.Stderr.Fd()))
	if !isTerminal {
		AppFileErrorLog("%s ", time.Now().Format(timestampLayout))
		AppFileErrorLog(format, a...)
		AppFileErrorLog("\n")
	}
}

// 记录日志到文件
type FileSync struct {
	asyncWrite bool
	w          *fsutil.FileWriter
}

func NewFileSync(filename string, asyncWrite bool) zapcore.WriteSyncer {
	switch filename {
	case "stdout":
		return os.Stdout
	case "stderr":
		return os.Stderr
	}
	w := fsutil.NewFileWriter(filename, 100, asyncWrite)
	return &FileSync{
		asyncWrite: asyncWrite,
		w:          w,
	}
}

func (s *FileSync) Write(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	//  data在zap里使用buffer pool管理，如果异步write需要自己做copy
	if s.asyncWrite {
		var b = make([]byte, len(data))
		copy(b, data)
		data = b
	}
	return s.w.Write(data)
}

// zap.WriteSyncer interface
func (s *FileSync) Sync() error {
	return nil
}

func (s *FileSync) Close() error {
	return s.w.Close()
}
