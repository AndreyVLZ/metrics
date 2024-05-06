package http

import (
	"context"
	"net/http"
)

type Server struct {
	server *http.Server
}

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

func (s Server) Start() error                   { return s.server.ListenAndServe() }
func (s Server) Stop(ctx context.Context) error { return s.server.Shutdown(ctx) }
