package agent_test

import (
	"context"

	"github.com/AndreyVLZ/metrics/agent"
	"github.com/AndreyVLZ/metrics/internal/log/slog"
	"github.com/AndreyVLZ/metrics/internal/store/inmemory"
)

func ExampleNew() {
	ctx := context.Background()
	log := slog.New()

	cfg := agent.NewConfig(
		agent.SetAddr("localhost:8080"),
		agent.SetPollInterval(10),
	)

	store := inmemory.New()
	agent := agent.New(cfg, store, log)

	if err := agent.Start(ctx); err != nil {
		log.Error("agent start", "err", err)
	}

	for err := range agent.Err() {
		log.Error("agent run", "err", err)
	}

	if err := agent.Stop(ctx); err != nil {
		log.Error("agent stop", "err", err)
	}

	// Output: nil
}
