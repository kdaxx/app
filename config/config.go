package config

// Configuration will be injected config with specified Prefix to ConfigurationBean
type Configuration interface {
	// Prefix returns `application.yaml` prefix
	Prefix() string
}

const (
	AppConfigFileName = "application.yaml"
)
