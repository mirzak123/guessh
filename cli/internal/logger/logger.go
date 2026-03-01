package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
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
	flags := log.Lshortfile | log.Ldate | log.Ltime

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
	logf(infoLogger, 3, msg, args...)
}

func Debug(msg string, args ...any) {
	logf(debugLogger, 3, msg, args...)
}

func Warn(msg string, args ...any) {
	logf(warnLogger, 3, msg, args...)
}

func Error(msg string, args ...any) {
	logf(errorLogger, 3, msg, args...)
}

func logf(l *log.Logger, depth int, format string, v ...any) {
	_ = l.Output(depth, fmt.Sprintf(format, v...))
}

func GetLogLevelFromEnv() LogLevel {
	logLevelStr, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		logLevelStr = "INFO"
	}

	logLevelStr = strings.ToUpper(logLevelStr)

	switch logLevelStr {
	case "DEBUG":
		return LogLevelDebug
	default:
		return LogLevelInfo
	}
}
