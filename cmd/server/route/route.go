package route

import (
	"fmt"
	"net/http"

	"github.com/AndreyVLZ/metrics/cmd/server/middleware"
	"github.com/go-chi/chi/v5"
)

type Handlers interface {
	UpdateHandler(http.ResponseWriter, *http.Request)
	GetValueHandler(http.ResponseWriter, *http.Request)
	ListHandler(http.ResponseWriter, *http.Request)
}

// serveMux
type serveMux struct {
	mux *http.ServeMux
}

func NewServeMux() *serveMux {
	return &serveMux{
		mux: http.NewServeMux(),
	}
}

func (s *serveMux) SetHandlers(mh Handlers) http.Handler {
	s.mux.Handle("/",
		middleware.Method(http.MethodGet,
			mh.ListHandler,
		),
	)

	s.mux.Handle("/value/",
		middleware.Method(http.MethodGet,
			mh.GetValueHandler,
		),
	)

	s.mux.Handle("/update/",
		middleware.Method(http.MethodPost,
			middleware.ContentType("text/plain",
				mh.UpdateHandler,
			),
		),
	)

	return s.mux
}

const (
	TypeChiConst  = "typeStr"
	NameChiConst  = "name"
	ValueChiConst = "val"
)

// chiMux
type chiMux struct {
	mux *chi.Mux
}

func NewChiMux() *chiMux {
	return &chiMux{
		mux: chi.NewRouter(),
	}
}

func (c *chiMux) SetHandlers(h Handlers) http.Handler {
	endPoint := fmt.Sprintf("/{%s}/{%s}/{%s}", TypeChiConst, NameChiConst, ValueChiConst)
	c.mux.Route("/", func(r chi.Router) {
		r.Get("/", h.ListHandler)
		r.Route("/update", func(r chi.Router) {
			r.Post(endPoint,
				middleware.ContentType("text/plain",
					h.UpdateHandler,
				),
			)
		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/{typeStr}/{name}", h.GetValueHandler)
		})
	})

	return c.mux
}
