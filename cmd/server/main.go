package main

import (
	"flag"
	"log"
	"os"

	"github.com/AndreyVLZ/metrics/cmd/server/metricserver"
	"github.com/AndreyVLZ/metrics/cmd/server/route"

	//sLog "github.com/AndreyVLZ/metrics/internal/log/slog"
	"github.com/AndreyVLZ/metrics/internal/log/zap"
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
	store := memstorage.New()

	// хендлеры
	//hand := handlers.NewChiHandle(store)
	//hand := handlers.NewMainHandle(store)

	// объявление роутера ChiMux
	//router := route.NewChiMux()
	router := route.NewServeMux(store)

	// установка хендлеров в роутер

	//logger slog
	//logger := sLog.New()
	//logger zap
	logger := zap.New(zap.DefaultConfig())
	// сервер
	srv := metricserver.New(logger, router, opts...)

	err := srv.Start()
	if err != nil {
		log.Println(err)
	}
}
