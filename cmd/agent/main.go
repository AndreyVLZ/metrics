package main

import (
	"context"
	"flag"
	"log/slog"
	"time"

	"github.com/AndreyVLZ/metrics/agent"
	"github.com/AndreyVLZ/metrics/pkg/env"
	mylog "github.com/AndreyVLZ/metrics/pkg/log"
	"github.com/AndreyVLZ/metrics/pkg/shutdown"
)

const timeout time.Duration = 7

func main() {
	var (
		addr           = agent.AddressDefault
		rateLimit      = agent.RateLimitDefault
		pollInterval   = agent.PollIntervalDefault
		reportInterval = agent.ReportIntervalDefault
		key            = ""
		logLevel       = mylog.LevelErr
	)

	flag.StringVar(&addr, "a", addr, "адрес эндпоинта HTTP-сервера")
	flag.IntVar(&pollInterval, "p", pollInterval, "частота опроса метрик из пакета runtime")
	flag.IntVar(&reportInterval, "r", reportInterval, "частота отправки метрик на сервер")
	flag.StringVar(&key, "k", key, "ключ")
	flag.IntVar(&rateLimit, "l", rateLimit, "количество одновременно исходящих запросов на сервер")
	flag.StringVar(&logLevel, "lvl", logLevel, "уровень логирования")

	flag.Parse()

	ctx := context.Background()
	logger := mylog.New(mylog.SlogKey, logLevel)

	vars := env.Array(
		env.String(&addr, "ADDRESS"),
		env.Int(&pollInterval, "POLL_INTERVAL"),
		env.Int(&reportInterval, "REPORT_INTERVAL"),
		env.String(&key, "KEY"),
		env.Int(&rateLimit, "RATE_LIMIT"),
	)

	for i := range vars {
		if err := vars[i](); err != nil {
			logger.DebugContext(ctx, "env", "info", err)
		}
	}

	opts := []agent.FuncOpt{
		agent.SetAddr(addr),
		agent.SetPollInterval(pollInterval),
		agent.SetReportInterval(reportInterval),
		agent.SetRateLimit(rateLimit),
		agent.SetKey(key),
	}

	if err := runAgent(ctx, timeout, logger, opts...); err != nil {
		logger.ErrorContext(ctx, "start agent", "error", err)
	}
}

func runAgent(ctx context.Context, timeout time.Duration, log *slog.Logger, opts ...agent.FuncOpt) error {
	agent := agent.New(log, opts...)
	shutdown := shutdown.New(agent, timeout)

	if err := shutdown.Start(ctx); err != nil {
		return err
	}

	return nil
}
