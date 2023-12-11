package main

import (
	"log"

	"github.com/AndreyVLZ/metrics/cmd/server/metricserver"
	"github.com/AndreyVLZ/metrics/internal/storage/memstorage"
)

func main() {
	gaugeRepo := memstorage.NewGaugeRepo()
	counterRepo := memstorage.NewCounterRepo()
	store := memstorage.New(gaugeRepo, counterRepo)
	server := metricserver.New(store)
	err := server.Start()
	if err != nil {
		log.Println(err)
	}
}
