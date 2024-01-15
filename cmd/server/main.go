package main

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/AndreyVLZ/metrics/cmd/server/metricserver"
	"github.com/AndreyVLZ/metrics/cmd/server/route/servemux"

	//sLog "github.com/AndreyVLZ/metrics/internal/log/slog"
	"github.com/AndreyVLZ/metrics/internal/log/zap"
	"github.com/AndreyVLZ/metrics/internal/storage/memstorage"
)

func main() {
	addrPtr := flag.String("a", "localhost:8080", "адрес эндпоинта HTTP-сервера")
	storeIntervalPtr := flag.Int("i", 300, "интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск")
	storagePathPtr := flag.String("f", "/tmp/metrics-db.json", "полное имя файла, куда сохраняются текущие значения")
	restorePrt := flag.Bool("r", true, "определяющее, загружать или нет ранее сохранённые значения из указанного файла при старте сервера ")
	flag.Parse()

	var opts []metricserver.FuncOpt
	opts = append(opts,
		metricserver.SetAddr(*addrPtr),
		metricserver.SetStoreInt(*storeIntervalPtr),
		metricserver.SetStorePath(*storagePathPtr),
		metricserver.SetRestore(*restorePrt),
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

	// хранилище
	store := memstorage.New()

	// объявление роутера
	route := servemux.New()
	//router := route.NewServeMux(wrapStore)

	logger := zap.New(zap.DefaultConfig())

	//сервер
	srv, err := metricserver.New(logger, route, store, opts...)
	if err != nil {
		log.Printf("!Err %n", err)
	}

	srv.Start()
}
