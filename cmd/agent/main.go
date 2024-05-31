package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/AndreyVLZ/metrics/agent"
	arg "github.com/AndreyVLZ/metrics/internal/argument"
	"github.com/AndreyVLZ/metrics/internal/log/zap"
	"github.com/AndreyVLZ/metrics/internal/shutdown"
	"github.com/AndreyVLZ/metrics/internal/store/inmemory"
)

func main() {
	var (
		addr           = agent.AddressDefault
		rateLimit      = agent.RateLimitDefault
		pollInterval   = agent.PollIntervalDefault
		reportInterval = agent.ReportIntervalDefault
		key            = agent.KeyDafault
	)

	args := arg.Array(
		arg.String(
			&addr,     // val
			"a",       // name Flag
			"ADDRESS", // name ENV
			"адрес эндпоинта HTTP-сервера", // desc
		),
		arg.Int(
			&pollInterval,
			"p",
			"POLL_INTERVAL",
			"частота опроса метрик из пакета runtime",
		),
		arg.Int(
			&reportInterval,
			"r",
			"REPORT_INTERVAL",
			"частота отправки метрик на сервер",
		),
		arg.String(
			&key,
			"k",
			"KEY",
			"ключ",
		),
		arg.Int(
			&rateLimit,
			"l",
			"RATE_LIMIT",
			"количество одновременно исходящих запросов на сервер",
		),
	)

	for i := range args {
		err := args[i]()
		if err != nil {
			log.Printf("err parse args: %v\n", err)
		}
	}

	flag.Parse()

	cfg := agent.NewConfig(
		agent.SetAddr(addr),
		agent.SetPollInterval(pollInterval),
		agent.SetReportInterval(reportInterval),
		agent.SetRateLimit(rateLimit),
		agent.SetKey(key),
	)

	run(cfg)
}

func run(cfg *agent.Config) {
	var timeout time.Duration = 7

	logger := zap.New(zap.DefaultConfig())
	agent := agent.New(cfg, inmemory.New(), logger)
	shutdown := shutdown.New(agent, timeout)

	if err := shutdown.Start(context.Background()); err != nil {
		log.Printf("start agent error: %v\n", err)
	}
}
