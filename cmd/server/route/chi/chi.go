package chi

import (
	"fmt"
	"net/http"

	"github.com/AndreyVLZ/metrics/cmd/server/route/handlers"
	"github.com/AndreyVLZ/metrics/cmd/server/route/mainhandler"
	"github.com/AndreyVLZ/metrics/internal/storage"
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

func (s *chiMux) SetStore(store storage.Storage) http.Handler {
	s.setHandlers(
		mainhandler.NewMainHandlers(
			store,
			NewChiHandle(),
		),
	)
	return s.mux
}

func (s *chiMux) setHandlers(mh handlers.Handlers) {
	updateEndPoint := fmt.Sprintf(
		"/{%s}/{%s}/{%s}",
		TypeChiConst, NameChiConst, ValueChiConst,
	)
	valueEndPoint := fmt.Sprintf(
		"/{%s}/{%s}",
		TypeChiConst, NameChiConst,
	)

	s.mux.Route("/", func(r chi.Router) {
		r.Get("/", mh.ListHandler)
		r.Route("/update", func(r chi.Router) {
			r.Post(
				updateEndPoint,
				mh.PostUpdateHandler().ServeHTTP,
			)
		})

		r.Route("/value", func(r chi.Router) {
			r.Get(
				valueEndPoint,
				mh.GetValueHandler().ServeHTTP,
			)
			r.Post(
				valueEndPoint,
				mh.PostValueHandler().ServeHTTP,
			)
		})
	})
}
