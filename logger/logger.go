package logger

import (
	"io"
)

type Logger interface {
	Debugf(format string, args ...interface{})
	Debug(args ...interface{})

	Infof(format string, args ...interface{})
	Info(args ...interface{})

	Warnf(format string, args ...interface{})
	Warn(args ...interface{})

	Errorf(format string, args ...interface{})
	Error(args ...interface{})

	DebugWriter() io.Writer
	InfoWriter() io.Writer
	WarnWriter() io.Writer
	ErrorWriter() io.Writer

	Close() error
}

var logger Logger = NewStandardLogger(DefaultConfig, 2)

func Override(l Logger) {
	if logger != nil {
		logger.Close()
	}
	logger = l
}

func Close() error {
	return logger.Close()
}
