package main

import (
	"fmt"
)

// 系统日志等级
type type_log_level int

const (
	LOG_LEVEL_DEBUG type_log_level = iota // 0
	LOG_LEVEL_INFO                        // 1
	LOG_LEVEL_ERROR                       // 2
)

// 系统日志类型
var g_log_level = LOG_LEVEL_DEBUG

type ModuleLogger struct {
	path string
}

var g_logger *ModuleLogger

func ChangeSysLogLevel(lv type_log_level) {
	g_log_level = lv
}

func New(path string) *ModuleLogger {
	return &ModuleLogger{
		path: path,
	}
}

func CreateLocalLog() {
	path := "log/gowebserv.log"
	g_logger = New(path)
}

func (lg *ModuleLogger) WriteFunc(lv type_log_level, cls, format string, args ...interface{}) {
	if lv < g_log_level {
		return
	}
	ctn := fmt.Sprintf(format, args...)
	ctn = fmt.Sprintf("[%s] %s", cls, ctn)
	WriteFile(lg.path, ctn)
}

func (lg *ModuleLogger) Debug(format string, args ...interface{}) {
	lg.WriteFunc(LOG_LEVEL_DEBUG, "Debug", format, args...)
}

func (lg *ModuleLogger) Info(format string, args ...interface{}) {
	lg.WriteFunc(LOG_LEVEL_INFO, "Info", format, args...)
}

func (lg *ModuleLogger) Error(format string, args ...interface{}) {
	lg.WriteFunc(LOG_LEVEL_ERROR, "Error", format, args...)
}
