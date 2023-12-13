package main

import (
	"flag"
	"log"

	"github.com/AndreyVLZ/metrics/cmd/server/config"
	"github.com/AndreyVLZ/metrics/cmd/server/metricserver"
	"github.com/AndreyVLZ/metrics/internal/handlers"
	"github.com/AndreyVLZ/metrics/internal/route"
	"github.com/AndreyVLZ/metrics/internal/storage/memstorage"
)

func main() {
	conf := config.Config{}
	flag.StringVar(&conf.Addr, "a", "localhost:8080", "docs")
	flag.Parse()
	//err := StartWhitServeMux()
	err := StartWhitChiMux(conf)
	if err != nil {
		log.Println(err)
	}
}

func StartWhitServeMux(conf config.Config) error {
	// сборка хранилища
	gaugeRepo := memstorage.NewGaugeRepo()
	counterRepo := memstorage.NewCounterRepo()
	store := memstorage.New(gaugeRepo, counterRepo)
	// хендлеры
	hand := handlers.NewMetricHandler(store)

	// объявление роутера ServeMux
	router := route.NewServeMux()

	// установка хендлеров в роутер
	handler := router.SetHandlers(hand)

	// сервер
	srv := metricserver.New(handler, conf)

	return srv.Start()
}

func StartWhitChiMux(conf config.Config) error {
	// сборка хранилища
	gaugeRepo := memstorage.NewGaugeRepo()
	counterRepo := memstorage.NewCounterRepo()
	store := memstorage.New(gaugeRepo, counterRepo)
	// хендлеры
	hand := handlers.NewChiHandler(store)

	// объявление роутера ChiMux
	router := route.NewChiMux()

	// установка хендлеров в роутер
	handler := router.SetHandlers(hand)

	// сервер
	srv := metricserver.New(handler, conf)

	return srv.Start()
}
