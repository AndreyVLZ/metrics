package http

import (
	"context"
	"net/http"
)

// Http server.
type Server struct {
	server *http.Server
}

// Конфиг для http.server.
type Config struct {
	Addr string
}

func NewServer(cfg Config, h http.Handler) Server {
	return Server{
		server: &http.Server{
			// ReadHeaderTimeout: 2 * time.Second,
			Addr:    cfg.Addr,
			Handler: h,
		},
	}
}

// Запуск http.server.
func (s Server) Start() error { return s.server.ListenAndServe() }

// Остановка http.server.
func (s Server) Stop(ctx context.Context) error { return s.server.Shutdown(ctx) }
