package unidireaction

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

// Perm   Menu
// |\     /|
// | \   / |
// |  v v  |
// |  Role |
// |   v   |
// |  User |
// |   v   |
//
//	->Web <-
type Role struct {
	perm *Perm
	menu *Menu
	Init bool
}

func (a *Role) Inject() any {
	return func(perm *Perm, menu *Menu) {
		a.perm = perm
		a.menu = menu
	}
}
func (a *Role) InitializeAfter() []reflect.Type {
	return []reflect.Type{
		reflect.TypeFor[*Perm](),
		reflect.TypeFor[*Menu](),
	}
}

func (a *Role) InitializeBefore() []reflect.Type {
	return []reflect.Type{
		reflect.TypeFor[*Web](),
	}
}

func (a *Role) Initialize() error {
	fmt.Println("Role Init")
	a.Init = true
	if !(a.perm.Init && a.menu.Init) {
		return errors.New("perm & menu Init never before role")
	}
	return nil
}

type Perm struct {
	web  *Web
	Init bool
}

func (b *Perm) Inject() any {
	return func(d *Web) {
		b.web = d
	}
}

func (b *Perm) Initialize() error {
	fmt.Println("Perm Init")
	b.Init = true
	if b.web.Init {
		return errors.New("perm Init never before web")
	}
	return nil
}

func (*Perm) InitializeBefore() []reflect.Type {
	return []reflect.Type{
		reflect.TypeFor[*Web](),
	}
}

type Menu struct {
	web  *Web
	Init bool
}

func (c *Menu) Inject() any {
	return func(b *Web) {
		c.web = b
	}
}

func (c *Menu) Initialize() error {
	fmt.Println("Menu Init")
	c.Init = true
	if c.web.Init {
		return errors.New("web Init never after menu")
	}
	return nil
}
func (c *Menu) InitializeBefore() []reflect.Type {
	return []reflect.Type{
		reflect.TypeFor[*Web](),
	}
}

type User struct {
	Init bool
}

func (d *User) Initialize() error {
	fmt.Println("User Init")
	d.Init = true
	return nil
}

func (d *User) InitializeAfter() []reflect.Type {
	return []reflect.Type{
		reflect.TypeFor[*Role](),
	}
}

func (d *User) InitializeBefore() []reflect.Type {
	return []reflect.Type{
		reflect.TypeFor[*Web](),
	}
}

type Web struct {
	Init bool
}

func (w *Web) Initialize() error {
	fmt.Println("Web Init")
	w.Init = true
	return nil
}

type InitRegistrar struct {
}

func (i *InitRegistrar) ApplyRegistry(registry container.BeanRegistry) {
	registry.RegisterBean(&Perm{}, &Web{}, &Role{}, &Menu{}, &User{})
}

func TestInitializer(t *testing.T) {
	a := app.NewApp()
	a.Enable(&InitRegistrar{})
	abs, err := filepath.Abs("../testdata")
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
