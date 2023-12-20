package main

import (
	"flag"
	"log"
	"os"

	"github.com/AndreyVLZ/metrics/cmd/server/handlers"
	"github.com/AndreyVLZ/metrics/cmd/server/metricserver"
	"github.com/AndreyVLZ/metrics/cmd/server/route"
	"github.com/AndreyVLZ/metrics/internal/storage/memstorage"
)

func main() {
	addrPtr := flag.String("a", "localhost:8080", "адрес эндпоинта HTTP-сервера")
	flag.Parse()

	var opts []metricserver.FuncOpt
	opts = append(opts, metricserver.SetAddr(*addrPtr))

	if addrENV, ok := os.LookupEnv("ADDRESS"); ok {
		opts = append(opts, metricserver.SetAddr(addrENV))
	}

	// сборка хранилища
	gaugeRepo := memstorage.NewGaugeRepo()
	counterRepo := memstorage.NewCounterRepo()
	store := memstorage.New(gaugeRepo, counterRepo)

	// хендлеры
	hand := handlers.NewChiHandle(store)

	// объявление роутера ChiMux
	router := route.NewChiMux()

	// установка хендлеров в роутер
	handler := router.SetHandlers(hand)

	// сервер
	srv := metricserver.New(handler, opts...)

	err := srv.Start()
	if err != nil {
		log.Println(err)
	}
}
