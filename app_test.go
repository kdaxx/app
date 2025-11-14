package app

import (
	"context"
	"errors"
	"github.com/kdaxx/app/logger"
	"github.com/kdaxx/common/task"
	"testing"
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
	logger.Debug("done")
	logger.Info("done")
	logger.Warn("done")
}
