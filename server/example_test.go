package server_test

import (
	"context"

	"github.com/AndreyVLZ/metrics/internal/log/slog"
	"github.com/AndreyVLZ/metrics/server"
)

func ExampleNew() {
	ctx := context.Background()
	log := slog.New()
	cfg := server.NewConfig(server.SetAddr("localhost:8080"))
	server := server.New(cfg, log)

	if err := server.Start(ctx); err != nil {
		log.Error("start server", "err", err)
	}

	if err := server.Stop(ctx); err != nil {
		log.Error("start server", "err", err)
	}

	// Output: nil
}
