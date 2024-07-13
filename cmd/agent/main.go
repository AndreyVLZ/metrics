// Запускает Агент.
// Параметры Агента (задаются через [configPath] и/или [falg] и/или [env]):
//   - адрес эндпоинта HTTP-сервера
//     ["localhost:8080"] [-a] [ADDRESS]
//   - путь до файла конфигурации
//     [""] [-c] [CONFIG]
//   - путь до файла с публичным ключом для шифрования
//     ["/tmp/public.pem"] [-crypto-key] [CRYPTO_KEY]
//   - ключ для подписи передаваемых данных
//     [""][-k] [KEY]
//   - количество одновременно исходящих запросов на сервер
//     [10] [-l] [RATE_LIMIT]
//   - частота опроса метрик из пакета runtime
//     [2] [-p] [POLL_INTERVAL]
//   - уровень логирования
//     ["err"] [-lvl] [LVL]
//   - частота отправки метрик на сервер
//     [10] [-r] [REPORT_INTERVAL]
package main

import (
	"context"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/AndreyVLZ/metrics/agent"
	"github.com/AndreyVLZ/metrics/agent/config"
	mylog "github.com/AndreyVLZ/metrics/pkg/log"
	"github.com/AndreyVLZ/metrics/pkg/shutdown"
)

const timeout time.Duration = 7

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	log.Printf("\nBuild version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)

	ctx := context.Background()

	cfg, err := config.New(os.Args[1:])
	if err != nil {
		log.Printf("new config: %v\n", err)
	}

	logger := mylog.New(mylog.SlogKey, cfg.LogLevel)

	agent := agent.New(cfg, logger)

	signals := []os.Signal{
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	}

	if err := shutdown.New(agent, timeout).Start(ctx, signals...); err != nil {
		logger.Error("run agent", "error", err)
	}
}
