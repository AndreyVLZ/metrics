package main

import (
	"flag"

	"github.com/AndreyVLZ/metrics/cmd/server/metricserver"

	arg "github.com/AndreyVLZ/metrics/internal/argument"
	"github.com/AndreyVLZ/metrics/internal/log/zap"
)

func main() {
	var (
		addr          = metricserver.AddressDefault
		storeInterval = metricserver.StoreIntervalDefault
		storagePath   = metricserver.StoragePathDefault
		databaseDNS   = ""
		isRestore     = metricserver.IsRestore
		key           = ""
	)

	args := arg.Array(
		arg.String(
			&addr,
			"a",
			"ADDRESS",
			"адрес эндпоинта HTTP-сервера",
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

	// объявление логера
	logger := zap.New(zap.DefaultConfig())

	for i := range args {
		err := args[i]()
		if err != nil {
			logger.Info("parse args", "err", err)
		}
	}

	flag.Parse()

	//сервер
	srv, err := metricserver.New(
		logger,
		metricserver.SetAddr(addr),
		metricserver.SetStoreInt(storeInterval),
		metricserver.SetStorePath(storagePath),
		metricserver.SetRestore(isRestore),
		metricserver.SetDatabaseDNS(databaseDNS),
		metricserver.SetKey(key),
	)
	if err != nil {
		logger.Error("Err server build %v\n", err)
	}

	srv.Start()
}
