package metricserver

import (
	"log"
	"net/http"
)

type FuncOpt func(*metricServer)

type metricServer struct {
	handler http.Handler
	addr    string
}

func New(handler http.Handler, opts ...FuncOpt) *metricServer {
	srv := &metricServer{
		handler: handler,
	}

	for _, opt := range opts {
		opt(srv)
	}

	return srv
}

func (s *metricServer) Start() error {
	log.Printf("start server %v\n", s.addr)
	return http.ListenAndServe(s.addr, s.handler)
}

func SetAddr(addr string) FuncOpt {
	return func(s *metricServer) {
		s.addr = addr
	}
}
