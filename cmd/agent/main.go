package main

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/AndreyVLZ/metrics/cmd/agent/metricagent"
)

func main() {
	addrPtr := flag.String("a", metricagent.AddressDefault, "адрес эндпоинта HTTP-сервера")
	pollIntervarPtr := flag.Int("p", metricagent.PollIntervalDefault, "частота опроса метрик из пакета runtime")
	reportIntervarPtr := flag.Int("r", metricagent.ReportIntervalDefault, "частота отправки метрик на сервер")
	flag.Parse()

	opts := []metricagent.FuncOpt{}
	fAddr := metricagent.SetAddr(*addrPtr)
	fPollInt := metricagent.SetPollInterval(*pollIntervarPtr)
	fReportInt := metricagent.SetReportInterval(*reportIntervarPtr)

	opts = append(opts, fAddr, fPollInt, fReportInt)

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

	client := metricagent.New(opts...)

	err := client.Start()
	if err != nil {
		log.Panicf("err: %v\n", err)
	}
}
