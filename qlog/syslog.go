// Copyright © 2020 ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

// +build linux

package qlog

import (
	"fmt"
	"log/syslog"

	"gopkg.in/qchencc/fatchoy.v1/x/strutil"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SyslogHook to send logs via syslog.
type SyslogHook struct {
	writer  *syslog.Writer
	network string
	raddr   string
}

// Creates a hook to be added to an instance of logger. This is called with
// `hook, err := NewSyslogHook("udp", "localhost:514", syslog.LOG_DEBUG, "")`
func NewSyslogHook(network, raddr string, priority syslog.Priority, tag string) Hooker {
	w, err := syslog.Dial(network, raddr, priority, tag)
	if err != nil {
		panic(fmt.Errorf("dial syslog: %v", err))
	}
	return &SyslogHook{
		writer:  w,
		network: network,
		raddr:   raddr,
	}
}

func (h *SyslogHook) Name() string {
	return "syslog"
}

func (h *SyslogHook) Fire(entry zapcore.Entry) error {
	var w = h.writer
	switch entry.Level {
	case zapcore.FatalLevel:
		return w.Emerg(entry.Message)
	case zapcore.PanicLevel:
		return w.Crit(entry.Message)
	case zapcore.ErrorLevel, zapcore.DPanicLevel:
		return w.Err(entry.Message)
	case zapcore.WarnLevel:
		return w.Warning(entry.Message)
	case zapcore.InfoLevel:
		return w.Info(entry.Message)
	case zapcore.DebugLevel:
		return w.Debug(entry.Message)
	default:
		return nil
	}
}

func toSysPriority(level zapcore.Level) syslog.Priority {
	switch level {
	case zapcore.FatalLevel:
		return syslog.LOG_EMERG
	case zapcore.PanicLevel:
		return syslog.LOG_CRIT
	case zapcore.ErrorLevel, zapcore.DPanicLevel:
		return syslog.LOG_ERR
	case zapcore.WarnLevel:
		return syslog.LOG_WARNING
	case zapcore.InfoLevel:
		return syslog.LOG_INFO
	case zapcore.DebugLevel:
		return syslog.LOG_DEBUG
	default:
		return syslog.LOG_INFO
	}
}

func HookSystemLog(args map[string]string) zap.Option {
	var network = "udp"
	var addr = "localhost:514"
	var priority = toSysPriority(zapcore.DebugLevel)
	var tag string
	if v, found := args["network"]; found {
		network = v
	}
	if v, found := args["addr"]; found {
		addr = v
	}
	if v, found := args["priority"]; found {
		priority = syslog.Priority(strutil.MustParseI32(v))
	}
	if v, found := args["tag"]; found {
		tag = v
	}
	hook := NewSyslogHook(network, addr, priority, tag)
	return zap.Hooks(hook.Fire)
}
