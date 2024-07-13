package api

import (
	"context"
	"errors"
	"log/slog"

	"github.com/AndreyVLZ/metrics/internal/model"
	"github.com/AndreyVLZ/metrics/server/api/grps"
	"github.com/AndreyVLZ/metrics/server/api/http"
	"github.com/AndreyVLZ/metrics/server/config"
)

type starter interface {
	Start() error
}

type iServer interface {
	starter
	Stop(ctx context.Context) error
}

type storager interface {
	Ping() error
	Get(ctx context.Context, mInfo model.Info) (model.Metric, error)
	Update(ctx context.Context, met model.Metric) (model.Metric, error)
	List(ctx context.Context) ([]model.Metric, error)
	AddBatch(ctx context.Context, arr []model.Metric) error
}

type Server struct {
	serverHTTP iServer
	serverGRPC iServer
}

func New(cfg *config.Config, store storager, log *slog.Logger) *Server {
	return &Server{
		serverHTTP: http.NewServer(cfg, store, log),
		serverGRPC: grps.NewServer(cfg, store, log),
	}
}

// Start Запуск HTTP и GRPC серверов.
func (s *Server) Start() error {
	chErr := make(chan error)
	defer close(chErr)

	startServers(chErr, s.serverHTTP, s.serverGRPC)

	for err := range chErr {
		if err != nil {
			return err
		}
	}

	return nil
}

// Stop Остановка HTTP и GRPC серверов.
func (s *Server) Stop(ctx context.Context) error {
	return errors.Join(
		s.serverHTTP.Stop(ctx),
		s.serverGRPC.Stop(ctx),
	)
}

func startServers(chErr chan<- error, servers ...starter) {
	for i := range servers {
		go func(srv starter, ch chan<- error) {
			if err := srv.Start(); err != nil {
				ch <- err
			}
		}(servers[i], chErr)
	}
}
