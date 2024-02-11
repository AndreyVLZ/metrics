package main

import (
	"flag"
	"log"

	"github.com/AndreyVLZ/metrics/cmd/agent/metricagent"
	arg "github.com/AndreyVLZ/metrics/internal/argument"
)

func main() {
	var (
		addr           = metricagent.AddressDefault
		rateLimit      = metricagent.RateLimitDefault
		pollInterval   = metricagent.PollIntervalDefault
		reportInterval = metricagent.ReportIntervalDefault
		key            = ""
	)

	args := arg.Array(
		arg.String(
			&addr,
			"a",
			"ADDRESS",
			"адрес эндпоинта HTTP-сервера",
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
			log.Println(err)
		}
	}

	flag.Parse()

	client := metricagent.New(
		metricagent.SetAddr(addr),
		metricagent.SetPollInterval(pollInterval),
		metricagent.SetReportInterval(reportInterval),
		metricagent.SetKey(key),
		metricagent.SetRateLimit(rateLimit),
	)

	client.Start()
}
