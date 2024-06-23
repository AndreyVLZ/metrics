package shutdown_test

import (
	"context"
	"fmt"
	"syscall"
	"time"

	"github.com/AndreyVLZ/metrics/agent"
	"github.com/AndreyVLZ/metrics/pkg/log"
	"github.com/AndreyVLZ/metrics/pkg/shutdown"
	"github.com/AndreyVLZ/metrics/server"
	"github.com/AndreyVLZ/metrics/server/adapter"
)

func ExampleShutdown_agent() {
	ctx := context.Background()
	log := log.New(log.SlogKey, log.LevelErr)
	agent := agent.New(log, agent.SetAddr("localhost:8081"))
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
	log := log.New(log.SlogKey, log.LevelErr)
	server := server.New(log, server.SetAddr("localhost:8082"), server.SetStorePath(""))
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
