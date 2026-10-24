package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/kdaxx/common/task"
	"github.com/kdaxx/container/v3"
)

func TestApp(t *testing.T) {
	app := NewApp()
	app.Enable(NewCore())

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
type InitA struct {
	c    *InitC
	b    *InitB
	Init bool
}

func (a *InitA) Inject() any {
	return func(c *InitC, b *InitB) {
		a.c = c
		a.b = b
	}
}
func (a *InitA) InitializeAfter() []reflect.Type {
	return []reflect.Type{
		reflect.TypeFor[*InitC](),
		reflect.TypeFor[*InitB](),
	}
}

func (a *InitA) Initialize() error {
	fmt.Println("A Init")
	a.Init = true
	if !(a.c.Init && a.b.Init) {
		return errors.New("b & c Init before a")
	}
	return nil
}

type InitB struct {
	d    *InitD
	Init bool
}

func (b *InitB) Inject() any {
	return func(d *InitD) {
		b.d = d
	}
}

func (b *InitB) Initialize() error {
	fmt.Println("B Init")
	b.Init = true
	if !b.d.Init {
		return errors.New("d Init before b")
	}
	return nil
}

func (*InitB) InitializeAfter() []reflect.Type {
	return []reflect.Type{
		reflect.TypeFor[*InitD](),
	}
}

type InitC struct {
	b    *InitB
	Init bool
}

func (c *InitC) Inject() any {
	return func(b *InitB) {
		c.b = b
	}
}

func (c *InitC) Initialize() error {
	fmt.Println("C Init")
	c.Init = true
	if !c.b.Init {
		return errors.New("b Init before c")
	}
	return nil
}
func (c *InitC) InitializeAfter() []reflect.Type {
	return []reflect.Type{
		reflect.TypeFor[*InitB](),
	}
}

type InitD struct {
	Init bool
}

func (d *InitD) Initialize() error {
	fmt.Println("D Init")
	d.Init = true
	return nil
}

type InitRegistrar struct {
}

func (i *InitRegistrar) ApplyRegistry(registry container.BeanRegistry) {
	registry.RegisterBean(&InitA{}, &InitB{}, &InitC{}, &InitD{})
}

func TestInitializer(t *testing.T) {
	app := NewApp()
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
	app := NewApp()
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
