package server

import (
	"context"
	"testing"
	"time"

	"github.com/AndreyVLZ/metrics/pkg/log"
	"github.com/AndreyVLZ/metrics/server/config"
)

func TestServerStartStop(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()

	// Контекст имитации работы сервера.
	ctxTimeout, cancelTimeout := context.WithTimeout(ctx, 2*time.Second)
	defer cancelTimeout()

	// Контекст остановки сервера.
	ctxStopTimeout, cancelStopTimeout := context.WithTimeout(ctx, 5*time.Second)
	defer cancelStopTimeout()

	log := log.New(log.SlogKey, log.LevelErr)

	cfg, err := config.New(
		config.SetStorePath(""),
	)

	if err != nil {
		t.Errorf("new config: %v\n", err)
	}

	srv := New(cfg, log)

	t.Cleanup(func() {
		if err := srv.Stop(ctxStopTimeout); err != nil {
			t.Errorf("stop server err: %v\n", err)
		}
	})

	chErr := make(chan error)
	go func() {
		defer close(chErr)

		if err := srv.Start(ctx); err != nil {
			chErr <- err
		}
	}()

	select {
	case <-ctxTimeout.Done():
	case err := <-chErr:
		t.Errorf("start server err: %v\n", err)
	}
}
