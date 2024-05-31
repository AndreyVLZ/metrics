package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/AndreyVLZ/metrics/cmd/server/metricserver"

	"github.com/AndreyVLZ/metrics/internal/log/zap"
)

func main() {
	//
	addrPtr := flag.String("a", "localhost:8080", "адрес эндпоинта HTTP-сервера")
	storeIntervalPtr := flag.Int("i", 300, "интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск")
	storagePathPtr := flag.String("f", "/tmp/metrics-db.json", "полное имя файла, куда сохраняются текущие значения")
	restorePrt := flag.Bool("r", true, "определяющее, загружать или нет ранее сохранённые значения из указанного файла при старте сервера ")
	databaseDNSPtr := flag.String("d", "", "строка с адресом подключения к БД")
	keyPtr := flag.String("k", "", "ключ")
	flag.Parse()

	var opts []metricserver.FuncOpt
	opts = append(opts,
		metricserver.SetAddr(*addrPtr),
		metricserver.SetStoreInt(*storeIntervalPtr),
		metricserver.SetRestore(*restorePrt),
		metricserver.SetStorePath(*storagePathPtr),
		metricserver.SetDatabaseDNS(*databaseDNSPtr),
		metricserver.SetKey(*keyPtr),
	)

	if addrENV, ok := os.LookupEnv("ADDRESS"); ok {
		opts = append(opts, metricserver.SetAddr(addrENV))
	}

	if storeIntENV, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		if ri, err := strconv.Atoi(storeIntENV); err == nil {
			opts = append(opts, metricserver.SetStoreInt(ri))
		}
	}

	if storagePathENV, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		opts = append(opts, metricserver.SetStorePath(storagePathENV))
	}

	if restoreENV, ok := os.LookupEnv("RESTORE"); ok {
		opts = append(opts, metricserver.SetRestore(restoreENV == "true"))
	}

	if databaseENV, ok := os.LookupEnv("DATABASE_DSN"); ok {
		opts = append(opts, metricserver.SetDatabaseDNS(databaseENV))
	}

	if keyENV, ok := os.LookupEnv("KEY"); ok {
		opts = append(opts, metricserver.SetKey(keyENV))
	}

	// объявление логера
	logger := zap.New(zap.DefaultConfig())

	//сервер
	srv, err := metricserver.New(logger, opts...)
	if err != nil {
		logger.Error("Err server build %v\n", err)
	}

	srv.Start()
}
