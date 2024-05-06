package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"

	arg "github.com/AndreyVLZ/metrics/internal/argument"
	"github.com/AndreyVLZ/metrics/internal/log/zap"
	"github.com/AndreyVLZ/metrics/server"
)

const (
	addr    = ":8080"  // адрес сервера
	maxSize = 10000000 // будем растить слайс до 10 миллионов элементов
)

func main() {
	go func() {
		http.ListenAndServe("localhost:8081", nil)
	}()

	var (
		addr          = server.AddressDefault
		storeInterval = server.StoreIntervalDefault
		storagePath   = server.StoragePathDefault
		databaseDNS   = ""
		isRestore     = server.IsRestoreDefault
		key           = ""
	)

	args := arg.Array(
		arg.String(
			&addr,     // val
			"a",       // name Flag
			"ADDRESS", // name ENV
			"адрес эндпоинта HTTP-сервера", // desc
		),
		arg.Int(
			&storeInterval,
			"i",
			"STORE_INTERVAL",
			"интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск",
		),
		arg.String(
			&storagePath,
			"f",
			"FILE_STORAGE_PATH",
			"полное имя файла, куда сохраняются текущие значения",
		),
		arg.Bool(
			&isRestore,
			"r",
			"RESTORE",
			"определяющее, загружать или нет ранее сохранённые значения из указанного файла при старте сервера",
		),
		arg.String(
			&databaseDNS,
			"d",
			"DATABASE_DSN",
			"строка с адресом подключения к БД",
		),
		arg.String(
			&key,
			"k",
			"KEY",
			"ключ",
		),
	)

	for i := range args {
		err := args[i]()
		if err != nil {
			log.Printf("err parse args: %v\n", err)
		}
	}

	flag.Parse()

	cfg := server.NewConfig(
		server.SetAddr(addr),
		server.SetStoreInt(storeInterval),
		server.SetStorePath(storagePath),
		server.SetRestore(isRestore),
		server.SetDatabaseDNS(databaseDNS),
		server.SetKey(key),
	)

	run(cfg)
}

func run(cfg *server.Config) {
	ctx := context.Background()

	// объявление логера
	logger := zap.New(zap.DefaultConfig())

	server := server.New(cfg, logger)

	chErr := make(chan error)
	ctxSinal, stopSignal := signal.NotifyContext(ctx, os.Interrupt)

	go func(ce chan<- error) {
		defer close(ce)
		ce <- server.Start(ctxSinal)
	}(chErr)

	select {
	case <-ctxSinal.Done():
		log.Println("signal")
	case err := <-chErr:
		stopSignal()

		if err != nil {
			log.Printf("server start err %v\n", err)
		}
	}

	if err := server.Stop(ctx); err != nil {
		log.Printf("server stop err: %v\n", err)
	} else {
		log.Println("all services stopped")
	}

	log.Println("server stopped")
}
