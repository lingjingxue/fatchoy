// Copyright © 2020-present simon@qchen.fun All rights reserved.
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

	"qchen.fun/fatchoy/x/fsutil"
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

func NewFileSync(filename string, mode fsutil.WriterMode) zapcore.WriteSyncer {
	switch filename {
	case "stdout":
		return os.Stdout
	case "stderr":
		return os.Stderr
	}
	return fsutil.NewFileWriter(filename, 100, mode)
}
