package handlers

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/AndreyVLZ/metrics/cmd/server/route"
	path "github.com/AndreyVLZ/metrics/cmd/server/urlpath"
	"github.com/AndreyVLZ/metrics/internal/storage"
	"github.com/go-chi/chi/v5"
)

const tpls = `
{{define "List"}}
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
</html>
{{end}}`

type MainHandle struct {
	tmpls *template.Template
	store storage.Storage
}

func NewMainHandle(store storage.Storage) *MainHandle {
	tmpls := template.Must(template.New("metrics").Parse(tpls))

	return &MainHandle{
		tmpls: tmpls,
		store: store,
	}
}

func (h *MainHandle) GetValueHandler(rw http.ResponseWriter, req *http.Request) {
	arrPath := strings.Split(req.URL.Path[1:], "/")
	h.getValue(path.NewGetURLPath(arrPath[1:]), rw)
}

func (h *MainHandle) UpdateHandler(rw http.ResponseWriter, req *http.Request) {
	arrPath := strings.Split(req.URL.Path[1:], "/")
	h.updateValue(path.NewUpdateURLPath(arrPath[1:]), rw)
}

func (h *MainHandle) ListHandler(rw http.ResponseWriter, req *http.Request) {
	data := struct {
		GRepo map[string]string
		CRepo map[string]string
	}{
		GRepo: h.store.GaugeRepo().List(),
		CRepo: h.store.CounterRepo().List(),
	}

	err := h.tmpls.ExecuteTemplate(rw, "List", data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h *MainHandle) getValue(getURLPath *path.GetURLPath, rw http.ResponseWriter) {
	if err := getURLPath.Validate(); err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	val, err := h.store.Get(getURLPath.Type(), getURLPath.Name())
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	_, err = rw.Write([]byte(val))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}
}

func (h *MainHandle) updateValue(urlPath *path.UpdateURLPath, rw http.ResponseWriter) {
	if err := urlPath.Validate(); err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	err := h.store.Set(urlPath.Type(), urlPath.Name(), urlPath.Value())
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}
}

type chiHandle struct {
	*MainHandle
}

func NewChiHandle(store storage.Storage) *chiHandle {
	return &chiHandle{
		NewMainHandle(store),
	}
}

func (h *chiHandle) GetValueHandler(rw http.ResponseWriter, req *http.Request) {
	getURLPath := path.NewGetURLPath([]string{
		chi.URLParam(req, route.TypeChiConst),
		chi.URLParam(req, route.NameChiConst),
	})

	h.getValue(getURLPath, rw)
}

func (h *chiHandle) UpdateHandler(rw http.ResponseWriter, req *http.Request) {
	updateURLPath := path.NewUpdateURLPath([]string{
		chi.URLParam(req, route.TypeChiConst),
		chi.URLParam(req, route.NameChiConst),
		chi.URLParam(req, route.ValueChiConst),
	})

	h.updateValue(updateURLPath, rw)
}
