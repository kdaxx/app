package logger

const (
	DebugLevel = "debug"
	InfoLevel  = "info"
	WarnLevel  = "warn"
	ErrorLevel = "error"
)

func NewFileConfig() *FileConfig {
	return &FileConfig{
		Level:           InfoLevel,
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
	MaxBackups      int    `mapstructure:"max-backups"`
	MaxReservedDays int    `mapstructure:"max-reserved-days"`
	MaxReservedMB   int    `mapstructure:"max-reserved-mb"`
	Compress        bool   `mapstructure:"compress"`
}

func (c *FileConfig) Prefix() string {
	return "log"
}
