// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qlog

import (
	"log"

	"go.uber.org/zap"
)

// 日志模块需求:
// 	控制台打印
// 	文件打印(filebeat)

func Setup(cfg *Config) {
	if cfg == nil {
		cfg = NewConfig("debug")
	}

	var l = NewLogger(cfg)
	_logger = l
	_sugar = l.Sugar()
}

func NewLogger(cfg *Config) *zap.Logger {
	if cfg == nil {
		cfg = NewConfig("debug")
	}
	return cfg.Build()
}

func Reset() {
	if err := _logger.Sync(); err != nil {
		log.Printf("sync logger: %v", err)
	}
	_logger = zap.L()
	_sugar = zap.S()
}

func Debugf(format string, args ...interface{}) {
	_sugar.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	_sugar.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	_sugar.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	_sugar.Errorf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	_sugar.Panicf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	_sugar.Fatalf(format, args...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	_sugar.Debugw(msg, keysAndValues)
}

func Infow(msg string, keysAndValues ...interface{}) {
	_sugar.Infow(msg, keysAndValues)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	_sugar.Warnw(msg, keysAndValues)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	_sugar.Errorw(msg, keysAndValues)
}

func Fatalw(msg string, keysAndValues ...interface{}) {
	_sugar.Fatalw(msg, keysAndValues)
}
