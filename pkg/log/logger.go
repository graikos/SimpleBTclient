package log

import (
	"fmt"
	"log"
)

const (
	DEBUG = iota
	WARN
	INFO
	NORMAL
)

type Logger interface {
	Println(...interface{})
	Printf(string, ...interface{})
	Set(int)
	Debug(...interface{})
	Warn(...interface{})
	Log(...interface{})
	Info(...interface{})
}

func NewLogger(level int) Logger {
	return &loggerImpl{level: level}
}

type loggerImpl struct {
	level int
}

func (l *loggerImpl) Println(v ...interface{}) {
	fmt.Println(v...)
}

func (l *loggerImpl) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

func (l *loggerImpl) Debug(v ...interface{}) {
	if l.level == DEBUG {
		log.Println(append([]interface{}{"DEBUG:"}, v...)...)
	}
}

func (l *loggerImpl) Warn(v ...interface{}) {
	if l.level == WARN || l.level == DEBUG {
		log.Println(append([]interface{}{"WARNING:"}, v...)...)
	}
}

func (l *loggerImpl) Info(v ...interface{}) {
	if l.level == WARN || l.level == DEBUG || l.level == INFO {
		log.Println(append([]interface{}{"INFO:"}, v...)...)
	}
}

func (l *loggerImpl) Log(v ...interface{}) {
	log.Println(v...)
}

func (l *loggerImpl) Set(level int) {
	l.level = level
}
