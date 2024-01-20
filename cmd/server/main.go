package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/AndreyVLZ/metrics/cmd/server/metricserver"
	"github.com/AndreyVLZ/metrics/cmd/server/route/servemux"

	"github.com/AndreyVLZ/metrics/internal/log/zap"
	"github.com/AndreyVLZ/metrics/internal/storage/memstorage"
)

func main() {
	addrPtr := flag.String("a", "localhost:8080", "адрес эндпоинта HTTP-сервера")
	storeIntervalPtr := flag.Int("i", 300, "интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск")
	storagePathPtr := flag.String("f", "/tmp/metrics-db.json", "полное имя файла, куда сохраняются текущие значения")
	restorePrt := flag.Bool("r", true, "определяющее, загружать или нет ранее сохранённые значения из указанного файла при старте сервера ")
	databaseDNSPtr := flag.String("d", "", "строка с адресом подключения к БД")
	flag.Parse()

	var opts []metricserver.FuncOpt
	opts = append(opts,
		metricserver.SetAddr(*addrPtr),
		metricserver.SetStoreInt(*storeIntervalPtr),
		metricserver.SetStorePath(*storagePathPtr),
		metricserver.SetRestore(*restorePrt),
		metricserver.SetDatabaseDNS(*databaseDNSPtr),
	)

	if addrENV, ok := os.LookupEnv("ADDRESS"); ok {
		opts = append(opts, metricserver.SetAddr(addrENV))
	}

	var ri int
	if storeIntENV, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		var err error
		if ri, err = strconv.Atoi(storeIntENV); err == nil {
			opts = append(opts, metricserver.SetStoreInt(ri))
		}
	}
	// NOTE проверить на пустое значение
	if storagePathENV, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		opts = append(opts, metricserver.SetStorePath(storagePathENV))
	}

	if restoreENV, ok := os.LookupEnv("RESTORE"); ok {
		var r bool
		if restoreENV == "true" {
			r = true
		}
		opts = append(opts, metricserver.SetRestore(r))
	}

	if databaseENV, ok := os.LookupEnv("DATABASE_DSN"); ok {
		opts = append(opts, metricserver.SetDatabaseDNS(databaseENV))
	}

	// хранилище
	store := memstorage.New()

	// объявление роутера
	route := servemux.New()

	// объявление логера
	logger := zap.New(zap.DefaultConfig())

	//сервер
	srv, err := metricserver.New(logger, route, store, opts...)
	if err != nil {
		logger.Error("Err server build %v\n", err)
	}

	srv.Start()
}
