package main

import (
	"flag"
	"log"

	"github.com/AndreyVLZ/metrics/cmd/agent/config"
	"github.com/AndreyVLZ/metrics/cmd/agent/metricagent"
)

func main() {
	conf := config.New()
	flag.StringVar(&conf.Addr, "a", "localhost:8080", "doc-1")
	flag.IntVar(&conf.ReportInterval, "r", 10, "doc-2")
	flag.IntVar(&conf.PollInterval, "p", 2, "doc1-3")
	flag.Parse()
	client := metricagent.New(conf)

	err := client.AddMetric("gauge", "meMetric", "123")
	if err != nil {
		log.Panicf("err: %v\n", err)
	}

	err = client.Start()
	if err != nil {
		log.Panicf("err: %v\n", err)
	}
}
