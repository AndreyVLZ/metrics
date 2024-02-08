package mainhandler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"net/http"

	"github.com/AndreyVLZ/metrics/cmd/server/urlpath"
	"github.com/AndreyVLZ/metrics/internal/metric"
	"github.com/AndreyVLZ/metrics/internal/storage"
)

const (
	ApplicationJSONConst = "application/json"
	TextPlainConst       = "text/plain"
	TextHTMLConst        = "text/html"
)

const tpls = `
{{define "List"}}
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>RunTime Metrics</title>
	<link rel="shortcut icon" href="http://www.example.com/my_empty_resource"/>
	</head>
	<body>
	<ol type="1">
	{{ range . }}
	<li><strong>{{ .Name }}</strong>[{{.Type}}]: {{ .String}}</li>
	{{ end }}
	</ol>
	</body>
</html>{{end}}`

type EmbedingHandlers interface {
	GetMetricDBFromRequest(*http.Request) (metric.MetricDB, error)
	GetUpdateMetricDBFromRequest(*http.Request) (metric.MetricDB, error)
}

type mainHandlers struct {
	tmpls *template.Template
	store storage.Storage
	EmbedingHandlers
}

func New(store storage.Storage) *mainHandlers {
	return &mainHandlers{
		tmpls: template.Must(template.New("metrics").Parse(tpls)),
		store: store,
	}
}

func (mh *mainHandlers) PingHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if err := mh.store.Ping(); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func (mh *mainHandlers) ListHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var buf bytes.Buffer
		if err := mh.tmpls.ExecuteTemplate(&buf, "List", mh.store.List(req.Context())); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		rw.Header().Set("Content-Type", TextHTMLConst)
		buf.WriteTo(rw)
	})
}

// Парсинг метрики из URL
// Чтение метрики из хранилища
// Запись метрики в Reader
func (mh *mainHandlers) GetValueHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		metricDB, err := mh.EmbedingHandlers.GetMetricDBFromRequest(req)
		if err != nil {
			if errors.Is(err, urlpath.ErrEmptyNameField) {
				http.Error(rw, err.Error(), http.StatusNotFound)
				return
			}
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		newMetricDB, err := mh.store.Get(req.Context(), metricDB)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}

		h1, _ := hash([]byte(newMetricDB.String()), []byte("test"))
		_ = h1

		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if _, err = rw.Write([]byte(newMetricDB.String())); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	})
}

func hash(data []byte, key []byte) ([]byte, error) {
	h := hmac.New(sha256.New, key)
	_, err := h.Write(data)
	if err != nil {
		return nil, errors.New("err hash")
	}

	sum := h.Sum(nil)
	return sum, nil
}

// Парсинг метрики из Body
// Чтение метрики из хранилища
// Запись метрики в Body
func (mh *mainHandlers) PostValueHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		metricDB, err := metricFromBoby(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}

		newMetricDB, err := mh.store.Get(req.Context(), metricDB)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}

		rw.Header().Set("Content-Type", ApplicationJSONConst)
		if err := json.NewEncoder(rw).Encode(newMetricDB); err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
		}
	})
}

func (mh *mainHandlers) PostUpdateHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Content-Type") == ApplicationJSONConst {
			mh.postJSONUpdate(rw, req)
			return
		}
		mh.postUpdate(rw, req)
	})
}

// Парсинг метрики из Body
// Запись метрики в хранилище
// Чтение метрики из хранилища
// Запись метрики в Body
func (mh *mainHandlers) postJSONUpdate(rw http.ResponseWriter, req *http.Request) {
	metricDB, err := metricFromBoby(req.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	newMetricDB, err := mh.store.Update(req.Context(), metricDB)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.Header().Set("Content-Type", ApplicationJSONConst)
	if err := json.NewEncoder(rw).Encode(newMetricDB); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}
}

// Парсинг метрики из URL
// Запись метрики в хранилище
func (mh *mainHandlers) postUpdate(rw http.ResponseWriter, req *http.Request) {
	metricDB, err := mh.EmbedingHandlers.GetUpdateMetricDBFromRequest(req)
	if err != nil {
		if errors.Is(err, urlpath.ErrEmptyNameField) {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = mh.store.Update(req.Context(), metricDB)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}
}

func (mh *mainHandlers) PostUpdatesHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Content-Type") != ApplicationJSONConst {
			http.Error(rw, "contentType is not appJson", http.StatusBadRequest)
			return
		}

		metricsDB, err := metricsFromBoby(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		if err := mh.store.UpdateBatch(req.Context(), metricsDB); err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
		}
	})
}

func metricsFromBoby(body io.ReadCloser) ([]metric.MetricDB, error) {
	defer body.Close()

	var metricsJSON []MetricsJSON
	var metricsDB []metric.MetricDB

	if err := json.NewDecoder(body).Decode(&metricsJSON); err != nil {
		return nil, err
	}

	metricsDB = make([]metric.MetricDB, len(metricsJSON))
	for i := range metricsJSON {
		metricDB, err := NewMetricDBFromMetricJSON(metricsJSON[i])
		if err != nil {
			return nil, err
		}

		metricsDB[i] = metricDB
	}

	return metricsDB, nil
}

var ErrJSONSyntax = errors.New("err json syntax")

func metricFromBoby(body io.ReadCloser) (metric.MetricDB, error) {
	defer body.Close()
	var metricJSON MetricsJSON
	if err := json.NewDecoder(body).Decode(&metricJSON); err != nil {
		return metric.MetricDB{}, ErrJSONSyntax
	}

	metricDB, err := NewMetricDBFromMetricJSON(metricJSON)
	if err != nil {
		return metric.MetricDB{}, err
	}

	return metricDB, nil
}

// NOTE
type MetricsJSON struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func NewMetricJSONFromMetricDB(metricDB metric.MetricDB) (MetricsJSON, error) {
	metricJSON := MetricsJSON{
		ID:    metricDB.Name(),
		MType: metricDB.Type(),
	}

	switch metricDB.Type() {
	case metric.CounterType.String():
		metricJSON.Delta = new(int64)
		metricDB.ReadTo(metricJSON.Delta)
	case metric.GaugeType.String():
		metricJSON.Value = new(float64)
		metricDB.ReadTo(metricJSON.Value)
	}

	return metricJSON, nil
}

var ErrTypeNotSupport = errors.New("not type support")

func NewMetricDBFromMetricJSON(metricJSON MetricsJSON) (metric.MetricDB, error) {
	var val metric.Valuer

	switch metricJSON.MType {
	case metric.CounterType.String():
		if metricJSON.Delta == nil {
			val = metric.Counter(0)
		} else {
			val = metric.Counter(*metricJSON.Delta)
		}
	case metric.GaugeType.String():
		if metricJSON.Value == nil {
			val = metric.Gauge(0)
		} else {
			val = metric.Gauge(*metricJSON.Value)
		}
	default:
		return metric.MetricDB{}, ErrTypeNotSupport
	}

	return metric.NewMetricDB(metricJSON.ID, val), nil
}
