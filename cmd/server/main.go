package main

import (
	"context"
	"flag"
	"log"
	"time"

	arg "github.com/AndreyVLZ/metrics/internal/argument"
	"github.com/AndreyVLZ/metrics/internal/log/zap"
	"github.com/AndreyVLZ/metrics/internal/shutdown"
	"github.com/AndreyVLZ/metrics/server"
	"github.com/AndreyVLZ/metrics/server/adapter"
)

func main() {
	var (
		addr          = server.AddressDefault
		storeInterval = server.StoreIntervalDefault
		storagePath   = server.StoragePathDefault
		databaseDNS   = ""
		isRestore     = server.IsRestoreDefault
		key           = ""
	)

	args := arg.Array(
		arg.String(&addr, "ADDRESS"),
		arg.Int(&storeInterval, "STORE_INTERVAL"),
		arg.String(&storagePath, "FILE_STORAGE_PATH"),
		arg.Bool(&isRestore, "RESTORE"),
		arg.String(&databaseDNS, "DATABASE_DSN"),
		arg.String(&key, "KEY"),
	)

	flag.StringVar(&addr, "a", addr, "адрес эндпоинта HTTP-сервера")
	flag.IntVar(&storeInterval, "i", storeInterval, "интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск")
	flag.StringVar(&storagePath, "f", storagePath, "полное имя файла, куда сохраняются текущие значения")
	flag.BoolVar(&isRestore, "r", isRestore, "определяющее, загружать или нет ранее сохранённые значения из указанного файла при старте сервера")
	flag.StringVar(&databaseDNS, "d", databaseDNS, "строка с адресом подключения к БД")
	flag.StringVar(&key, "k", key, "ключ")

	flag.Parse()

	for i := range args {
		err := args[i]()
		if err != nil {
			log.Printf("err parse args: %v\n", err)
		}
	}

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
	var timeout time.Duration = 7

	logger := zap.New(zap.DefaultConfig())
	server := server.New(cfg, logger)
	shutdown := shutdown.New(
		adapter.NewShutdown(&server),
		timeout,
	)

	if err := shutdown.Start(context.Background()); err != nil {
		log.Printf("server error: %v\n", err)
	}
}
