package metricserver

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/AndreyVLZ/metrics/cmd/server/producer"
	"github.com/AndreyVLZ/metrics/cmd/server/route"
	"github.com/AndreyVLZ/metrics/cmd/server/service/restoreservice"
	"github.com/AndreyVLZ/metrics/cmd/server/service/saveservice"
	"github.com/AndreyVLZ/metrics/cmd/server/wrapstore"
	"github.com/AndreyVLZ/metrics/internal/storage"
)

type FuncOpt func(*metricServer)

type Service interface {
	Name() string
	Start() error
	Stop() error
}

type ServerConfig struct {
	storageType storage.StorageType
	dbDNS       string
	storePath   string
	isRestore   bool
	storeInt    int //при старте
	key         string
}

type metricServer struct {
	wg       sync.WaitGroup
	log      *slog.Logger
	server   *http.Server
	store    storage.Storage
	cfg      *ServerConfig
	services []Service
}

func New(log *slog.Logger, opts ...FuncOpt) (*metricServer, error) {
	srv := &metricServer{
		log:    log,
		server: &http.Server{},
		cfg:    &ServerConfig{},
	}

	for _, opt := range opts {
		opt(srv)
	}

	err := srv.initStore()
	if err != nil {
		return nil, err
	}

	// добавляет сервисы и меняет store
	srv.configureServices()

	router, err := route.New(route.Config{
		RouteType: route.RouteTypeServeMux,
		//RouteType: route.RouteTypeChi,
		Store:     srv.store,
		Log:       srv.log,
		SecretKey: srv.cfg.key,
	})
	if err != nil {
		return nil, err
	}

	srv.server.Handler = router

	return srv, nil
}

func (s *metricServer) Start() {
	mainCtx := context.Background()
	ctx, stop := signal.NotifyContext(mainCtx, os.Interrupt)
	defer stop()

	// запускаем хранилище
	err := s.store.Open()
	if err != nil {
		s.log.Error("store Open", "err", err)
		//return
	}

	// запускаем сервисы
	for i := range s.services {
		err := s.services[i].Start()
		s.log.Info("services Start", "name", s.services[i].Name())
		if err != nil {
			s.log.Error("ERR Start services", s.services[i].Name(), err)
		}
	}

	s.log.LogAttrs(mainCtx,
		slog.LevelInfo, "start server",
		slog.String("addr", s.server.Addr),
		slog.Group("flags",
			slog.Int("storeInterval", s.cfg.storeInt),
			slog.String("storePath", s.cfg.storePath),
			slog.Bool("restore", s.cfg.isRestore),
			slog.String("dbDNS", s.cfg.dbDNS),
			slog.String("key", s.cfg.key),
		),
	)

	servicesStopedCtx, cancelStopped := context.WithCancel(mainCtx)
	defer cancelStopped()

	// регистрируем функции для остановки
	s.registerOnShutdown()

	// слушаем сервер
	go s.listenAndServe()

	// ловим сигналы выхода
	<-ctx.Done()
	s.log.Info("Ctrl+C")
	stop()

	// Ждем когда остановятся все сервисы
	go func() {
		s.wg.Wait()
		s.log.Info("all services stopped")
		// отменяем контекс для сервисов
		cancelStopped()
	}()

	// контекс для отмены на 10сек
	timeoutCtx, cancel := context.WithTimeout(mainCtx, 10*time.Second)
	defer cancel()

	if err := s.server.Shutdown(timeoutCtx); err != nil {
		s.log.Error("err shutdown", err)
	}

	select {
	case <-timeoutCtx.Done():
		if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
			s.log.Error("deadline")
		}
	case <-servicesStopedCtx.Done():
		s.log.Info("server stop")
	}
}

func (s *metricServer) registerOnShutdown() {
	for i := range s.services {
		s.server.RegisterOnShutdown(
			s.registerByWG(&s.wg, s.services[i]),
		)
	}
}

func (s *metricServer) registerByWG(wg *sync.WaitGroup, service Service) func() {
	wg.Add(1)
	s.log.Info("registry service", "name", service.Name())
	return func() {
		defer wg.Done()
		err := service.Stop()
		if err != nil {
			s.log.Error("stop service", "name", service.Name(), "err", err)
		}
		s.log.Info("stop sevice", "name", service.Name())
	}
}

func (s *metricServer) initStore() error {
	storageType := storage.StorageTypeInmemory

	if s.cfg.dbDNS != "" {
		storageType = storage.StorageTypePostgres
	}

	store, err := storage.New(storage.Config{
		StorageType: storageType,
		ConnDB:      s.cfg.dbDNS,
	})
	if err != nil {
		return err
	}

	s.store = store

	return nil
}

func (s *metricServer) addService(servise Service) {
	s.services = append(s.services, servise)
}

func (s *metricServer) configureServices() {
	if s.cfg.storePath == "" {
		return
	}

	// Запись в файл
	if s.cfg.isRestore {
		// restoreService опционально
		s.addService(
			restoreservice.New(s.store, s.cfg.storePath),
		)
	}

	prod := producer.New(s.cfg.storePath)
	switch s.cfg.storeInt {
	case 0:
		// wrapStore для записи в файл синхронно
		s.store = wrapstore.NewWrapStore(s.store, prod)
	default:
		// saveService дфл записи в файл по интервалу
		s.addService(saveservice.New(
			s.store,
			s.cfg.storeInt,
			prod,
		))
	}
}

func (s *metricServer) listenAndServe() {
	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("err ListenAndServe: %v", err)
	}
}

func SetAddr(addr string) FuncOpt {
	return func(s *metricServer) {
		s.server.Addr = addr
	}
}

func SetStoreInt(interval int) FuncOpt {
	return func(s *metricServer) {
		s.cfg.storeInt = interval
	}
}

func SetStorePath(path string) FuncOpt {
	return func(s *metricServer) {
		s.cfg.storePath = path
	}
}

func SetRestore(b bool) FuncOpt {
	return func(s *metricServer) {
		s.cfg.isRestore = b
	}
}

func SetDatabaseDNS(dns string) FuncOpt {
	return func(s *metricServer) {
		s.cfg.dbDNS = dns
	}
}

func SetKey(key string) FuncOpt {
	return func(s *metricServer) {
		s.cfg.key = key
	}
}
