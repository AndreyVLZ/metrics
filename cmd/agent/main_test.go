package main

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/AndreyVLZ/metrics/agent/config"
)

func TestRunAgent(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()

	ctxTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	cfg := config.Default()

	chErr := make(chan error)

	go func() {
		if err := runAgent(cfg, slog.Default()); err != nil {
			chErr <- err
		}
	}()

	select {
	case <-ctxTimeout.Done():
	case err := <-chErr:
		if err != nil {
			t.Errorf("run agent err: %v", err)
		}
	}
}
