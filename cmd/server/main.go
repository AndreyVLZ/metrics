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
	"log/slog"
	_ "net/http/pprof"
	"os"
	"syscall"
	"time"

	mylog "github.com/AndreyVLZ/metrics/pkg/log"
	"github.com/AndreyVLZ/metrics/pkg/parser"
	"github.com/AndreyVLZ/metrics/pkg/parser/convert"
	"github.com/AndreyVLZ/metrics/pkg/parser/env"
	"github.com/AndreyVLZ/metrics/pkg/parser/field"
	"github.com/AndreyVLZ/metrics/pkg/parser/flag"
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

	var (
		addr          = config.AddressDefault
		storeInterval = config.StoreIntervalDefault
		storePath     = config.StorePathDefault
		isRestore     = config.IsRestoreDefault
		logLevel      = mylog.LevelErr
		cryptoKeyPath = ""
		connDB        = ""
		configPath    = ""
		key           = ""
	)

	parser.File(&configPath,
		flag.String("c", "путь до файла конфигурации"),
		env.String("CONFIG"),
	)

	parser.Value(&storeInterval,
		field.Duration("store_interval"),
		convert.IntToDuration(time.Second,
			flag.Int("i", "интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск"),
			env.Int("STORE_INTERVAL"),
		),
	)

	parser.Value(&addr,
		field.String("address"),
		flag.String("a", "адрес эндпоинта HTTP-сервера"),
		env.String("ADDRESS"),
	)

	parser.Value(&storePath,
		field.String("store_file"),
		flag.String("f", "полное имя файла, куда сохраняются текущие значения"),
		env.String("FILE_STORAGE_PATH"),
	)

	parser.Value(&connDB,
		field.String("database_dsn"),
		flag.String("d", "строка с адресом подключения к БД"),
		env.String("DATABASE_DSN"),
	)

	parser.Value(&isRestore,
		field.Bool("restore"),
		flag.Bool("r", "определяющее, загружать или нет ранее сохранённые значения из указанного файла при старте сервера"),
		env.Bool("RESTORE"),
	)

	parser.Value(&cryptoKeyPath,
		field.String("database_dsn"),
		flag.String("crypto-key", "путь до файла с приватным ключом"),
		env.String("CRYPTO_KEY"),
	)

	parser.Value(&key,
		flag.String("k", "ключ"),
		env.String("KEY"),
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
		config.SetAddr(addr),
		config.SetStoreInt(storeInterval),
		config.SetStorePath(storePath),
		config.SetRestore(isRestore),
		config.SetCryptoKeyPath(cryptoKeyPath),
		config.SetDatabaseDNS(connDB),
		config.SetConfigPath(configPath),
		config.SetKey(key),
		config.SetLogLevel(logLevel),
	)

	if err != nil {
		log.Printf("new config: %v\n", err)

		return
	}

	logger := mylog.New(mylog.SlogKey, cfg.LogLevel)

	if err := runServer(cfg, logger); err != nil {
		logger.Error("start server", "error", err)
	}
}

func runServer(cfg *config.Config, mlog *slog.Logger) error {
	ctx := context.Background()

	server := server.New(cfg, mlog)

	shutdown := shutdown.New(
		adapter.NewShutdown(&server),
		timeout,
	)

	signals := []os.Signal{
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	}

	return shutdown.Start(ctx, signals...)
}

/*
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
*/
