package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/AndreyVLZ/metrics/agent"
	arg "github.com/AndreyVLZ/metrics/internal/argument"
	"github.com/AndreyVLZ/metrics/internal/store/memstore"
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
	ctx := context.Background()

	agent := agent.New(cfg, memstore.New())

	ctxSinal, stopSignal := signal.NotifyContext(ctx, os.Interrupt)
	chErr := make(chan error)

	go func(ce chan<- error) {
		defer close(ce)
		ce <- agent.Start(ctxSinal)
	}(chErr)

	select {
	case <-ctxSinal.Done():
		log.Println("signal")
	case err := <-chErr:
		stopSignal()

		if err != nil {
			log.Printf("agent start err %v\n", err)
		}
	}

	if err := agent.Stop(ctx); err != nil {
		log.Printf("agent stop err: %v\n", err)
	} else {
		log.Println("all services stopped")
	}

	log.Println("agent stopped")
}
