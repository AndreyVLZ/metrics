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
		key            = ""
	)

	args := arg.Array(
		arg.String(&addr, "ADDRESS"),
		arg.Int(&pollInterval, "POLL_INTERVAL"),
		arg.Int(&reportInterval, "REPORT_INTERVAL"),
		arg.String(&key, "KEY"),
		arg.Int(&rateLimit, "RATE_LIMIT"),
	)

	flag.StringVar(&addr, "a", addr, "адрес эндпоинта HTTP-сервера")
	flag.IntVar(&pollInterval, "p", pollInterval, "частота опроса метрик из пакета runtime")
	flag.IntVar(&reportInterval, "r", reportInterval, "частота отправки метрик на сервер")
	flag.StringVar(&key, "k", key, "ключ")
	flag.IntVar(&rateLimit, "l", rateLimit, "количество одновременно исходящих запросов на сервер")

	flag.Parse()

	for i := range args {
		err := args[i]()
		if err != nil {
			log.Printf("err parse args: %v\n", err)
		}
	}

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
