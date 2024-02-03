package servemux

import (
	"net/http"

	"github.com/AndreyVLZ/metrics/cmd/server/route/handlers"
	"github.com/AndreyVLZ/metrics/cmd/server/route/middleware"
)

type serveMux struct {
	mux *http.ServeMux
}

func New() *serveMux {
	return &serveMux{
		mux: http.NewServeMux(),
	}
}

func (s *serveMux) SetHandlers(mh handlers.Handlers) http.Handler {
	s.mux.Handle("/",
		middleware.Get(
			mh.ListHandler(),
		),
	)

	s.mux.Handle("/ping",
		middleware.Get(
			mh.PingHandler(),
		),
	)

	s.mux.Handle("/update/",
		middleware.Post(
			mh.PostUpdateHandler(),
		),
	)

	s.mux.Handle("/updates/",
		middleware.Post(
			mh.PostUpdatesHandler(),
		),
	)

	s.mux.Handle("/value/",
		middleware.Methods(
			middleware.Get(
				mh.GetValueHandler(),
			),
			middleware.Post(
				mh.PostValueHandler(),
			),
		),
	)

	return s.mux
}
