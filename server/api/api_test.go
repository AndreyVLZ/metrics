package api

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/AndreyVLZ/metrics/internal/store"
	"github.com/AndreyVLZ/metrics/server/config"
)

func TestStartStop(t *testing.T) {
	ctx := context.Background()

	ctxTimeout, cancelTimeout := context.WithTimeout(ctx, 2*time.Second)
	defer cancelTimeout()

	ctxStopTimeout, cancelStopTimeout := context.WithTimeout(ctx, 2*time.Second)
	defer cancelStopTimeout()

	cfg := config.Default()
	cfg.SetOpts(
		config.SetAddr("localhost:8083"),
		config.SetAddGRPC(":3203"),
	)

	server := New(cfg, store.New(cfg.StorageConfig), slog.Default())

	chErr := make(chan error)
	go func(srv *Server, ce chan<- error) {
		if err := srv.Start(); err != nil {
			ce <- err
		}
	}(server, chErr)

	select {
	case <-ctxTimeout.Done():
	case err := <-chErr:
		if err != nil {
			t.Errorf("err start: %v\n", err)
		}
	}

	if err := server.Stop(ctxStopTimeout); err != nil {
		t.Errorf("err stop: %v\n", err)
	}
}
