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
	a := &App{
		container: inject.NewInjectContainer(),
	}
	a.injector = NewInjector(a.container)
	return a
}

type App struct {
	container container.ProcessableContainer
	injector  *Injector
}

// AddConfigPath add the directory path containing the configuration file (ConfigFileName).
func (a *App) AddConfigPath(path string) {
	a.injector.AddConfigPath(path)
}

func (a *App) RunApplication(ctx context.Context) error {
	logger.Info("application start")

	appContext, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := a.injectConfig(); err != nil {
		return err
	}

	// apply ContainerAware
	a.injectContainer()

	if err := a.applyPreparers(); err != nil {
		return err
	}
	// inject dependencies
	if err := a.container.Process(); err != nil {
		return err
	}

	if err := a.applyInitializers(); err != nil {
		return err
	}
	if err := a.runApplications(); err != nil {
		return err
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-quit:
	case <-appContext.Done():
	}

	a.stopApplications()
	return appContext.Err()
}

func (a *App) injectConfig() error {
	return a.injector.InjectConfig()
}

func (a *App) applyInitializers() error {

	// find all Initializer
	beanDefinitions, ok := a.container.GetBeanByType(
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

		initAdapter := newSeqAdapter()

		pre, ok := bean.(PreInitializer)
		if ok {
			initAdapter.preInitializer = pre
		}

		post, ok := bean.(PostInitializer)
		if ok {
			initAdapter.postInitializer = post
		}

		nodes = append(nodes, &InitNode{
			Bean: beanDefinition,
			Init: initAdapter,
			Name: beanDefinition.Type().String(),
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

func (a *App) applyPreparers() error {
	beanDefinitions, ok := a.container.GetBeanByType(reflect.TypeFor[Preparer]())
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

func (a *App) injectContainer() {
	beanDefinitions, ok := a.container.GetBeanByType(reflect.TypeFor[ContainerAware]())
	if !ok {
		return
	}
	for _, beanDefinition := range beanDefinitions {
		initializer := beanDefinition.Bean().(ContainerAware)
		initializer.SetContainer(a.container)
	}
}

func (a *App) runApplications() error {
	beanDefinitions, ok := a.container.GetBeanByType(reflect.TypeFor[Runnable]())
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
func (a *App) stopApplications() {
	defer func() {
		logger.Info("application stopped")
	}()
	beanDefinitions, ok := a.container.GetBeanByType(reflect.TypeFor[Stoppable]())
	if !ok {
		return
	}

	latch := task.NewLatch()
	latch.Add(len(beanDefinitions))

	var wait = 5
	logger.Infof("a will be stopped in %d seconds", wait)

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

func (a *App) Enable(registrars ...container.BeanRegistrar) {
	a.container.ApplyRegistrar(registrars...)
}

func newSeqAdapter() *seqAdapter {
	return &seqAdapter{}
}

type seqAdapter struct {
	preInitializer  PreInitializer
	postInitializer PostInitializer
}

func (s *seqAdapter) Before() []reflect.Type {
	if s.preInitializer != nil {
		return s.preInitializer.InitializeBefore()
	}
	return []reflect.Type{}
}

func (s *seqAdapter) After() []reflect.Type {
	if s.postInitializer != nil {
		return s.postInitializer.InitializeAfter()
	}
	return nil
}
