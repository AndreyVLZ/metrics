package route

import (
	"net/http"

	"github.com/AndreyVLZ/metrics/cmd/server/route/chi"
	"github.com/AndreyVLZ/metrics/cmd/server/route/mainhandler"
	"github.com/AndreyVLZ/metrics/cmd/server/route/servemux"
	"github.com/AndreyVLZ/metrics/internal/storage"
)

func NewServeMux(store storage.Storage) http.Handler {
	return servemux.New(
		mainhandler.NewMainHandlers(
			store,
			servemux.NewServeMuxHandle(),
		),
	)
}

func NewChi(store storage.Storage) http.Handler {
	return chi.New(
		mainhandler.NewMainHandlers(
			store,
			chi.NewChiHandle(),
		),
	)
}
