package servemux

import (
	"net/http"

	"github.com/AndreyVLZ/metrics/cmd/server/route/handlers"
	"github.com/AndreyVLZ/metrics/cmd/server/route/middleware"
)

type serveMux struct {
	mux *http.ServeMux
}

func New(mh handlers.Handlers) http.Handler {
	s := &serveMux{
		mux: http.NewServeMux(),
	}

	s.setHandlers(mh)

	return s.mux
}

func (s *serveMux) setHandlers(mh handlers.Handlers) {
	s.mux.Handle("/",
		middleware.Get(
			middleware.GzipMiddleware(
				http.HandlerFunc(mh.ListHandler),
			),
		),
	)

	s.mux.Handle("/update/",
		middleware.Post(
			middleware.GzipMiddleware(
				mh.PostUpdateHandler().ServeHTTP,
			),
		),
	)

	s.mux.Handle("/value/",
		middleware.Methods(
			middleware.Get(
				mh.GetValueHandler(),
			),
			middleware.Post(
				middleware.GzipMiddleware(
					mh.PostValueHandler().ServeHTTP,
				),
			),
		),
	)
}
