package metricserver

import (
	"net/http"

	"github.com/AndreyVLZ/metrics/cmd/server/config"
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
	conf    config.Config
}

func New(handler http.Handler, conf config.Config) *metricServer {
	return &metricServer{
		handler: handler,
		conf:    conf,
	}
}

func (s *metricServer) Start() error {
	return http.ListenAndServe(s.conf.Addr, s.handler)
}
