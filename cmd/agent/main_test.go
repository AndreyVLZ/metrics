package main

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/AndreyVLZ/metrics/agent"
)

func TestRunAgent(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()

	ctxTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	opts := []agent.FuncOpt{
		agent.SetPollInterval(1),
		agent.SetReportInterval(2),
		agent.SetRateLimit(10),
		agent.SetKey("key"),
	}

	chErr := make(chan error)

	go func() {
		if err := runAgent(ctxTimeout, 3, slog.Default(), opts...); err != nil {
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

	t.Log("OK-3")
}
