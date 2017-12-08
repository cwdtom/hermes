// io相关工具 author chenweidong

package http_server

import (
	"fmt"
	"time"
)

const (
	DEBUG = iota
	INFO
	WARN
	ERROR
)

func NewLog(level int) *log {
	return &log{level: level}
}

type log struct {
	level int
}

func (l *log) SetLevel(level int) {
	l.level = level
}

func (l *log) GetLevel() int {
	return l.level
}

func (l *log) Debug(s string, a ...interface{}) {
	if l.level <= DEBUG {
		fmt.Printf("["+time.Now().Format("2006-01-02 15:04:05")+"] DEBUG: ["+s+"]\n", a...)
	}
}

func (l *log) Info(s string, a ...interface{}) {
	if l.level <= INFO {
		fmt.Printf("["+time.Now().Format("2006-01-02 15:04:05")+"] INFO: "+s+"\n", a...)
	}
}

func (l *log) Warn(s string, a ...interface{}) {
	if l.level <= WARN {
		fmt.Printf("["+time.Now().Format("2006-01-02 15:04:05")+"] WARN: "+s+"\n", a...)
	}
}

func (l *log) Error(s string, a ...interface{}) {
	if l.level <= ERROR {
		fmt.Printf("["+time.Now().Format("2006-01-02 15:04:05")+"] ERROR: "+s+"\n", a...)
	}
}