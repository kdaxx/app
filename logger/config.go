package logger

import "github.com/sirupsen/logrus"

func NewFileConfig() *FileConfig {
	return &FileConfig{
		Level:           logrus.InfoLevel.String(),
		Format:          "2006-01-02-15-04-05.000",
		Filepath:        "log/app.log",
		MaxBackups:      10,
		MaxReservedDays: 15,
		MaxReservedMB:   10,
		Compress:        false,
	}
}

type FileConfig struct {
	Filepath        string `mapstructure:"filepath"`
	Level           string `mapstructure:"level"`
	Format          string `mapstructure:"format"`
	MaxBackups      int    `mapstructure:"max-backups"`
	MaxReservedDays int    `mapstructure:"max-reserved-days"`
	MaxReservedMB   int64  `mapstructure:"max-reserved-mb"`
	Compress        bool   `mapstructure:"compress"`
}

func (c *FileConfig) Prefix() string {
	return "log"
}

func NewAppLogger() *AppLogger {
	return &AppLogger{}
}

type AppLogger struct {
}

func (a *AppLogger) Bootstrap() any {
	return func(config *FileConfig) {
		Override(NewStandardLogger(&Config{
			Level:      config.Level,
			Format:     config.Format,
			Filepath:   config.Filepath,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxReservedDays,
			MaxBytes:   1024 * 1024 * config.MaxReservedMB,
			Compress:   config.Compress,
		}, 2))
	}
}
