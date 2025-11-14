package app

import (
	"context"
	"github.com/kdaxx/container/v3"
)

// ContainerAware allows beans to inject the app's container at runtime.
type ContainerAware interface {
	SetContainer(container container.ProcessableContainer)
}

// Initializer is application-level initialization hook, where dependency injection is not yet performed,
// and allow you to add or remove beans from the app container,
// or perform initialization actions that do not require dependencies.
type Initializer interface {
	Initialize() error
}

// Runnable is to perform a task.
// it is recommended to initialize pre-startup initializations here,
// as all dependencies have been injected at this stage,
// making it the most stable period for application execution.
type Runnable interface {
	Run() error
}

// Stoppable will be called before the application closes.
// if a task can perform a stop operation, it should implement it.
type Stoppable interface {
	Stop(ctx context.Context) error
}
