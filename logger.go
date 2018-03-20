package main

import (
	"fmt"
	"log"
	"os"
	lpath "path"
	"time"
)

type Logger struct {
	path       string
	fh         *os.File
	logger     *log.Logger
	createTime time.Time
}

var all_rpc_logs = make(map[string]*Logger)

const day_seconds = 24 * 3600

func GetZeroClock(t time.Time) time.Time {
	zeroClock := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return zeroClock
}

func CreateLog(path string) *Logger {
	lg, ok := all_rpc_logs[path]
	if !ok {
		dir := lpath.Dir(path)
		os.MkdirAll(dir, 0666)
		fh, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("[CreateLog] failed: %s: %s", path, err.Error())
			return nil
		}
		l := log.New(fh, "", log.LstdFlags)
		now := time.Now()
		zeroClock := GetZeroClock(now)
		lg = &Logger{
			path:       path,
			fh:         fh,
			logger:     l,
			createTime: zeroClock,
		}
		all_rpc_logs[path] = lg
	}
	return lg
}

func GetLog(path string) *Logger {
	lg, ok := all_rpc_logs[path]
	if !ok {
		return nil
	}
	return lg
}

func WriteFile(path, str string) {
	lg := GetLog(path)
	if lg == nil {
		lg = CreateLog(path)
	} else {
		now := time.Now()
		sec := lg.createTime.Second() + day_seconds
		if now.Second() >= sec {
			RollFile(now)
			lg = CreateLog(path)
		}
	}
	if lg == nil {
		log.Printf("no log file: %s\n", path)
		return
	}

	lg.logger.Printf("%s\n", str)
	lg.fh.Sync()
}

// 每天更换一次日志文件名
func RollFile(now time.Time) {
	var newName string
	for k, v := range all_rpc_logs {
		v.fh.Sync()
		v.fh.Close()
		delete(all_rpc_logs, k)
		y, m, d := v.createTime.Date()
		newName = fmt.Sprintf("%s.%d%02d%02d", k, y, int(m), d)
		os.Rename(k, newName)
	}
}

func Test() {
	path := "log/key/test.log"
	logger := New(path)
	logger.Debug("%s==%d", "哈哈", 45623312)
	logger.Info("%s===%s", "hee", "545rt")
	logger.Error("%s====%s", "啊大大", "xxxxxxxxxxx")
	logger.Debug("%s=====%s", "sfsasf dfgd", "cvdfasfasf")
	//RollFile(time.Now())
	//logger.Debug((path, "after rename")
}
