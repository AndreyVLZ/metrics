package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/AndreyVLZ/metrics/pkg/env"
	mylog "github.com/AndreyVLZ/metrics/pkg/log"
	"github.com/AndreyVLZ/metrics/pkg/shutdown"
	"github.com/AndreyVLZ/metrics/server"
	"github.com/AndreyVLZ/metrics/server/adapter"
)

const timeout time.Duration = 7

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	log.Printf("\nBuild version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)

	var (
		addr          = server.AddressDefault
		storeInterval = server.StoreIntervalDefault
		storagePath   = server.StoragePathDefault
		databaseDNS   = ""
		isRestore     = server.IsRestoreDefault
		key           = ""
		logLevel      = mylog.LevelInfo
	)

	flag.StringVar(&addr, "a", addr, "адрес эндпоинта HTTP-сервера")
	flag.IntVar(&storeInterval, "i", storeInterval, "интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск")
	flag.StringVar(&storagePath, "f", storagePath, "полное имя файла, куда сохраняются текущие значения")
	flag.BoolVar(&isRestore, "r", isRestore, "определяющее, загружать или нет ранее сохранённые значения из указанного файла при старте сервера")
	flag.StringVar(&databaseDNS, "d", databaseDNS, "строка с адресом подключения к БД")
	flag.StringVar(&key, "k", key, "ключ")
	flag.StringVar(&logLevel, "lvl", logLevel, "уровень логирования")

	flag.Parse()

	ctx := context.Background()
	logger := mylog.New(mylog.SlogKey, logLevel)

	vars := env.Array(
		env.String(&addr, "ADDRESS"),
		env.Int(&storeInterval, "STORE_INTERVAL"),
		env.String(&storagePath, "FILE_STORAGE_PATH"),
		env.Bool(&isRestore, "RESTORE"),
		env.String(&databaseDNS, "DATABASE_DSN"),
		env.String(&key, "KEY"),
	)

	for i := range vars {
		err := vars[i]()
		if err != nil {
			logger.DebugContext(ctx, err.Error())
		}
	}

	opts := []server.FuncOpt{
		server.SetAddr(addr),
		server.SetStoreInt(storeInterval),
		server.SetStorePath(storagePath),
		server.SetRestore(isRestore),
		server.SetDatabaseDNS(databaseDNS),
		server.SetKey(key),
	}

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	if err := runServer(ctx, timeout, logger, opts...); err != nil {
		logger.ErrorContext(ctx, "start server", "error", err)
	}
}

func runServer(ctx context.Context, timeout time.Duration, log *slog.Logger, opts ...server.FuncOpt) error {
	server := server.New(log, opts...)
	shutdown := shutdown.New(
		adapter.NewShutdown(&server),
		timeout,
	)

	return shutdown.Start(ctx)
}
