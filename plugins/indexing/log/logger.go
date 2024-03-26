package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/hashicorp/go-hclog"
)

var (
	logFileEnvName  = "PLUGIN_LOG_FILE"
	logLevelEnvName = "PLUGIN_LOG_LEVEL"
)

type Logger struct {
	baseLogger hclog.Logger
}

func (l *Logger) Trace(msg string, args ...interface{}) {
	l.baseLogger.Trace(msg, args...)
}

func (l *Logger) Debug(msg string, args ...interface{}) {
	l.baseLogger.Debug(msg, args...)
}

func (l *Logger) Info(msg string, args ...interface{}) {
	l.baseLogger.Info(msg, args...)
}

func (l *Logger) Warn(msg string, args ...interface{}) {
	l.baseLogger.Warn(msg, args...)
}

func (l *Logger) Error(msg string, args ...interface{}) {
	l.baseLogger.Error(msg, args...)
}

// Fatal is equivalent to Error(msg, "error", err) followed by a call to panic(err)
func (l *Logger) Fatal(msg string, err error) {
	l.baseLogger.Error(msg, "error", err)
	// Also printing to stdout so the plugin server error message is more informative
	fmt.Println(err)
	panic(err)
}

func GetLogFile() (*os.File, error) {
	logFileName, found := os.LookupEnv(logFileEnvName)
	if !found {
		err := fmt.Errorf("missing environment variable '%s'", logFileEnvName)
		return nil, err
	}

	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o640)
	if err != nil {
		return nil, err
	}

	return logFile, nil
}

func NewLogger(output io.Writer) *Logger {
	logLevel, found := os.LookupEnv(logLevelEnvName)
	if !found {
		logLevel = "default"
	}

	return &Logger{
		baseLogger: hclog.New(&hclog.LoggerOptions{
			Output: output,
			Level:  toHclogLevel(logLevel),
		}),
	}
}

func toHclogLevel(s string) hclog.Level {
	switch s {
	case "trace":
		return hclog.Trace
	case "debug":
		return hclog.Debug
	case "info":
		return hclog.Info
	case "warn":
		return hclog.Warn
	case "error":
		return hclog.Error
	default:
		return hclog.DefaultLevel
	}
}
