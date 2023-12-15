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
	opts = append(opts, metricserver.SerAddr(*addrPtr))

	if addrENV, ok := os.LookupEnv("ADDRESS"); ok {
		opts = append(opts, metricserver.SerAddr(addrENV))
	}

	//err := StartWhitServeMux(opts...)
	err := StartWhitChiMux(opts...)
	if err != nil {
		log.Println(err)
	}
}

func StartWhitServeMux(opts ...metricserver.FuncOpt) error {
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
	srv := metricserver.New(handler, opts...)

	return srv.Start()
}

func StartWhitChiMux(opts ...metricserver.FuncOpt) error {
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
	srv := metricserver.New(handler, opts...)

	return srv.Start()
}
