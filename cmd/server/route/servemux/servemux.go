package servemux

import (
	"net/http"

	"github.com/AndreyVLZ/metrics/cmd/server/route/handlers"
	"github.com/AndreyVLZ/metrics/cmd/server/route/mainhandler"
	"github.com/AndreyVLZ/metrics/cmd/server/route/middleware"
	"github.com/AndreyVLZ/metrics/internal/storage"
)

type serveMux struct {
	mux   *http.ServeMux
	store storage.Storage
}

func New() *serveMux {
	return &serveMux{
		mux: http.NewServeMux(),
	}
}

func (s *serveMux) SetStore(store storage.Storage) http.Handler {
	s.setHandlers(
		mainhandler.NewMainHandlers(
			store,
			NewServeMuxHandle(),
		),
	)
	return s.mux
}

func (s *serveMux) setHandlers(mh handlers.Handlers) {
	s.mux.Handle("/",
		middleware.Get(
			middleware.GzipMiddleware(
				http.HandlerFunc(mh.ListHandler().ServeHTTP),
			),
		),
	)

	s.mux.Handle("/ping",
		middleware.Get(
			mh.PingHandler(),
		),
	)

	s.mux.Handle("/update/",
		middleware.Post(
			middleware.GzipMiddleware(
				mh.PostUpdateHandler().ServeHTTP,
			),
		),
	)

	s.mux.Handle("/updates/",
		middleware.Post(
			middleware.GzipMiddleware(
				mh.PostUpdatesHandler().ServeHTTP,
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
