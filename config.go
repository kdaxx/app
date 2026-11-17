package app

// Configuration will be injected config with specified Prefix to ConfigurationBean
type Configuration interface {
	// Prefix returns `application.yaml` prefix
	Prefix() string
}

const (
	ConfigFileName = "application"
)

func NewFileConfig() *FileConfig {
	return &FileConfig{}
}

type FileConfig struct {
	Mode string `mapstructure:"mode"`
}

func (c *FileConfig) Prefix() string {
	return ConfigPrefix
}
