package logger

import (
	"io"
	"log"
)

type LogLevel string

const (
	LogLevelInfo  LogLevel = "INFO"
	LogLevelDebug LogLevel = "DEBUG"
)

var (
	infoLogger  *log.Logger
	debugLogger *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
)

func Init(out io.Writer, level LogLevel) {
	flags := log.Lshortfile

	infoLogger = log.New(out, "[INFO] ", flags)
	warnLogger = log.New(out, "[WARN] ", flags)
	errorLogger = log.New(out, "[ERROR] ", flags)

	if level == LogLevelDebug {
		debugLogger = log.New(out, "[DEBUG] ", flags)
	} else {
		debugLogger = log.New(io.Discard, "", 0)
	}

	Info("Log level set to: %s", level)
}

func Info(msg string, args ...any) {
	infoLogger.Printf(msg, args...)
}

func Debug(msg string, args ...any) {
	debugLogger.Printf(msg, args...)
}

func Warn(msg string, args ...any) {
	warnLogger.Printf(msg, args...)
}

func Error(msg string, args ...any) {
	errorLogger.Printf(msg, args...)
}
