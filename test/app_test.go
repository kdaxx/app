package test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/kdaxx/app"
	"github.com/kdaxx/common/task"
	"github.com/kdaxx/container/v3"
)

func TestApp(t *testing.T) {
	a := app.NewApp()
	a.Enable(app.NewCore())
	abs, err := filepath.Abs("./testdata")
	if err != nil {
		log.Fatal(err)
	}
	a.AddConfigPath(abs)

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	latch := task.NewLatch()
	latch.Add(1)
	go func() {
		defer latch.Done()
		err := a.RunApplication(ctx)
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
	fmt.Println("loopA Init")
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
	fmt.Println("loopB Init")
	return nil
}

type LoopC struct {
}

func (l *LoopC) Initialize() error {
	fmt.Println("loopC Init")
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
	a := app.NewApp()
	a.Enable(&InitLoopRegistrar{})
	abs, err := filepath.Abs("./testdata")
	if err != nil {
		log.Fatal(err)
	}
	a.AddConfigPath(abs)
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	latch := task.NewLatch()
	latch.Add(1)
	go func() {
		defer latch.Done()
		err := a.RunApplication(ctx)
		if err == nil || errors.Is(err, context.Canceled) {
			t.Error("no loop detected")
		}
		t.Log(err)
	}()

	cancelFunc()
	<-latch.Wait()
	log.Println("done")
}
