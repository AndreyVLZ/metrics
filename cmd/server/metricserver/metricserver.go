package metricserver

import (
	"log/slog"
	"net/http"

	"github.com/AndreyVLZ/metrics/cmd/server/middleware"
)

type FuncOpt func(*metricServer)

type metricServer struct {
	handler http.Handler
	addr    string
	log     *slog.Logger
}

func New(log *slog.Logger, handler http.Handler, opts ...FuncOpt) *metricServer {
	srv := &metricServer{
		handler: handler,
		log:     log,
	}

	for _, opt := range opts {
		opt(srv)
	}

	return srv
}

func (s *metricServer) Start() error {
	s.log.Info("start server", slog.String("addr", s.addr))
	/*
		child := s.log.With(
			slog.Group("program_info",
				slog.Int("pid", os.Getpid()),
			),
		)

		child.Info("test")
		//log.Printf("start server %v\n", s.addr)
	*/
	return http.ListenAndServe(s.addr,
		middleware.Logging(s.log, s.handler),
	)
}

func SetAddr(addr string) FuncOpt {
	return func(s *metricServer) {
		s.addr = addr
	}
}
