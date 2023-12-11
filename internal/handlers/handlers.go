package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/AndreyVLZ/metrics/internal/storage"
)

var ErrNoCorrectURLPath error = errors.New("no correct url path")
var ErrOnlyPostRequest error = errors.New("only Post")

type metricHandler struct {
	store      storage.Storage
	conentType string
}

func NewMetricHandler(store storage.Storage) *metricHandler {
	return &metricHandler{
		store:      store,
		conentType: "text/plain",
	}
}

func (h *metricHandler) UpdateHandler(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(rw, ErrOnlyPostRequest.Error(), http.StatusBadRequest)
		return
	}

	var err error

	h.setContentType(rw)

	arrPath, err := parseURLPath(req.URL.Path)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	err = h.store.Set(arrPath[0], arrPath[1], arrPath[2])
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	h.store.GaugeRepo().Range()
	h.store.CounterRepo().Range()
	fmt.Println("OK")
}

// setContentType Установка заголовка Content-Type из contentType для текущего ответа
func (h *metricHandler) setContentType(rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", h.conentType)
}

// parseURLPath Парсинг строки URL
func parseURLPath(path string) ([]string, error) {
	arrPath := strings.Split(path[1:], "/")

	if len(arrPath) != 4 {
		return nil, ErrNoCorrectURLPath
	}

	return arrPath[1:4], nil
}
