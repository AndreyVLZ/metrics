package route

import (
	"errors"
	"net/http"

	"github.com/AndreyVLZ/metrics/cmd/server/route/chi"
	"github.com/AndreyVLZ/metrics/cmd/server/route/servemux"
	"github.com/AndreyVLZ/metrics/internal/storage"
)

type RouteType string

const (
	RouteTypeServeMux RouteType = "servemux"
	RouteTypeChi      RouteType = "chi"
)

type Route interface {
	SetStore(storage.Storage) http.Handler
}

type Config struct {
	RouteType RouteType
}

func New(cfg Config) (Route, error) {
	switch cfg.RouteType {
	case RouteTypeServeMux:
		return servemux.New(), nil
	case RouteTypeChi:
		return chi.New(), nil
	default:
		return nil, errors.New("not suppot type router")
	}
}
