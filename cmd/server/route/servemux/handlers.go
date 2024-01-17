package servemux

import (
	"net/http"
	"strings"

	"github.com/AndreyVLZ/metrics/cmd/server/urlpath"
	"github.com/AndreyVLZ/metrics/internal/metric"
)

type serveMuxHandle struct{}

func NewServeMuxHandle() *serveMuxHandle {
	return &serveMuxHandle{}
}

func (h *serveMuxHandle) GetMetricDBFromRequest(req *http.Request) (metric.MetricDB, error) {
	arrPath := strings.Split(req.URL.Path[1:], "/")
	getURLPath := urlpath.NewGetURLPath(arrPath[1:]...)

	if err := getURLPath.Validate(); err != nil {
		return metric.MetricDB{}, err
	}

	return metric.URLParse(
		getURLPath.Type(),
		getURLPath.Name(),
		"0",
	)
}

func (h *serveMuxHandle) GetUpdateMetricDBFromRequest(req *http.Request) (metric.MetricDB, error) {
	arrPath := strings.Split(req.URL.Path[1:], "/")
	updateURLPath := urlpath.NewUpdateURLPath(arrPath[1:]...)

	if err := updateURLPath.Validate(); err != nil {
		return metric.MetricDB{}, err
	}

	return metric.URLParse(
		updateURLPath.Type(),
		updateURLPath.Name(),
		updateURLPath.Value(),
	)
}
