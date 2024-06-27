package shutdown_test

import (
	"context"
	"fmt"
	"log/slog"
	"syscall"
	"time"

	"github.com/AndreyVLZ/metrics/agent"
	acfg "github.com/AndreyVLZ/metrics/agent/config"
	"github.com/AndreyVLZ/metrics/pkg/shutdown"
	"github.com/AndreyVLZ/metrics/server"
	"github.com/AndreyVLZ/metrics/server/adapter"
	scfg "github.com/AndreyVLZ/metrics/server/config"
)

func ExampleShutdown_agent() {
	ctx := context.Background()

	cfg, err := acfg.New(acfg.SetAddr("localhost:8081"))
	if err != nil {
		fmt.Println(err)
	}

	agent := agent.New(cfg, slog.Default())
	shuwdown := shutdown.New(agent, 2*time.Second)

	// Имитация сигнала прерывания.
	go func() {
		time.Sleep(2 * time.Second)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	if err := shuwdown.Start(ctx); err != nil {
		fmt.Println(err)
	}

	// Output:
	// context canceled
}

func ExampleShutdown_server() {
	ctx := context.Background()

	cfg, err := scfg.New(scfg.SetAddr("localhost:8082"), scfg.SetStorePath(""))
	if err != nil {
		fmt.Println(err)
	}

	server := server.New(cfg, slog.Default())
	shuwdown := shutdown.New(adapter.NewShutdown(&server), 2*time.Second)

	// Имитация сигнала прерывания.
	go func() {
		time.Sleep(2 * time.Second)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	if err := shuwdown.Start(ctx); err != nil {
		fmt.Println(err)
	}

	// Output:
	// context canceled
}
