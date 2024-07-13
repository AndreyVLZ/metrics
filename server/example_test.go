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

	// Контекс запуска сервера
	ctxStart, cancelStart := context.WithTimeout(ctx, 2*time.Second)
	defer cancelStart()

	// Контекст остановки сервера.
	ctxStop, cancelStop := context.WithTimeout(ctx, 2*time.Second)
	defer cancelStop()

	log := log.New(log.SlogKey, log.LevelErr)

	cfg := config.Default()

	srv := server.New(cfg, log)

	chErr := make(chan error)
	go func() {
		defer close(chErr)
		chErr <- srv.Start(ctxStart)
	}()

	select {
	case <-ctxStart.Done():
	case err := <-chErr:
		fmt.Printf("start server err: %v\n", err)
	}

	if err := srv.Stop(ctxStop); err != nil {
		fmt.Printf("stop server err: %v\n", err)
	}

	// Output:
}
