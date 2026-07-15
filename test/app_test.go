package test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"testing"

	APP "github.com/kdaxx/app"
	"github.com/kdaxx/common/task"
	"github.com/kdaxx/container/v3"
)

func TestApp(t *testing.T) {
	app := APP.NewApp()
	app.Enable(APP.NewCore())

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	latch := task.NewLatch()
	latch.Add(1)
	go func() {
		defer latch.Done()
		err := app.RunApplication(ctx)
		if !errors.Is(err, context.Canceled) {
			t.Error(err)
			return
		}

	}()

	cancelFunc()
	<-latch.Wait()
	log.Println("done")
}

// A -> C
// |  /
// v v
// B
// |
// v
// D
type InitA struct{}

func (a *InitA) InitializeAfter() []reflect.Type {
	return []reflect.Type{
		reflect.TypeFor[*InitC](),
		reflect.TypeFor[*InitB](),
	}
}

func (*InitA) Initialize() error {
	fmt.Println("A init")
	return nil
}

type InitB struct{}

func (*InitB) Initialize() error {
	fmt.Println("B init")
	return nil
}

func (*InitB) InitializeAfter() []reflect.Type {
	return []reflect.Type{
		reflect.TypeFor[*InitD](),
	}
}

type InitC struct {
}

func (*InitC) Initialize() error {
	fmt.Println("C init")
	return nil
}
func (a *InitC) InitializeAfter() []reflect.Type {
	return []reflect.Type{
		reflect.TypeFor[*InitB](),
	}
}

type InitD struct {
}

func (*InitD) Initialize() error {
	fmt.Println("D init")
	return nil
}

type InitRegistrar struct {
}

func (i *InitRegistrar) ApplyRegistry(registry container.BeanRegistry) {
	registry.RegisterBean(&InitA{}, &InitB{}, &InitC{}, &InitD{})
}

func TestInitializer(t *testing.T) {
	app := APP.NewApp()
	app.Enable(&InitRegistrar{})
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	latch := task.NewLatch()
	latch.Add(1)
	go func() {
		defer latch.Done()
		err := app.RunApplication(ctx)
		if !errors.Is(err, context.Canceled) {
			t.Error(err)
			return
		}

	}()

	cancelFunc()
	<-latch.Wait()
	log.Println("done")
}

// C -> A -> B -> C
type LoopA struct {
}

func (l *LoopA) Initialize() error {
	fmt.Println("loopA init")
	return nil
}

func (l *LoopA) InitializeBefore() []reflect.Type {
	return []reflect.Type{
		reflect.TypeFor[*LoopB](),
	}
}

type LoopB struct {
}

func (l *LoopB) InitializeBefore() []reflect.Type {
	return []reflect.Type{
		reflect.TypeFor[*LoopC](),
	}
}

func (l *LoopB) Initialize() error {
	fmt.Println("loopB init")
	return nil
}

type LoopC struct {
}

func (l *LoopC) Initialize() error {
	fmt.Println("loopC init")
	return nil
}

func (l *LoopC) InitializeBefore() []reflect.Type {
	return []reflect.Type{
		reflect.TypeFor[*LoopA](),
	}
}

type InitLoopRegistrar struct {
}

func (i *InitLoopRegistrar) ApplyRegistry(registry container.BeanRegistry) {
	registry.RegisterBean(&LoopA{}, &LoopB{}, &LoopC{})
}

func TestInitializerLoop(t *testing.T) {
	app := APP.NewApp()
	app.Enable(&InitLoopRegistrar{})
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	latch := task.NewLatch()
	latch.Add(1)
	go func() {
		defer latch.Done()
		err := app.RunApplication(ctx)
		if err == nil || errors.Is(err, context.Canceled) {
			t.Error("no loop detected")
		}
		t.Log(err)

	}()

	cancelFunc()
	<-latch.Wait()
	log.Println("done")
}
