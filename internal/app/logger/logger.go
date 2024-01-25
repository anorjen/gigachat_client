package logger

import (
	"fmt"
	"os"
	"time"
)

type Logger struct{}

var fatalFormat = "[%s] FATAL: %s %s\n"
var errorFormat = "[%s] ERROR: %s %s\n"
var infoFormat = "[%s] INFO: %s\n"
var debugFormat = "[%s] DEBUG: %s\n"

const (
	NONE = iota
	FATAL
	ERROR
	INFO
	DEBUG
	ALL
)

var level int8
var file = os.Stderr

func (l *Logger) SetLogLevel(levelStr string) {
	switch levelStr {
	case "all":
		level = ALL
	case "debug":
		level = DEBUG
	case "info":
		level = INFO
	case "error":
		level = ERROR
	default:
		level = NONE
	}
}

func (f *Logger) SetFile(filePath string) {
	var err error

	if len(filePath) == 0 {
		return
	}

	file, err = os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		f.LogError("Error to open log file: ", err)
		file = os.Stderr
		return
	}
}

func (f *Logger) Close() {
	if file != nil && file != os.Stderr {
		_ = file.Close()
	}
}

func (l *Logger) LogDebug(msg string) {
	if level >= DEBUG {
		_, _ = fmt.Fprintf(file, debugFormat, time.Now().Format("2006-01-02 15:04:05"), msg)
	}
}

func (l *Logger) LogInfo(msg string) {
	if level >= INFO {
		_, _ = fmt.Fprintf(file, infoFormat, time.Now().Format("2006-01-02 15:04:05"), msg)
	}
}

func (l *Logger) LogError(msg string, err error) {
	if level >= ERROR {
		errMsg := ""

		if err != nil {
			errMsg = err.Error()
		}
		_, _ = fmt.Fprintf(file, errorFormat, time.Now().Format("2006-01-02 15:04:05"), msg, errMsg)
	}
}

func (l *Logger) LogFatal(msg string, err error) {
	errMsg := ""

	if err != nil {
		errMsg = err.Error()
	}
	_, _ = fmt.Fprintf(file, fatalFormat, time.Now().Format("2006-01-02 15:04:05"), msg, errMsg)
	os.Exit(1)
}
