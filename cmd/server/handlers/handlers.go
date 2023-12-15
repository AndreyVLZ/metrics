package handlers

import (
	"errors"
	"html/template"
	"net/http"
	"strings"

	"github.com/AndreyVLZ/metrics/internal/storage"
	"github.com/go-chi/chi/v5"
)

var ErrNoCorrectURLPath error = errors.New("no correct url path")

type metricHandler struct {
	store storage.Storage
}

func NewMetricHandler(store storage.Storage) *metricHandler {
	return &metricHandler{
		store: store,
	}
}

type chiHandler struct {
	store storage.Storage
}

func NewChiHandler(store storage.Storage) *chiHandler {
	return &chiHandler{
		store: store,
	}
}

func (h *chiHandler) ListHandler(rw http.ResponseWriter, req *http.Request) {
	const tpl = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>RunTime Metrics</title>
	</head>
	<body>
	{{ range $key, $value := .GRepo }}
		<li><strong>{{ $key }}</strong>: {{ $value }}</li>
	{{ end }}
	{{ range $key, $value := .CRepo }}
		<li><strong>{{ $key }}</strong>: {{ $value }}</li>
	{{ end }}
	</body>
</html>`
	t, err := template.New("webpage").Parse(tpl)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		GRepo map[string]string
		CRepo map[string]string
	}{
		GRepo: h.store.GaugeRepo().List(),
		CRepo: h.store.CounterRepo().List(),
	}

	err = t.Execute(rw, data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *chiHandler) UpdateHandler(rw http.ResponseWriter, req *http.Request) {
	typeStr := chi.URLParam(req, "typeStr")
	name := chi.URLParam(req, "name")
	val := chi.URLParam(req, "val")
	if typeStr == "" || name == "" || val == "" {
		http.Error(rw, "bad", http.StatusNotFound)
		return
	}

	err := h.store.Set(typeStr, name, val)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *chiHandler) GetValueHandler(rw http.ResponseWriter, req *http.Request) {
	typeStr := chi.URLParam(req, "typeStr")
	name := chi.URLParam(req, "name")

	if typeStr == "" || name == "" {
		http.Error(rw, "bad", http.StatusBadRequest)
		return
	}

	val, err := h.store.Get(typeStr, name)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	_, err = rw.Write([]byte(val))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (h *metricHandler) ListHandler(rw http.ResponseWriter, req *http.Request) {

	const tpl = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>RunTime Metrics</title>
	</head>
	<body>
	<ol type="1">
	{{ range $key, $value := .GRepo }}
		<li><strong>{{ $key }}</strong>: {{ $value }}</li>
	{{ end }}
	{{ range $key, $value := .CRepo }}
		<li><strong>{{ $key }}</strong>: {{ $value }}</li>
	{{ end }}
	</ol>
	</body>
</html>`
	t, err := template.New("webpage").Parse(tpl)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		GRepo map[string]string
		CRepo map[string]string
	}{
		GRepo: h.store.GaugeRepo().List(),
		CRepo: h.store.CounterRepo().List(),
	}

	err = t.Execute(rw, data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *metricHandler) GetValueHandler(rw http.ResponseWriter, req *http.Request) {
	arrPath, err := parseURLPath(req.URL.Path, 3)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	val, err := h.store.Get(arrPath[0], arrPath[1])
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	_, err = rw.Write([]byte(val))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	//rw.WriteHeader(http.StatusOK)
}

func (h *metricHandler) UpdateHandler(rw http.ResponseWriter, req *http.Request) {
	var err error

	arrPath, err := parseURLPath(req.URL.Path, 4)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	err = h.store.Set(arrPath[0], arrPath[1], arrPath[2])
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

// parseURLPath Парсинг строки URL
func parseURLPath(path string, l int) ([]string, error) {
	arrPath := strings.Split(path[1:], "/")

	if len(arrPath) != l {
		return nil, ErrNoCorrectURLPath
	}

	return arrPath[1:l], nil
}
