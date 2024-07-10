package main

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/AndreyVLZ/metrics/server/config"
)

func TestServerRun(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()

	ctxStart, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	cfg := config.Default()

	log := slog.Default()
	chErr := make(chan error)

	go func() {
		defer close(chErr)

		if err := runServer(cfg, log); err != nil {
			chErr <- err
		}
	}()

	cancel()
	select {
	case <-ctxStart.Done():
	case err := <-chErr:
		if err != nil {
			t.Errorf("run agent err: %v", err)
		}
	}
}
