// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fsutil

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	DefaultMaxFileBackup = 20   //
	DefaultMaxMBPerFile  = 100  // 100M
	DefaultCapacity      = 5000 //
)

type FileWriter struct {
	done       chan struct{}
	wg         sync.WaitGroup //
	asyncWrite bool
	bus        chan []byte       // 待写内容
	logger     lumberjack.Logger // log rotation
}

func NewFileWriter(filename string, maxMBPerFile int, asyncWrite bool) *FileWriter {
	if maxMBPerFile <= 0 {
		maxMBPerFile = DefaultMaxMBPerFile
	}
	w := &FileWriter{
		asyncWrite: asyncWrite,
		done:       make(chan struct{}),
		bus:        make(chan []byte, DefaultCapacity),
		logger: lumberjack.Logger{
			Filename:   filename,
			MaxSize:    maxMBPerFile,
			MaxBackups: DefaultMaxFileBackup,
			Compress:   true,
		},
	}
	w.wg.Add(1)
	go w.writePump()

	return w
}

func (w *FileWriter) Close() error {
	close(w.done)
	w.wg.Wait()
	close(w.bus)
	w.bus = nil
	w.done = nil
	return w.logger.Close()
}

func (w *FileWriter) Write(b []byte) (int, error) {
	if w.asyncWrite {
		w.bus <- b
		return 0, nil
	}
	return w.Flush(b)
}

func (w *FileWriter) Flush(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	n, err := w.logger.Write(data)
	return n, err
}

// 刷新线程，定时执行刷盘
func (w *FileWriter) writePump() {
	defer w.wg.Done()
	for {
		select {
		case data := <-w.bus:
			if _, err := w.Flush(data); err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
			}

		case <-w.done:
			return
		}
	}
}
