// Запускает Сервер.
// Параметры Сервера (задаются через [configPath] и/или [falg] и/или [env]):
//   - адрес эндпоинта HTTP-сервера
//     ["localhost:8080"] [-a] [ADDRESS]
//   - путь до файла конфигурации
//     [""] [-c] [CONFIG]
//   - путь до файла с приватным ключом
//     ["/tmp/private.pem"] [-crypto-key] [CRYPTO_KEY]
//   - строка с адресом подключения к БД
//     [""] [-d] [DATABASE_DSN]
//   - полное имя файла, куда сохраняются текущие значения
//     [""] [-f] [FILE_STORAGE_PATH]
//   - интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск
//     [300] [-i] [STORE_INTERVAL]
//   - ключ
//     [""] [-k] [KEY]
//   - уровень логирования
//     ["err"] [-lvl] [LVL]
//   - определяющее, загружать или нет ранее сохранённые значения из указанного файла при старте сервера
//     [true] [-r] [RESTORE]
package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"syscall"
	"time"

	mylog "github.com/AndreyVLZ/metrics/pkg/log"
	"github.com/AndreyVLZ/metrics/pkg/shutdown"
	"github.com/AndreyVLZ/metrics/server"
	"github.com/AndreyVLZ/metrics/server/adapter"
	"github.com/AndreyVLZ/metrics/server/config"
)

const timeout time.Duration = 7

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	log.Printf("\nBuild version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)

	cfg, err := config.New(os.Args[1:])
	if err != nil {
		log.Printf("new config: %v\n", err)

		return
	}

	ctx := context.Background()
	logger := mylog.New(mylog.SlogKey, cfg.LogLevel)

	go func() {
		logger.ErrorContext(ctx, "pprof", "err", http.ListenAndServe("localhost:6060", nil))
	}()

	server := server.New(cfg, logger)
	shutdown := shutdown.New(adapter.NewShutdown(&server), timeout)
	signals := []os.Signal{syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT}

	if err := shutdown.Start(ctx, signals...); err != nil {
		logger.ErrorContext(ctx, "start server", "error", err)
	}
}
