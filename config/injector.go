package config

import (
	"errors"
	"github.com/kdaxx/common/errs"
	"github.com/kdaxx/container/v3"
	"github.com/spf13/viper"
	"reflect"
)

func NewInjector() *Injector {
	return &Injector{}
}

type Injector struct {
	container container.ProcessableContainer
}

func (i *Injector) SetContainer(container container.ProcessableContainer) {
	i.container = container
}

func (i *Injector) Initialize() error {
	viper.SetConfigFile(AppConfigFileName)
	viper.SetConfigType("yaml")

	// empty value treat to nil
	viper.AllowEmptyEnv(false)

	if err := viper.ReadInConfig(); err != nil {
		if errors.As(err, &viper.ConfigFileNotFoundError{}) {
			// Config file not found; ignore error if desired
			return errs.Newf("config file [%s] not found: %v\n", AppConfigFileName, err)
		} else {
			// Config file was found but another error was produced
			return errs.Newf("read config file [%s] failed: %v\n", AppConfigFileName, err)
		}
	}
	configs, ok := i.container.GetBeanByType(reflect.TypeFor[Configuration]())
	if !ok {
		return nil
	}
	for _, config := range configs {
		configuration := config.Bean().(Configuration)
		err := viper.UnmarshalKey(configuration.Prefix(), configuration)
		if err != nil {
			return errs.Newf("inject %s to config bean %v failed! %v\n",
				configuration.Prefix(), reflect.TypeOf(configuration), err)
		}
	}
	return nil
}
