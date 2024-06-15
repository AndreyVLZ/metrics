package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AndreyVLZ/metrics/pkg/log"
	"github.com/AndreyVLZ/metrics/server"
)

func TestServerRun(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()

	ctxStart, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	opts := []server.FuncOpt{
		server.SetStorePath(""),
		server.SetRestore(false),
	}

	log := log.New(log.SlogKey, log.LevelErr)
	chErr := make(chan error)

	go func() {
		defer close(chErr)

		if err := runServer(ctxStart, 2, log, opts...); err != nil {
			chErr <- err
		}
	}()

	cancel()

	if err := <-chErr; err != nil && !errors.Is(err, context.Canceled) {
		t.Errorf("run server err: %v", err)
	}
}
