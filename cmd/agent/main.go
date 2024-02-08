package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/AndreyVLZ/metrics/cmd/agent/metricagent"
)

func main() {
	addrPtr := flag.String("a", metricagent.AddressDefault, "адрес эндпоинта HTTP-сервера")
	pollIntervarPtr := flag.Int("p", metricagent.PollIntervalDefault, "частота опроса метрик из пакета runtime")
	reportIntervarPtr := flag.Int("r", metricagent.ReportIntervalDefault, "частота отправки метрик на сервер")
	keyPtr := flag.String("k", "", "ключ")
	rateLimitPtr := flag.Int("l", 1, "количество одновременно исходящих запросов на сервер")
	flag.Parse()

	opts := []metricagent.FuncOpt{}
	opts = append(opts,
		metricagent.SetAddr(*addrPtr),
		metricagent.SetPollInterval(*pollIntervarPtr),
		metricagent.SetReportInterval(*reportIntervarPtr),
		metricagent.SetKey(*keyPtr),
		metricagent.SetRateLimit(*rateLimitPtr),
	)

	if v, ok := os.LookupEnv("ADDRESS"); ok {
		opts = append(opts, metricagent.SetAddr(v))
	}

	if v, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		if ri, err := strconv.Atoi(v); err == nil {
			opts = append(opts, metricagent.SetReportInterval(ri))
		}
	}

	if v, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		if pi, err := strconv.Atoi(v); err == nil {
			opts = append(opts, metricagent.SetPollInterval(pi))
		}
	}

	if keyENV, ok := os.LookupEnv("KEY"); ok {
		opts = append(opts, metricagent.SetKey(keyENV))
	}

	if rateLimitStr, ok := os.LookupEnv("RATE_LIMIT"); ok {
		if rateLimitInt, err := strconv.Atoi(rateLimitStr); err == nil {
			opts = append(opts, metricagent.SetRateLimit(rateLimitInt))
		}
	}

	client := metricagent.New(opts...)

	client.Start()
}
