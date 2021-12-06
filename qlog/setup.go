// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package qlog

import (
	"io"
	"io/ioutil"
	"os"
	"sort"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/ssh/terminal"
	"qchen.fun/fatchoy/x/fsutil"
)

var (
	_logger = zap.L() // core logger
	_sugar  = zap.S() // sugared logger
)

func isTerminal(w io.Writer) bool {
	switch v := w.(type) {
	case *os.File:
		return terminal.IsTerminal(int(v.Fd()))
	default:
		return false
	}
}

// 创建一个logger
func createLoggerBy(conf *zap.Config, opts []zap.Option) *zap.Logger {
	var encoder zapcore.Encoder
	switch conf.Encoding {
	case "console":
		encoder = zapcore.NewConsoleEncoder(conf.EncoderConfig)
	case "json":
		encoder = zapcore.NewJSONEncoder(conf.EncoderConfig)
	default:
		panic("unrecognized encoding: " + conf.Encoding)
	}

	writer := openLoggerSink(conf.OutputPaths)
	writeErr := openLoggerSink(conf.ErrorOutputPaths)

	logger := zap.New(
		zapcore.NewCore(encoder, writer, conf.Level),
		buildOptions(conf, writeErr)...,
	)
	logger = logger.WithOptions(opts...)
	return logger
}

// 创建自定义的FileSync（以支持log rotation）
func openLoggerSink(paths []string) zapcore.WriteSyncer {
	if len(paths) == 0 {
		return zapcore.AddSync(ioutil.Discard)
	}
	var writers = make([]zapcore.WriteSyncer, 0, len(paths))
	for _, path := range paths {
		var w = NewFileSync(path, fsutil.WriterSync)
		writers = append(writers, w)
	}
	return zap.CombineWriteSyncers(writers...)
}

// from go.uber.org/zap/config.go
func buildOptions(cfg *zap.Config, errSink zapcore.WriteSyncer) []zap.Option {
	var opts = make([]zap.Option, 1)
	opts[0] = zap.ErrorOutput(errSink)

	if cfg.Development {
		opts = append(opts, zap.Development())
	}

	if !cfg.DisableCaller {
		opts = append(opts, zap.AddCaller())
	}

	stackLevel := zap.ErrorLevel
	if cfg.Development {
		stackLevel = zap.WarnLevel
	}
	if !cfg.DisableStacktrace {
		opts = append(opts, zap.AddStacktrace(stackLevel))
	}

	if scfg := cfg.Sampling; scfg != nil {
		opts = append(opts, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			var samplerOpts []zapcore.SamplerOption
			if scfg.Hook != nil {
				samplerOpts = append(samplerOpts, zapcore.SamplerHook(scfg.Hook))
			}
			return zapcore.NewSamplerWithOptions(
				core,
				time.Second,
				cfg.Sampling.Initial,
				cfg.Sampling.Thereafter,
				samplerOpts...,
			)
		}))
	}

	if len(cfg.InitialFields) > 0 {
		fs := make([]zap.Field, 0, len(cfg.InitialFields))
		keys := make([]string, 0, len(cfg.InitialFields))
		for k := range cfg.InitialFields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fs = append(fs, zap.Any(k, cfg.InitialFields[k]))
		}
		opts = append(opts, zap.Fields(fs...))
	}

	return opts
}
