package server_test

import (
	"context"
	"fmt"
	"time"

	"github.com/AndreyVLZ/metrics/pkg/log"
	"github.com/AndreyVLZ/metrics/server"
	"github.com/AndreyVLZ/metrics/server/config"
)

func ExampleNew() {
	ctx := context.Background()

	// Контекст остановки сервера.
	ctxStop, cancelStop := context.WithTimeout(ctx, 5*time.Second)
	defer cancelStop()

	log := log.New(log.SlogKey, log.LevelErr)

	cfg, err := config.New(
		config.SetStorePath(""),
	)

	if err != nil {
		log.Error("new config", "err", err)
	}

	srv := server.New(cfg, log)

	chErr := make(chan error)
	go func() {
		defer close(chErr)
		chErr <- srv.Start(ctx)
	}()

	select {
	case <-time.After(2 * time.Second):
	case err := <-chErr:
		fmt.Printf("start server err: %v\n", err)
	}

	if err := srv.Stop(ctxStop); err != nil {
		fmt.Printf("stop server err: %v\n", err)
	}

	// Output:
}
