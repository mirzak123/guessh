package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
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

	setupOnce sync.Once
)

func EnsureLoggerSetup(logFileName string) {
	setupOnce.Do(func() {
		logFile, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open log file %s: %v. Falling back to stderr.\n", logFileName, err)
			initLogger(os.Stderr, getLogLevelFromEnv())
			return
		}

		initLogger(logFile, getLogLevelFromEnv())
	})
}

func initLogger(out io.Writer, level LogLevel) {
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

func getLogLevelFromEnv() LogLevel {
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
