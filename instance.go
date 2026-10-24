package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/kdaxx/app/logger"
	"github.com/kdaxx/common/task"
	"github.com/kdaxx/container/v3"
	"github.com/kdaxx/container/v3/inject"
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

	if err := app.applyPreparers(); err != nil {
		return err
	}
	// inject dependencies
	if err := app.container.Process(); err != nil {
		return err
	}

	if err := app.applyInitializers(); err != nil {
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
func newBefore(preInitializer PreInitializer) Before {
	return &beforeAdapter{
		preInitializer: preInitializer,
	}
}

type beforeAdapter struct {
	preInitializer PreInitializer
}

func (b *beforeAdapter) Before() []reflect.Type {
	return b.preInitializer.InitializeBefore()
}

func newAfter(postInitializer PostInitializer) After {
	return &afterAdapter{
		postInitializer: postInitializer,
	}
}

type afterAdapter struct {
	postInitializer PostInitializer
}

func (b *afterAdapter) After() []reflect.Type {
	return b.postInitializer.InitializeAfter()
}

func (app *App) applyInitializers() error {

	// find all Initializer
	beanDefinitions, ok := app.container.GetBeanByType(
		reflect.TypeFor[Initializer](),
	)

	if !ok {
		return nil
	}

	// =========================
	// 1. create Init nodes
	// =========================

	nodes := make([]*InitNode, 0, len(beanDefinitions))

	for _, beanDefinition := range beanDefinitions {

		bean := beanDefinition.Bean()

		_, ok := bean.(Initializer)
		if !ok {
			continue
		}

		var initAdapter any

		pre, ok := bean.(PreInitializer)
		if ok {
			initAdapter = newBefore(pre)
		}

		post, ok := bean.(PostInitializer)
		if ok {
			initAdapter = newAfter(post)
		}

		nodes = append(nodes, &InitNode{
			Bean: beanDefinition,
			Init: initAdapter,
			Name: reflect.TypeOf(bean).String(),
		})
	}

	// =========================
	// 2. establish dependencies
	// =========================

	if err := buildInitializerGraph(nodes); err != nil {
		return err
	}

	// =========================
	// 3. DFS ordering
	// =========================

	order, err := topoSort(nodes)

	if err != nil {
		return err
	}

	// =========================
	// 4. initialize in order
	// =========================
	for _, node := range order {
		initializer := node.Bean.Bean().(Initializer)
		if err := initializer.Initialize(); err != nil {
			return fmt.Errorf(
				"initialize %s failed: %w",
				node.Name,
				err,
			)
		}
	}
	return nil
}

func (app *App) applyPreparers() error {
	beanDefinitions, ok := app.container.GetBeanByType(reflect.TypeFor[Preparer]())
	if !ok {
		return nil
	}
	for _, beanDefinition := range beanDefinitions {
		initializer := beanDefinition.Bean().(Preparer)
		err := initializer.Prepare()
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
