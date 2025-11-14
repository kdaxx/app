package app

import (
	"context"
	"github.com/kdaxx/container/v3"
)

// ContainerAware allows beans to inject the app's container at runtime.
type ContainerAware interface {
	SetContainer(container container.ProcessableContainer)
}

// Preparer is an application-level lifecycle hook that is called before dependency injection.
// At this point, the beans being added to the container can still be automatically
// injected into the required beans. and actions that do not require dependencies can
// also be performed at this stage.
type Preparer interface {
	Prepare() error
}

// Initializer is application-level initialization hook, where dependency injection is performed.
// adding or removing container beans no longer affects beans that have already undergone dependency injection.
// some initialization tasks that requires dependencies can be run at this stage,
// but this does not mean that the bean has started working.
type Initializer interface {
	Initialize() error
}

// Runnable means that the bean can execute a task,
// and all initialization operations have been completed at this point, meaning the bean can begin executing the task.
type Runnable interface {
	Run() error
}

// Stoppable will be called before the application closes.
// if a task can perform a stop operation, it should implement it.
type Stoppable interface {
	Stop(ctx context.Context) error
}
