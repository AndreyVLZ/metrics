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
