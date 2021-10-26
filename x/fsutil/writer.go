// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package fsutil

import (
	"fmt"
	"os"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type WriterMode int

const (
	WriterSync      WriterMode = 0
	WriterAsync     WriterMode = 1
	WriterAsyncCopy WriterMode = 2
)

const (
	DefaultMaxFileBackup = 20   //
	DefaultMaxMBPerFile  = 100  // 100M
	DefaultCapacity      = 6000 //
)

type FileWriter struct {
	done   chan struct{}
	wg     sync.WaitGroup    //
	sync   chan struct{}     // 同步信号
	mode   WriterMode        // 异步模式
	bus    chan []byte       // 待写内容
	logger lumberjack.Logger // log rotation
}

func NewFileWriter(filename string, maxMBPerFile int, mode WriterMode) *FileWriter {
	if maxMBPerFile <= 0 {
		maxMBPerFile = DefaultMaxMBPerFile
	}
	w := &FileWriter{
		mode: mode,
		logger: lumberjack.Logger{
			Filename:   filename,
			MaxSize:    maxMBPerFile,
			MaxBackups: DefaultMaxFileBackup,
			LocalTime:  true,
			Compress:   true,
		},
	}
	if (mode | WriterAsync) > 0 {
		w.sync = make(chan struct{})
		w.done = make(chan struct{})
		w.bus = make(chan []byte, DefaultCapacity)
		w.wg.Add(1)
		go w.writePump()
	}

	return w
}

func (w *FileWriter) Close() error {
	if (w.mode | WriterAsync) > 0 {
		close(w.done)
		w.wg.Wait()
		close(w.bus)
		w.bus = nil
		w.done = nil
	}
	return w.logger.Close()
}

func (w *FileWriter) Write(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	if (w.mode | WriterAsync) > 0 {
		if (w.mode | WriterAsyncCopy) > 0 {
			var newb = make([]byte, len(data))
			copy(newb, data)
			data = newb
		}
		w.bus <- data // this may block current goroutine
		return 0, nil
	}
	return w.logger.Write(data)
}

func (w *FileWriter) flush() error {
	var err error
	for {
		select {
		case data := <-w.bus:
			if _, err = w.logger.Write(data); err != nil {
				fmt.Fprintf(os.Stderr, "FileWriter: %v", err)
			}
		default:
			return err
		}
	}
}

// 立即刷新
func (w *FileWriter) Sync() error {
	select {
	case w.sync <- struct{}{}:
	default:
		return nil
	}
	return nil
}

// 刷新线程，定时执行刷盘
func (w *FileWriter) writePump() {
	defer w.wg.Done()
	var ticker = time.NewTicker(time.Millisecond * 30)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			w.flush()

		case <-w.sync:
			w.flush()

		case <-w.done:
			return
		}
	}
}
