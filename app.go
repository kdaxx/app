package app

import (
	"context"
	"github.com/kdaxx/app/logger"
)

const (
	ConfigPrefix = "app"
)

const (
	Release string = "release"
	Dev     string = "dev"
)

func NewAppLogger() *Logger {
	return &Logger{}
}

type Logger struct {
}

func (a *Logger) Stop(ctx context.Context) error {
	return logger.Close()
}

func (a *Logger) Bootstrap() any {
	return func(config *logger.FileConfig, appConfig *FileConfig) {
		logger.Override(logger.NewStandardLogger(&logger.Config{
			Filename:    config.Filepath,
			LogLevel:    config.Level,
			MaxBackups:  config.MaxBackups,
			MaxAge:      config.MaxReservedDays,
			Compress:    config.Compress,
			MaxSize:     config.MaxReservedMB,
			ProductMode: appConfig.Mode == Release,
		}, 2))
	}
}
