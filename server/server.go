package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/AndreyVLZ/metrics/internal/store"
	api "github.com/AndreyVLZ/metrics/server/http"
	m "github.com/AndreyVLZ/metrics/server/http/middleware"
	"github.com/AndreyVLZ/metrics/server/service"
)

// Интерфейс для http.Server.
type iAPI interface {
	Start() error
	Stop(ctx context.Context) error
}

// Интерфейс для сервиса.
type IService interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Сервер.
type Server struct {
	api      iAPI
	cfg      *Config
	log      *slog.Logger
	services []IService
}

func New(cfg *Config, log *slog.Logger) Server {
	store := store.New(
		store.Config{
			ConnDB:    cfg.dbDNS,
			StorePath: cfg.storePath,
			IsRestore: cfg.isRestore,
			StoreInt:  cfg.storeInt,
		},
	)

	srv := service.New(store)
	mux := api.NewRoute(srv, log)
	handler := m.Logging(log,
		m.Decrypt(cfg.privateKey,
			m.Gzip(
				m.Hash(cfg.key,
					mux,
				),
			),
		),
	)

	httpServer := api.NewServer(
		api.Config{
			Addr: cfg.addr,
		},
		handler,
	)

	return Server{
		cfg:      cfg,
		api:      httpServer,
		services: []IService{store},
		log:      log,
	}
}

// Запуск сервера.
func (srv *Server) Start(ctx context.Context) error {
	srv.log.DebugContext(ctx, "start server",
		slog.String("addr", srv.cfg.addr),
		slog.Group("flags",
			slog.String("storeInterval", srv.cfg.storeInt.String()),
			slog.String("storePath", srv.cfg.storePath),
			slog.Bool("restore", srv.cfg.isRestore),
			slog.String("dbDNS", srv.cfg.dbDNS),
			slog.String("key", srv.cfg.key),
			slog.String("privateKeyPath", srv.cfg.cryptoKeyPath),
		),
	)

	for i := range srv.services {
		if err := srv.services[i].Start(ctx); err != nil {
			return fmt.Errorf("%w", err)
		}

		srv.log.DebugContext(ctx, "services started", "name", srv.services[i].Name())
	}

	return srv.api.Start()
}

// Остановка сервера.
func (srv *Server) Stop(ctx context.Context) error {
	errs := make([]error, 0, len(srv.services)+1)
	if err := srv.api.Stop(ctx); err != nil {
		errs = append(errs, err)
	}

	for i := range srv.services {
		if err := srv.services[i].Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("service [%s] err: %w", srv.services[i].Name(), err))
		} else {
			srv.log.DebugContext(ctx, "services stopped", "name", srv.services[i].Name())
		}
	}

	return errors.Join(errs...)
}
