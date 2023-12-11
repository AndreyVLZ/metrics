package main

import (
	"log"

	"github.com/AndreyVLZ/metrics/cmd/agent/metricagent"
)

func main() {
	client := metricagent.New()

	err := client.AddMetric("gauge", "meMetric", "123")
	if err != nil {
		log.Panicf("err: %v\n", err)
	}

	err = client.Start()
	if err != nil {
		log.Panicf("err: %v\n", err)
	}
}
