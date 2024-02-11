package route

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/AndreyVLZ/metrics/cmd/server/route/chi"
	"github.com/AndreyVLZ/metrics/cmd/server/route/handlers"
	"github.com/AndreyVLZ/metrics/cmd/server/route/mainhandler"
	"github.com/AndreyVLZ/metrics/cmd/server/route/middleware"
	"github.com/AndreyVLZ/metrics/cmd/server/route/servemux"
	"github.com/AndreyVLZ/metrics/internal/storage"
)

type RouteType string

const (
	RouteTypeServeMux RouteType = "servemux"
	RouteTypeChi      RouteType = "chi"
)

var ErrNotSupportTypeRoute = errors.New("not suppot type router")

type Route interface {
	SetHandlers(mh handlers.Handlers) http.Handler
}

type Config struct {
	RouteType RouteType
	Store     storage.Storage
	Log       *slog.Logger
	SecretKey string
}

func New(cfg Config) (http.Handler, error) {
	mh := mainhandler.New(cfg.Store)
	var r Route
	switch cfg.RouteType {
	case RouteTypeServeMux:
		mh.EmbedingHandlers = servemux.NewServeMuxHandle()
		r = servemux.New()
	case RouteTypeChi:
		mh.EmbedingHandlers = chi.NewChiHandle()
		r = chi.New()
	default:
		return nil, ErrNotSupportTypeRoute
	}

	return middleware.Logging(
			cfg.Log,
			middleware.Gzip(
				middleware.Hash(
					cfg.SecretKey,
					r.SetHandlers(mh),
				),
			),
		),
		nil
}
