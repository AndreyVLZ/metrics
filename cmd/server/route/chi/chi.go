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

func New(mh handlers.Handlers) http.Handler {
	c := &chiMux{
		mux: chi.NewRouter(),
	}

	c.setHandlers(mh)

	return c.mux
}

func (c *chiMux) setHandlers(mh handlers.Handlers) http.Handler {
	updateEndPoint := fmt.Sprintf(
		"/{%s}/{%s}/{%s}",
		TypeChiConst, NameChiConst, ValueChiConst,
	)
	valueEndPoint := fmt.Sprintf(
		"/{%s}/{%s}",
		TypeChiConst, NameChiConst,
	)

	c.mux.Route("/", func(r chi.Router) {
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

	return c.mux
}
