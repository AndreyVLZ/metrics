package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/AndreyVLZ/metrics/internal/model"
	m "github.com/AndreyVLZ/metrics/server/api/http/middleware"
	"github.com/AndreyVLZ/metrics/server/api/http/service"
	"github.com/AndreyVLZ/metrics/server/config"
)

type storager interface {
	Ping() error
	Get(ctx context.Context, mInfo model.Info) (model.Metric, error)
	Update(ctx context.Context, met model.Metric) (model.Metric, error)
	List(ctx context.Context) ([]model.Metric, error)
	AddBatch(ctx context.Context, arr []model.Metric) error
}

// Http server.
type Server struct {
	server *http.Server
}

func NewServer(cfg *config.Config, store storager, log *slog.Logger) *Server {
	srv := service.New(store)
	mux := NewRoute(srv, log)

	handler := m.Logging(log,
		m.Subnet(cfg.TrustedSubnet,
			m.Decrypt(cfg.PrivateKey,
				m.Gzip(
					m.Hash(cfg.Key,
						mux,
					),
				),
			),
		),
	)

	return &Server{
		server: &http.Server{
			Addr:    cfg.Addr,
			Handler: handler,
		},
	}
}

// Запуск http.server.
func (s *Server) Start() error { return s.server.ListenAndServe() }

// Остановка http.server.
func (s *Server) Stop(ctx context.Context) error { return s.server.Shutdown(ctx) }
