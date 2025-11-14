package app

import (
	"github.com/kdaxx/app/config"
	"github.com/kdaxx/app/logger"
	"github.com/kdaxx/container/v3"
)

func NewCore() *Core {
	return &Core{}
}

// Core provides core components for application involved logging and configuration injection.
type Core struct {
}

// ApplyRegistry register app core components.
func (c *Core) ApplyRegistry(register container.BeanRegistry) {
	register.RegisterBean(
		// Injector enables the app to have automatic configuration injection.
		config.NewInjector(),

		// Logger enables app to record, rotate logs.
		NewFileConfig(),
		logger.NewFileConfig(),
		logger.NewAppLogger(),
	)
}
