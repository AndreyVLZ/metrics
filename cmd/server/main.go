// Запускает Сервер.
// Параметры Сервера (задаются через [falg] и/или [env]):
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/AndreyVLZ/metrics/pkg/flagenv"
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
		addr           string
		storeInterval  int
		storagePath    string
		databaseDNS    string
		isRestore      bool
		key            string
		logLevel       string
		privateKeyPath string
		configPath     string
	)

	if err := flagenv.New(
		flagenv.String(&addr, "ADDRESS", "a", server.AddressDefault, "адрес эндпоинта HTTP-сервера"),
		flagenv.Int(&storeInterval, "STORE_INTERVAL", "i", server.StoreIntervalDefault, "интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск"),
		flagenv.String(&storagePath, "FILE_STORAGE_PATH", "f", server.StoragePathDefault, "полное имя файла, куда сохраняются текущие значения"),
		flagenv.String(&databaseDNS, "DATABASE_DSN", "d", "", "строка с адресом подключения к БД"),
		flagenv.String(&key, "KEY", "k", "", "ключ"),
		flagenv.Bool(&isRestore, "RESTORE", "r", server.IsRestoreDefault, "определяющее, загружать или нет ранее сохранённые значения из указанного файла при старте сервера"),
		flagenv.String(&privateKeyPath, "CRYPTO_KEY", "crypto-key", server.CryptoKeyPathDefault, "путь до файла с приватным ключом"),
		flagenv.String(&configPath, "CONFIG", "c", "", "путь до файла конфигурации"),
		flagenv.String(&logLevel, "LVL", "lvl", mylog.LevelErr, "уровень логирования"),
	); err != nil {
		log.Printf("flagenv: %v\n", err)

		return
	}

	flag.Parse()

	opts := []server.FuncOpt{
		server.SetAddr(addr),
		server.SetStoreInt(storeInterval),
		server.SetStorePath(storagePath),
		server.SetRestore(isRestore),
		server.SetDatabaseDNS(databaseDNS),
		server.SetKey(key),
		server.SetCryptoKeyPath(privateKeyPath),
	}

	ctx := context.Background()
	logger := mylog.New(mylog.SlogKey, logLevel)

	cfg, err := server.NewConfig(configPath, opts...)
	if err != nil {
		logger.Error("new config", "error", err)

		return
	}

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	server := server.New(cfg, logger)
	shutdown := shutdown.New(
		adapter.NewShutdown(&server),
		timeout,
	)

	if err := shutdown.Start(ctx); err != nil {
		logger.Error("start server", "error", err)
	}
}

/*
func runServer(logger *slog.Logger, opts ...server.FuncOpt) error {
	ctx := context.Background()

	cfg := server.NewConfig(opts...)
	server := server.New(cfg, logger)
	shutdown := shutdown.New(
		adapter.NewShutdown(&server),
		timeout,
	)

	return shutdown.Start(ctx)
}
*/
