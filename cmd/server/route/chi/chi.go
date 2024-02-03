package chi

import (
	"fmt"
	"net/http"

	"github.com/AndreyVLZ/metrics/cmd/server/route/handlers"
	"github.com/go-chi/chi/v5"
)

const (
	TypeChiConst  = "typeStr"
	NameChiConst  = "name"
	ValueChiConst = "val"
)

type chiMux struct {
	mux *chi.Mux
}

func New() *chiMux {
	return &chiMux{
		mux: chi.NewRouter(),
	}
}

func (s *chiMux) SetHandlers(mh handlers.Handlers) http.Handler {
	updateEndPoint := fmt.Sprintf(
		"/{%s}/{%s}/{%s}",
		TypeChiConst, NameChiConst, ValueChiConst,
	)
	valueEndPoint := fmt.Sprintf(
		"/{%s}/{%s}",
		TypeChiConst, NameChiConst,
	)
	_ = updateEndPoint
	s.mux.Route("/", func(r chi.Router) {
		r.Get("/", mh.ListHandler().ServeHTTP)

		r.Get("/ping", mh.PingHandler().ServeHTTP)

		//r.Post("/update/", mh.PostUpdateHandler().ServeHTTP)

		r.Post("/updates/", mh.PostUpdatesHandler().ServeHTTP)

		r.Route("/value", func(r chi.Router) {
			r.Get(
				valueEndPoint,
				mh.GetValueHandler().ServeHTTP,
			)
			r.Post(
				"/",
				mh.PostValueHandler().ServeHTTP,
			)
		})

		r.Route("/update", func(r chi.Router) {
			r.Post(
				"/",
				mh.PostUpdateHandler().ServeHTTP,
			)
			r.Post(
				updateEndPoint,
				mh.PostUpdateHandler().ServeHTTP,
			)
		})
	})

	return s.mux
}
