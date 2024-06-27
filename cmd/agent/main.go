// Запускает Агент.
// Параметры Агента (задаются через [falg] и/или [env]):
// - адрес эндпоинта HTTP-сервера (default "localhost:8080") [-a] [ADDRESS]
// - частота опроса метрик из пакета runtime (default 2) [-p] [POLL_INTERVAL]
// - частота отправки метрик на сервер (default 10) [-r] [REPORT_INTERVAL]
// - количество одновременно исходящих запросов на сервер (default 10) [-l] [RATE_LIMIT]
// - ключ для подписи передаваемых данных [-k] [KEY]
// - путь до файла с публичным ключом для шифрования (default "/tmp/public.pem") [-crypto-key] [CRYPTO_KEY]
// - уровень логирования (default "err") [-lvl] [LVL]
package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/AndreyVLZ/metrics/agent"
	"github.com/AndreyVLZ/metrics/agent/config"
	"github.com/AndreyVLZ/metrics/pkg/flagenv"
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

	var (
		addr           string
		pollInterval   int
		reportInterval int
		key            string
		rateLimit      int
		publicKeyPath  string
		logLevel       string
	)

	if err := flagenv.New(
		flagenv.String(&addr, "ADDRESS", "a", agent.AddressDefault, "адрес эндпоинта HTTP-сервера"),
		flagenv.Int(&pollInterval, "POLL_INTERVAL", "p", agent.PollIntervalDefault, "частота опроса метрик из пакета runtime"),
		flagenv.Int(&reportInterval, "REPORT_INTERVAL", "r", agent.ReportIntervalDefault, "частота отправки метрик на сервер"),
		flagenv.String(&key, "KEY", "k", "", "ключ"),
		flagenv.Int(&rateLimit, "RATE_LIMIT", "l", agent.ReportIntervalDefault, "количество одновременно исходящих запросов на сервер"),
		flagenv.String(&publicKeyPath, "CRYPTO_KEY", "crypto-key", agent.CryproKeyPathDefault, "путь до файла с публичным ключом"),
		flagenv.String(&logLevel, "LVL", "lvl", mylog.LevelErr, "уровень логирования"),
	); err != nil {
		log.Printf("flagenv: %v\n", err)

		return
	}

	flag.Parse()
	ctx := context.Background()
	logger := mylog.New(mylog.SlogKey, logLevel)

	cfg, err := config.New(
		config.SetAddr(addr),
		config.SetPollInterval(pollInterval),
		config.SetReportInterval(reportInterval),
		config.SetRateLimit(rateLimit),
		config.SetKey(key),
		config.SetCryptoKeyPath(publicKeyPath),
	)

	if err != nil {
		logger.Error("config new", "error", err)

		return
	}

	agent := agent.New(logger, cfg)

	if err := shutdown.New(agent, timeout).Start(ctx); err != nil {
		logger.Error("shutdown", "error", err)
	}
}
