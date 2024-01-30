package mainhandler

import (
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
	</head>
	<body>
	<ol type="1">
	{{ range . }}
	<li><strong>{{ .Name }}</strong>[{{.Type}}]: {{ .String}}</li>
	{{ end }}
	</ol>
	</body>
</html>
{{end}}`

type EmbedingHandlers interface {
	GetMetricDBFromRequest(*http.Request) (metric.MetricDB, error)
	GetUpdateMetricDBFromRequest(*http.Request) (metric.MetricDB, error)
}

type mainHandlers struct {
	tmpls *template.Template
	store storage.Storage
	EmbedingHandlers
}
type responseWriter struct {
	w http.ResponseWriter
}

func (rw *responseWriter) WriteString(dataStr string) error {
	_, err := rw.w.Write([]byte(dataStr))

	return err
}

func (rw *responseWriter) WriteAsJSON(newMetricDB metric.MetricDB) error {
	return json.NewEncoder(rw.w).Encode(newMetricDB)
}

type funcHandle func(http.ResponseWriter, *http.Request) (int, error)

func (mh *mainHandlers) handlerFunc(fn funcHandle) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		httpStatus, err := fn(rw, req)
		if err != nil {
			http.Error(rw, err.Error(), httpStatus)
		}
	})
}

func NewMainHandlers(store storage.Storage, embHandlers EmbedingHandlers) *mainHandlers {
	return &mainHandlers{
		tmpls:            template.Must(template.New("metrics").Parse(tpls)),
		store:            store,
		EmbedingHandlers: embHandlers,
	}
}

func (mh *mainHandlers) PingHandler() http.Handler {
	return mh.handlerFunc(func(w http.ResponseWriter, req *http.Request) (int, error) {
		if err := mh.store.Ping(); err != nil {
			return http.StatusInternalServerError, err
		}
		return http.StatusOK, nil
	})
}

func (mh *mainHandlers) ListHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", TextHTMLConst)
	rw.WriteHeader(http.StatusOK)
	err := mh.tmpls.ExecuteTemplate(rw, "List", mh.store.List(req.Context()))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

// Парсинг метрики из URL
// Чтение метрики из хранилища
// Запись метрики в Reader
func (mh *mainHandlers) GetValueHandler() http.Handler {
	return mh.handlerFunc(func(w http.ResponseWriter, req *http.Request) (int, error) {
		rw := responseWriter{w}
		metricDB, err := mh.EmbedingHandlers.GetMetricDBFromRequest(req)
		if err != nil {
			if err.Error() == urlpath.ErrEmptyNameField.Error() {
				return http.StatusNotFound, err
			}
			return http.StatusBadRequest, err
		}

		newMetricDB, err := mh.store.Get(req.Context(), metricDB)
		if err != nil {
			return http.StatusNotFound, err
		}

		// NOTE
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		err = rw.WriteString(newMetricDB.String())
		if err != nil {
			return http.StatusNotFound, err
		}

		return 0, nil
	})
}

// Парсинг метрики из Body
// Чтение метрики из хранилища
// Запись метрики в Body
func (mh *mainHandlers) PostValueHandler() http.Handler {
	return mh.handlerFunc(func(w http.ResponseWriter, req *http.Request) (int, error) {
		rw := responseWriter{w}

		metricDB, err := metricFromBoby(req.Body)

		if err != nil {
			return http.StatusNotFound, err
		}

		newMetricDB, err := mh.store.Get(req.Context(), metricDB)
		if err != nil {
			return http.StatusNotFound, err
		}

		w.Header().Set("Content-Type", ApplicationJSONConst)
		rw.w.WriteHeader(http.StatusOK)

		err = rw.WriteAsJSON(newMetricDB)
		if err != nil {
			return http.StatusBadRequest, err
		}

		return 0, nil
	})
}

func (mh *mainHandlers) PostUpdatesHandler() http.Handler {
	return mh.handlerFunc(func(rw http.ResponseWriter, req *http.Request) (int, error) {
		if req.Header.Get("Content-Type") != ApplicationJSONConst {
			return http.StatusBadRequest, errors.New("contentType is not appJson")
		}

		metricsDB, err := metricsFromBoby(req.Body)
		if err != nil {
			return http.StatusBadRequest, err
		}

		rw.Header().Set("Content-Type", ApplicationJSONConst)
		rw.WriteHeader(http.StatusOK)

		if err := mh.store.UpdateBatch(req.Context(), metricsDB); err != nil {
			return http.StatusBadRequest, err
		}

		return 0, nil
	})
}

func (mh *mainHandlers) PostUpdateHandler() http.Handler {
	return mh.handlerFunc(func(rw http.ResponseWriter, req *http.Request) (int, error) {
		if req.Header.Get("Content-Type") == ApplicationJSONConst {
			return mh.postJSONUpdate(rw, req)
		}

		return mh.postUpdate(rw, req)
	})
}

// Парсинг метрики из Body
// Запись метрики в хранилище
// Чтение метрики из хранилища
// Запись метрики в Body
func (mh *mainHandlers) postJSONUpdate(w http.ResponseWriter, req *http.Request) (int, error) {
	rw := responseWriter{w}

	metricDB, err := metricFromBoby(req.Body)
	if err != nil {
		return http.StatusNotFound, err
	}

	newMetricDB, err := mh.store.Update(req.Context(), metricDB)
	if err != nil {
		return http.StatusBadRequest, err
	}

	rw.w.Header().Set("Content-Type", ApplicationJSONConst)
	rw.w.WriteHeader(http.StatusOK)

	err = rw.WriteAsJSON(newMetricDB)
	if err != nil {
		return http.StatusBadRequest, err
	}

	return 0, nil
}

// Парсинг метрики из URL
// Запись метрики в хранилище
func (mh *mainHandlers) postUpdate(w http.ResponseWriter, req *http.Request) (int, error) {
	metricDB, err := mh.EmbedingHandlers.GetUpdateMetricDBFromRequest(req)
	if err != nil {
		if err.Error() == urlpath.ErrEmptyNameField.Error() {
			return http.StatusNotFound, err
		}
		return http.StatusBadRequest, err
	}

	_, err = mh.store.Update(req.Context(), metricDB)
	if err != nil {
		return http.StatusNotFound, err
	}

	return 0, nil
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

func metricFromBoby(body io.ReadCloser) (metric.MetricDB, error) {
	defer body.Close()

	var metricJSON MetricsJSON
	if err := json.NewDecoder(body).Decode(&metricJSON); err != nil {
		return metric.MetricDB{}, err
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
		return metric.MetricDB{}, errors.New("not type support")
	}

	return metric.NewMetricDB(metricJSON.ID, val), nil
}
