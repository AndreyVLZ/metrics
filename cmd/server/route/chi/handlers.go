package chi

import (
	"net/http"

	"github.com/AndreyVLZ/metrics/cmd/server/urlpath"
	"github.com/AndreyVLZ/metrics/internal/metric"
	"github.com/go-chi/chi/v5"
)

//var _ mainhandler.EmbedingHandlers = &chiHandle{}

type chiHandle struct{}

func NewChiHandle() *chiHandle {
	return &chiHandle{}
}

func (h *chiHandle) GetMetricDBFromRequest(req *http.Request) (metric.MetricDB, error) {
	getURLPath := urlpath.NewGetURLPath(
		chi.URLParam(req, TypeChiConst),
		chi.URLParam(req, NameChiConst),
	)

	if err := getURLPath.Validate(); err != nil {
		return metric.MetricDB{}, nil
	}

	return metric.URLParse(
		getURLPath.Type(),
		getURLPath.Name(),
		"",
	)
}
func (h *chiHandle) GetUpdateMetricDBFromRequest(req *http.Request) (metric.MetricDB, error) {
	updateURLPath := urlpath.NewUpdateURLPath(
		chi.URLParam(req, TypeChiConst),
		chi.URLParam(req, NameChiConst),
		chi.URLParam(req, ValueChiConst),
	)

	if err := updateURLPath.Validate(); err != nil {
		return metric.MetricDB{}, nil
	}

	return metric.URLParse(
		updateURLPath.Type(),
		updateURLPath.Name(),
		updateURLPath.Value(),
	)
}
