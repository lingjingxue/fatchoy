// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Hooker interface {
	Name() string
	Fire(entry zapcore.Entry) error
}

// 配置参数
type Config struct {
	Level           string // 等级
	Filepath        string // 日志文件
	ErrFilepath     string // err文件
	FileAsyncWrite  bool   // 异步写文件
	IsProduction    bool   // 是否product模式

	Conf *zap.Config
}

func NewConfig(level string) *Config {
	return &Config{
		Level: level,
	}
}

func (c *Config) makeOptions() []zap.Option {
	var opts = []zap.Option{
		zap.AddCallerSkip(1),
	}
	if c.Filepath != "" {
		c.Conf.OutputPaths = append(c.Conf.OutputPaths, c.Filepath)
	}
	c.Conf.Development = !c.IsProduction
	return opts
}

func (c *Config) makeDefault() {
	if c.Conf != nil {
		return
	}
	if c.IsProduction {
		var conf = zap.NewProductionConfig()
		c.Conf = &conf
	} else {
		var conf = zap.NewDevelopmentConfig()
		c.Conf = &conf
	}
	c.Conf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	if isTerminal(os.Stdout) {
		c.Conf.Encoding = "console"
		c.Conf.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
}

func (c *Config) Build() *zap.Logger {
	c.makeDefault()
	if c.Level != "" {
		if err := c.Conf.Level.UnmarshalText([]byte(c.Level)); err != nil {
			panic(err)
		}
	} else {
		c.Conf.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	var opts = c.makeOptions()
	return createLoggerBy(c.Conf, opts, c.FileAsyncWrite)
}
