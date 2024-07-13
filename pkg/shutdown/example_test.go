package shutdown_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
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

	cfg := acfg.Default()
	cfg.SetOpts(
		acfg.SetAddr("localhost:8081"),
		acfg.SetAddGRPC(":3201"),
	)

	agent := agent.New(cfg, slog.Default())
	shuwdown := shutdown.New(agent, 2*time.Second)

	// Имитация сигнала прерывания.
	go func() {
		time.Sleep(2 * time.Second)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	signals := []os.Signal{
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	}

	if err := shuwdown.Start(ctx, signals...); err != nil {
		fmt.Println(err)
	}

	// Output:
	// context canceled
}

func ExampleShutdown_server() {
	ctx := context.Background()

	cfg := scfg.Default()
	cfg.SetOpts(
		scfg.SetAddr("localhost:8082"),
		scfg.SetAddGRPC(":3202"),
	)

	server := server.New(cfg, slog.Default())
	shuwdown := shutdown.New(adapter.NewShutdown(&server), 2*time.Second)

	signals := []os.Signal{
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	}

	// Имитация сигнала прерывания.
	go func() {
		time.Sleep(2 * time.Second)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	if err := shuwdown.Start(ctx, signals...); err != nil {
		fmt.Println(err)
	}

	// Output:
	// context canceled
}
