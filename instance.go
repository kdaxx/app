package app

import (
	"context"
	"github.com/kdaxx/app/logger"
	"github.com/kdaxx/common/task"
	"github.com/kdaxx/container/v3"
	"github.com/kdaxx/container/v3/inject"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"
)

func NewApp() *App {
	return &App{
		container: inject.NewInjectContainer(),
	}
}

type App struct {
	container container.ProcessableContainer
}

func (app *App) RunApplication(ctx context.Context) error {
	logger.Info("application start")

	appContext, cancel := context.WithCancel(ctx)
	defer cancel()
	// apply ContainerAware
	app.injectContainer()

	if err := app.applyInitializers(); err != nil {
		return err
	}

	if err := app.container.Process(); err != nil {
		return err
	}

	if err := app.runApplications(); err != nil {
		return err
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-quit:
	case <-appContext.Done():
	}

	app.stopApplications()
	return appContext.Err()
}

func (app *App) applyInitializers() error {
	beanDefinitions, ok := app.container.GetBeanByType(reflect.TypeFor[Initializer]())
	if !ok {
		return nil
	}
	for _, beanDefinition := range beanDefinitions {
		initializer := beanDefinition.Bean().(Initializer)
		err := initializer.Initialize()
		if err != nil {
			return err
		}
	}
	return nil
}

func (app *App) injectContainer() {
	beanDefinitions, ok := app.container.GetBeanByType(reflect.TypeFor[ContainerAware]())
	if !ok {
		return
	}
	for _, beanDefinition := range beanDefinitions {
		initializer := beanDefinition.Bean().(ContainerAware)
		initializer.SetContainer(app.container)
	}
}

func (app *App) runApplications() error {
	beanDefinitions, ok := app.container.GetBeanByType(reflect.TypeFor[Runnable]())
	if !ok {
		return nil
	}
	for _, beanDefinition := range beanDefinitions {
		runnable := beanDefinition.Bean().(Runnable)
		err := runnable.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

// forced exit (e.g., calling Close(), or directly exiting the main goroutine) does only "stop waiting",
// but it does not mean that all goroutines have been released or exited.
func (app *App) stopApplications() {
	defer func() {
		logger.Info("application stopped")
	}()
	beanDefinitions, ok := app.container.GetBeanByType(reflect.TypeFor[Stoppable]())
	if !ok {
		return
	}

	latch := task.NewLatch()
	latch.Add(len(beanDefinitions))

	var wait = 5
	logger.Infof("app will be stopped in %d seconds", wait)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(wait)*time.Second)
	defer cancel()

	for _, beanDefinition := range beanDefinitions {
		go func() {
			defer latch.Done()
			runnable := beanDefinition.Bean().(Stoppable)
			err := runnable.Stop(ctx)
			if err != nil {
				logger.Warn(err)
			}
		}()
	}

	select {
	case <-ctx.Done():
	case <-latch.Wait():
	}

}

func (app *App) Enable(registrars ...container.BeanRegistrar) {
	app.container.ApplyRegistrar(registrars...)
}
