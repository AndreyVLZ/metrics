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
	cfg      *Config
	api      iAPI
	services []IService
	log      *slog.Logger
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
		m.Gzip(
			m.Hash(
				cfg.key,
				mux,
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
	srv.log.LogAttrs(ctx,
		slog.LevelInfo, "start server",
		slog.String("addr", srv.cfg.addr),
		slog.Group("flags",
			slog.Int("storeInterval", srv.cfg.storeInt),
			slog.String("storePath", srv.cfg.storePath),
			slog.Bool("restore", srv.cfg.isRestore),
			slog.String("dbDNS", srv.cfg.dbDNS),
			slog.String("key", srv.cfg.key),
		),
	)

	for i := range srv.services {
		if err := srv.services[i].Start(ctx); err != nil {
			return fmt.Errorf("%w", err)
		}

		srv.log.Info("services Start", "name", srv.services[i].Name())
	}

	return srv.api.Start()
}

// Остановка сервера.
func (srv *Server) Stop(ctx context.Context) error {
	errs := make([]error, 0, len(srv.services)+1)
	if err := srv.api.Stop(ctx); err != nil {
		errs = append(errs, err)
	}

	for _, srv := range srv.services {
		if err := srv.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("service [%s] err: %w", srv.Name(), err))
		}
	}

	return errors.Join(errs...)
}
