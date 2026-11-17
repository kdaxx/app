package app

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/kdaxx/common/errs"
	"github.com/kdaxx/container/v3"
	"github.com/spf13/viper"
)

func NewInjector(container container.ProcessableContainer) *Injector {
	return &Injector{
		container: container,
	}
}

type Injector struct {
	container container.ProcessableContainer
}

func (i *Injector) SetContainer(container container.ProcessableContainer) {
	i.container = container
}

func (i *Injector) AddConfigPath(path string) {
	viper.AddConfigPath(path)
}
func (i *Injector) InjectConfig() error {
	if err := i.readInConfig(); err != nil {
		return err
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

func (i *Injector) readInConfig() error {
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")
	// empty value treat to nil
	viper.AllowEmptyEnv(false)

	// reading config
	viper.SetConfigName(ConfigFileName)
	if err := viper.ReadInConfig(); err != nil {
		if errors.As(err, &viper.ConfigFileNotFoundError{}) {
			// Config file not found.
			return errs.Newf("config file [%s] not found: %v\n", ConfigFileName, err)
		}

		// Config file was found but another error was produced
		return errs.Newf("read config file [%s] failed: %v\n", ConfigFileName, err)
	}

	appMode := viper.GetString("app.profile")
	if appMode != "" {
		viper.SetConfigName(fmt.Sprintf("%s-%s.yaml", ConfigFileName, appMode))
		if err := viper.MergeInConfig(); err != nil {
			if errors.As(err, &viper.ConfigFileNotFoundError{}) {
				// Config file not found; ignore error if desired.
				return nil
			}

			// Config file was found but another error was produced
			return errs.Newf("read config file [%s-%s] failed: %v\n", ConfigFileName, appMode, err)
		}
	}
	return nil
}
