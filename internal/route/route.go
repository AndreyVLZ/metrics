package route

import (
	"net/http"

	"github.com/AndreyVLZ/metrics/cmd/server/metricserver"
	"github.com/go-chi/chi/v5"
)

// serveMux
type serveMux struct {
	mux *http.ServeMux
}

func NewServeMux() *serveMux {
	return &serveMux{
		mux: http.NewServeMux(),
	}
}

func (s *serveMux) SetHandlers(mh metricserver.Handlers) http.Handler {
	s.mux.HandleFunc("/", mh.ListHandler)
	s.mux.HandleFunc("/update/", mh.UpdateHandler)
	s.mux.HandleFunc("/value/", mh.GetValueHandler)
	return s.mux
}

// chiMux
type chiMux struct {
	mux *chi.Mux
}

func NewChiMux() *chiMux {
	return &chiMux{
		mux: chi.NewRouter(),
	}
}

func (c *chiMux) SetHandlers(h metricserver.Handlers) http.Handler {
	c.mux.Route("/", func(r chi.Router) {
		r.Get("/", h.ListHandler)
		r.Route("/update", func(r chi.Router) {
			r.Post("/{typeStr}/{name}/{val}", h.UpdateHandler)
		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/{typeStr}/{name}", h.GetValueHandler)
		})
	})

	return c.mux
}
