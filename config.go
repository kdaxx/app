package app

func NewFileConfig() *FileConfig {
	return &FileConfig{}
}

type FileConfig struct {
	Mode string `mapstructure:"mode"`
}

func (c *FileConfig) Prefix() string {
	return ConfigPrefix
}
