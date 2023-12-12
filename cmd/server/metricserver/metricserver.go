package metricserver

import (
	"net/http"
)

type Handlers interface {
	UpdateHandler(http.ResponseWriter, *http.Request)
	GetValueHandler(http.ResponseWriter, *http.Request)
	ListHandler(http.ResponseWriter, *http.Request)
}

type Router interface {
	SetHandlers(Handlers) http.Handler
}

type metricServer struct {
	handler http.Handler
}

func New(handler http.Handler) *metricServer {
	return &metricServer{
		handler: handler,
	}
}

func (s *metricServer) Start() error {
	return http.ListenAndServe("localhost:8080", s.handler)
}
