package logger

import (
	"fmt"
	"github.com/fahedouch/go-logrotate"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
)

var logger Logger = NewStandardLogger(DefaultConfig, 2)

func Override(l Logger) {
	logger = l
}

var DefaultConfig = &Config{
	Level:      logrus.InfoLevel.String(),
	Format:     "2006-01-02-15-04-05.000",
	Filepath:   "log/app.log",
	MaxBackups: 10,
	MaxAge:     15,
	MaxBytes:   1024 * 1024 * 10,
	Compress:   false,
}

// Formatter provides log format for logger
type Formatter struct {
}

func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	// current log level
	logLevel := entry.Level
	timeFormatted := entry.Time.Format("2006-01-02 15:04:05")

	msg := entry.Message

	return []byte(fmt.Sprintf("[%-7s][%s]  %s\n",
		strings.ToUpper(logLevel.String()), timeFormatted, msg)), nil

}

type Config struct {
	Filepath   string
	Level      string
	Format     string
	MaxBackups int
	MaxAge     int
	MaxBytes   int64
	Compress   bool
}

func NewStandardLogger(config *Config, callerSkip int) Logger {
	l := logrus.New()
	l.SetFormatter(&Formatter{})
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	l.SetLevel(level)

	// compatible system log with debug level
	log.SetOutput(l.WriterLevel(logrus.DebugLevel))

	output := io.MultiWriter(os.Stdout, &logrotate.Logger{
		Filename:           config.Filepath,
		FilenameTimeFormat: config.Format,
		MaxBytes:           config.MaxBytes,
		MaxBackups:         config.MaxBackups,
		MaxAge:             config.MaxAge, //days
		Compress:           config.Compress,
	})
	l.SetOutput(output)
	return &StandardLogger{
		logger:     l,
		callerSkip: callerSkip,
	}
}

type StandardLogger struct {
	logger     *logrus.Logger
	callerSkip int
}

func (s *StandardLogger) DebugWriter() io.Writer {
	return s.logger.WriterLevel(logrus.DebugLevel)
}
func DebugWriter() io.Writer {
	return logger.DebugWriter()
}

func (s *StandardLogger) InfoWriter() io.Writer {
	return s.logger.WriterLevel(logrus.InfoLevel)
}
func InfoWriter() io.Writer {
	return logger.InfoWriter()
}

func (s *StandardLogger) WarnWriter() io.Writer {
	return s.logger.WriterLevel(logrus.WarnLevel)
}
func WarnWriter() io.Writer {
	return logger.WarnWriter()
}

func (s *StandardLogger) ErrorWriter() io.Writer {
	return s.logger.WriterLevel(logrus.ErrorLevel)
}
func ErrorWriter() io.Writer {
	return logger.ErrorWriter()
}

func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func (s *StandardLogger) Debugf(format string, args ...interface{}) {
	_, file, line, ok := runtime.Caller(s.callerSkip)
	if ok {
		s.logger.Debugf("%s:%d :"+format, append([]interface{}{file, line}, args...)...)
	} else {
		s.logger.Debugf(format, args...)
	}
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}
func (s *StandardLogger) Debug(args ...interface{}) {
	_, file, line, ok := runtime.Caller(s.callerSkip)
	if ok {
		s.logger.Debug(append([]interface{}{fmt.Sprintf("%s:%d :", file, line)}, args...)...)
	} else {
		s.logger.Debug(args...)
	}
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}
func (s *StandardLogger) Infof(format string, args ...interface{}) {
	s.logger.Infof(format, args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}
func (s *StandardLogger) Info(args ...interface{}) {
	s.logger.Info(args...)
}

func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

func (s *StandardLogger) Warnf(format string, args ...interface{}) {
	s.logger.Warnf(format, args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}
func (s *StandardLogger) Warn(args ...interface{}) {
	s.logger.Warn(args...)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

func (s *StandardLogger) Errorf(format string, args ...interface{}) {
	s.logger.Errorf(format, args...)
}
func Error(args ...interface{}) {
	logger.Error(args...)
}
func (s *StandardLogger) Error(args ...interface{}) {
	s.logger.Error(args...)
}
