package metricserver

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/AndreyVLZ/metrics/cmd/server/consumer"
	"github.com/AndreyVLZ/metrics/cmd/server/producer"
	"github.com/AndreyVLZ/metrics/cmd/server/route/middleware"
	"github.com/AndreyVLZ/metrics/cmd/server/wrapstore"
	"github.com/AndreyVLZ/metrics/internal/storage"
)

type Router interface {
	SetStore(storage.Storage) http.Handler
}

type FuncOpt func(*metricServer)

type metricServer struct {
	server    http.Server
	store     storage.Storage
	log       *slog.Logger
	handler   http.Handler
	storeInt  int //при старте
	storePath string
	isRestore bool
	consumer  *consumer.Consumer // для чтения метрик
	producer  *producer.Producer // для записи метрик
}

func (s *metricServer) configure(router Router, store storage.Storage, storePath string) error {
	consumer, err := consumer.NewConsumer(storePath)
	if err != nil {
		return err
	}
	s.consumer = consumer

	producer, err := producer.NewProducer(storePath)
	if err != nil {
		return err
	}
	s.producer = producer

	if s.storeInt == 0 {
		store = wrapstore.NewWrapStore(store, producer)
	}

	s.server.Handler = middleware.Logging(s.log, router.SetStore(store))

	return nil
}

func New(log *slog.Logger, router Router, store storage.Storage, opts ...FuncOpt) (*metricServer, error) {
	srv := &metricServer{
		log:   log,
		store: store,
	}

	for _, opt := range opts {
		opt(srv)
	}

	err := srv.configure(router, store, srv.storePath)
	if err != nil {
		return nil, err
	}

	return srv, nil
}

func (s *metricServer) Start() {
	// загрузка из файла
	if s.isRestore {
		err := s.restore()
		if err != nil {
			log.Printf("err reStore %v\n", err)
		}
	}

	s.log.Info("start server", slog.String("addr", s.server.Addr))

	ctxMain, cancelMain := context.WithCancel(context.Background())

	// регистрируем функции для отмены
	s.registerOnShutdown(cancelMain)

	// слушаем сервер
	go s.listenAndServe()

	// старт функции для периодического сохранения
	go s.savedByContex(ctxMain)

	// ловим сигналы выхода
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	// блокируем
	<-ctx.Done()
	log.Println("signal [Ctrl+C]")

	// останавливаем сервер с конкетстом
	s.log.Info("server stopped....")
	if err := s.stop(ctxMain); err != nil {
		s.log.Error("err stopped server:", err)
	}

	s.log.Info("server stop OK")
	os.Exit(0)
}

func (s *metricServer) stop(ctxMain context.Context) error {
	// контекс для отмены на 10сек
	timeoutCtx, cancel := context.WithTimeout(ctxMain, 10*time.Second)
	defer cancel()

	go s.shutdown(timeoutCtx)

	// блокируем
	<-timeoutCtx.Done()

	switch timeoutCtx.Err() {
	case context.Canceled:
		return nil
	case context.DeadlineExceeded:
		return errors.New("deadLine")
	default:
		return timeoutCtx.Err()
	}
}

// Востановить метрики из файла
func (s *metricServer) restore() error {
	arr, err := s.consumer.ReadMetric()
	if err != nil {
		return err
	}

	for i := range arr {
		if err := s.store.Set(arr[i]); err != nil {
			return err
		}
	}

	return nil
}

func (s *metricServer) savedByContex(ctx context.Context) {
	for {
		time.Sleep(time.Duration(s.storeInt) * time.Second)
		select {
		case <-ctx.Done():
			return
		default:
			if err := s.saved(); err != nil {
				log.Printf("Err save metrics %v\n", err)
			}
		}
	}
}

func (s *metricServer) registerOnShutdown(cancel func()) {
	s.server.RegisterOnShutdown(s.shutdownFunc(cancel))
}

func (s *metricServer) shutdownFunc(cancelFn func()) func() {
	return func() {
		if err := s.saved(); err != nil {
			log.Printf("Err save metrics %v\n", err)
		}

		if err := s.consumer.Close(); err != nil {
			log.Printf("Err close consumer %v\n", err)
		}
		if err := s.producer.Close(); err != nil {
			log.Printf("Err close producer %v\n", err)
		}

		cancelFn()
	}
}

func (s *metricServer) saved() error {
	err := s.producer.Trunc()
	if err != nil {
		return err
	}

	arr := s.store.List()
	for _, m := range arr {
		err := s.producer.WriteMetric(&m)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *metricServer) shutdown(timeoutCtx context.Context) {
	if err := s.server.Shutdown(timeoutCtx); err != nil {
		log.Printf("err shutdown %v\n", err)
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
		s.storeInt = interval
	}
}

func SetStorePath(path string) FuncOpt {
	return func(s *metricServer) {
		s.storePath = path
	}
}

func SetRestore(b bool) FuncOpt {
	return func(s *metricServer) {
		s.isRestore = b
	}
}
