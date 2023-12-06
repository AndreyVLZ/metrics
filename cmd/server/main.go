package main

import (
	"log"

	"github.com/AndreyVLZ/metrics/internal/metricserver"
	"github.com/AndreyVLZ/metrics/internal/storage/memstorage"
)

func main() {
	store := memstorage.New()
	server := metricserver.New(store)

	err := server.Start()
	if err != nil {
		log.Println(err)
	}
}
