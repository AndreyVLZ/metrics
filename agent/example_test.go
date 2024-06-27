package agent_test

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/AndreyVLZ/metrics/agent"
	"github.com/AndreyVLZ/metrics/agent/config"
)

func ExampleNew() {
	ctx := context.Background()

	// Контекст запуска агента.
	ctxStart, cancelStart := context.WithCancel(ctx)
	defer cancelStart()

	// Контекст имитации работы агента.
	ctxTimeout, cancelTimeout := context.WithTimeout(ctx, 2*time.Second)
	defer cancelTimeout()

	// Контекст остановки агента.
	ctxStop, cancelStop := context.WithTimeout(ctx, 5*time.Second)
	defer cancelStop()

	cfg := config.Default()

	agent := agent.New(cfg, slog.Default())

	// Функция agent.Start неблокирующая.
	// Ошибки при работе Агента идут в agent.Err
	if err := agent.Start(ctxStart); err != nil {
		fmt.Printf("start Agent err: %v\n", err)
	}

	select {
	case <-ctxTimeout.Done(): // по прошествии timeout отменяем работу агента.
		cancelStart()
	case err := <-agent.Err(): // читаем ошибки, которые могут возникнуть при работе агента.
		fmt.Printf("run Agent err: %v\n", err)
	}

	if err := agent.Stop(ctxStop); err != nil {
		fmt.Printf("stop Agent err: %v\n", err)
	}

	// Output:
}
