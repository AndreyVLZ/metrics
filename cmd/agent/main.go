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
	"log/slog"
	"os"
	"time"

	"github.com/AndreyVLZ/metrics/agent"
	"github.com/AndreyVLZ/metrics/agent/config"
	mylog "github.com/AndreyVLZ/metrics/pkg/log"
	"github.com/AndreyVLZ/metrics/pkg/parser"
	"github.com/AndreyVLZ/metrics/pkg/parser/convert"
	"github.com/AndreyVLZ/metrics/pkg/parser/env"
	"github.com/AndreyVLZ/metrics/pkg/parser/field"
	"github.com/AndreyVLZ/metrics/pkg/parser/flag"
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
		addr           = config.AddressDefault
		rateLimit      = config.RateLimitDefault
		logLevel       = config.LogLevelDefault
		pollInterval   = config.PollIntervalDefault
		reportInterval = config.ReportIntervalDefault
		configPath     = ""
		key            = ""
		cryptoKeyPath  = ""
	)

	parser.File(&configPath,
		flag.String("c", "путь до файла конфигурации"),
		env.String("CONFIG"),
	)

	parser.Value(&pollInterval,
		field.Duration("poll_interval"),
		convert.IntToDuration(time.Second,
			flag.Int("p", "частота опроса метрик из пакета runtime"),
			env.Int("POLL_INTERVAL"),
		),
	)

	parser.Value(&reportInterval,
		field.Duration("report_interval"),
		convert.IntToDuration(time.Second,
			flag.Int("r", "частота отправки метрик на сервер"),
			env.Int("POLL_INTERVAL"),
		),
	)

	parser.Value(&rateLimit,
		flag.Int("l", "количество одновременно исходящих запросов на сервер"),
		env.Int("RATE_LIMIT"),
	)

	parser.Value(&addr,
		field.String("address"),
		flag.String("a", "адрес эндпоинта HTTP-сервера"),
		env.String("ADDRESS"),
	)

	parser.Value(&key,
		flag.String("k", "ключ"),
		env.String("KEY"),
	)

	parser.Value(&cryptoKeyPath,
		field.String("database_dsn"),
		flag.String("crypto-key", "путь до файла с приватным ключом"),
		env.String("CRYPTO_KEY"),
	)

	parser.Value(&logLevel,
		flag.String("lvl", "уровень логирования"),
		env.String("LVL"),
	)

	if err := parser.Parse(os.Args[1:]); err != nil {
		log.Printf("err:%v\n", err)

		return
	}

	cfg, err := config.New(
		config.SetRateLimit(rateLimit),
		config.SetAddr(addr),
		config.SetPollInterval(pollInterval),
		config.SetReportInterval(reportInterval),
		config.SetKey(key),
		config.SetConfigPath(configPath),
		config.SetCryptoKeyPath(cryptoKeyPath),
		config.SetLogLevel(logLevel),
	)
	if err != nil {
		log.Printf("new config: %v\n", err)

		return
	}

	logger := mylog.New(mylog.SlogKey, logLevel)

	if err := runAgent(cfg, logger); err != nil {
		logger.Error("shutdown", "error", err)
	}
}

func runAgent(cfg *config.Config, mlog *slog.Logger) error {
	ctx := context.Background()

	agent := agent.New(cfg, mlog)

	return shutdown.New(agent, timeout).Start(ctx)
}
