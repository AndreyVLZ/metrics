package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/AndreyVLZ/metrics/internal/model"
)

const (
	ApplicationJSONConst = "application/json" // Константа для Content-Type app/json.
	TextHTMLConst        = "text/html"        // Константа для Content-Type text/html.
)

type srvUpdater interface {
	Update(ctx context.Context, met model.MetricJSON) (model.MetricJSON, error)
}

type srvBatch interface {
	List(ctx context.Context) ([]model.MetricJSON, error)
	AddBatch(ctx context.Context, arr []model.MetricJSON) error
}

type srvGetter interface {
	Get(ctx context.Context, info model.Info) (model.MetricJSON, error)
}

type srvPing interface {
	Ping() error
}

// Обновление метрики. [POST-JSON].
// Чтение Body, запись в ResponseWriter ответа от service.
func PostJSONUpdateHandle(srv srvUpdater, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		met, err := metricFromBoby(req.Body)
		if err != nil {
			log.Error("postJsonUpdHandler", "parseBody", err)
			http.Error(rw, err.Error(), http.StatusBadRequest)

			return
		}

		metDB, err := srv.Update(req.Context(), met)
		if err != nil {
			log.Error("postJsonUpdHandler", "update", err)
			http.Error(rw, err.Error(), http.StatusBadRequest)

			return
		}

		rw.Header().Set("Content-Type", ApplicationJSONConst)

		if err := json.NewEncoder(rw).Encode(metDB); err != nil {
			log.Error("postJsonUpdHandler", "encode", err)
			http.Error(rw, err.Error(), http.StatusBadRequest)
		}
	})
}

// Обновление метрики. [POST].
// Чтение Request, запись в ResponseWriter ответа от service.
func PostUpdateHandle(srv srvUpdater, log *slog.Logger, fn func(req *http.Request) model.MetricStr) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		metStr := fn(req)

		met, err := parseMetricJSON(metStr)
		if err != nil {
			log.Error("postUpdateHandler", "parse", err)
			http.Error(rw, err.Error(), http.StatusBadRequest)

			return
		}

		metDB, err := srv.Update(req.Context(), met)
		if err != nil {
			log.Error("postUpdateHandler", "update error", err)
			http.Error(rw, err.Error(), http.StatusBadRequest)

			return
		}

		rw.Header().Set("Content-Type", ApplicationJSONConst)

		if err := json.NewEncoder(rw).Encode(metDB); err != nil {
			log.Error("postUpdateHandler", "encode error", err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	})
}

// Получние метрики. [GET].
// Чтение Request, запись в ResponseWriter ответа от service.
func GetValueHandle(srv srvGetter, log *slog.Logger, fn func(*http.Request) model.InfoStr) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		infoStr := fn(req)

		mInfo, err := model.ParseInfo(infoStr.Name, infoStr.MType)
		if err != nil {
			log.Error("getValueHandler", "error", err)
			http.Error(rw, err.Error(), http.StatusBadRequest)

			return
		}

		met, err := srv.Get(req.Context(), mInfo)
		if err != nil {
			log.Error("getValueHandler", "srvGet error", err)
			http.Error(rw, err.Error(), http.StatusNotFound)

			return
		}

		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")

		if _, err = rw.Write([]byte(met.String())); err != nil {
			log.Error("getValueHandler", "write data error", err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	})
}

// Получние метрики. [POST].
// Чтение Body, запись в ResponseWriter ответа от service.
func PostValueHandle(srv srvGetter, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		met, err := metricFromBoby(req.Body)
		if err != nil {
			log.Error("postValueHandler", "parse body error", err)
			http.Error(rw, err.Error(), http.StatusBadRequest)

			return
		}

		mInfo, err := model.ParseInfo(met.ID, met.MType)
		if err != nil {
			log.Error("postValueHandler", "error", err)
			http.Error(rw, err.Error(), http.StatusBadRequest)

			return
		}

		rw.Header().Set("Content-Type", ApplicationJSONConst)

		metDB, err := srv.Get(req.Context(), mInfo)
		if err != nil {
			log.Error("postValueHandler", "srvGet error", err, "mInfo", mInfo)
			http.Error(rw, err.Error(), http.StatusNotFound)

			return
		}

		if err := json.NewEncoder(rw).Encode(metDB); err != nil {
			log.Error("postValueHandler", "encode error", err)
			http.Error(rw, err.Error(), http.StatusBadRequest)
		}
	})
}

// Обновление списка метрик. [POST].
// Чтение Body, запись в ResponseWriter ответа от service.
func PostUpdatesHandler(srv srvBatch, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var list []model.MetricJSON

		body := req.Body
		defer body.Close()

		if err := json.NewDecoder(body).Decode(&list); err != nil {
			log.Error("postUpdatesHandler", "encode error", err)
			http.Error(rw, err.Error(), http.StatusBadRequest)

			return
		}

		if err := srv.AddBatch(req.Context(), list); err != nil {
			log.Error("postUpdatesHandler", "srvAddBatch error", err)

			http.Error(rw, err.Error(), http.StatusBadRequest)

			return
		}
	})
}

// Получение списка метрик. [GET].
// Запись в ResponseWriter ответа от service.
func ListHandle(srv srvBatch, tmpl *template.Template, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		list, err := srv.List(req.Context())
		if err != nil {
			log.Error("listHandler", "srvList error", err)
			http.Error(rw, err.Error(), http.StatusBadRequest)

			return
		}

		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, "List", list); err != nil {
			log.Error("listHandler", "exec tmpl error", err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)

			return
		}

		rw.Header().Set("Content-Type", TextHTMLConst)

		if _, err := buf.WriteTo(rw); err != nil {
			log.Error("listHandler", "write data error", err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	})
}

// Вызов service.Ping. [GET].
func PingHandler(srv srvPing, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		if err := srv.Ping(); err != nil {
			log.Error("pingHandler", "srvPing error", err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)

			return
		}

		rw.WriteHeader(http.StatusOK)
	})
}

// Чтение метрики из Body.
func metricFromBoby(body io.ReadCloser) (model.MetricJSON, error) {
	defer body.Close()

	var metricJSON model.MetricJSON

	if err := json.NewDecoder(body).Decode(&metricJSON); err != nil {
		return model.MetricJSON{}, fmt.Errorf("%w", err)
	}

	return metricJSON, nil
}

// parseMetricJSON возвращает model.MetricJSON из model.MetricStr.
// Возможные ошибки:
// - при конвертации delta[string] > int64, если тип counter,
// - при конвертации value[string] > float64, если тип gauge,
// - model.ErrTypeNotSupport, если тип не поддерживается.
func parseMetricJSON(metStr model.MetricStr) (model.MetricJSON, error) {
	switch metStr.MType {
	case model.TypeCountConst.String():
		val, err := strconv.ParseInt(metStr.Val, 10, 64)
		if err != nil {
			return model.MetricJSON{}, fmt.Errorf("parseInt: %w", err)
		}

		return model.MetricJSON{ID: metStr.Name, MType: metStr.MType, Delta: &val}, nil
	case model.TypeGaugeConst.String():
		val, err := strconv.ParseFloat(metStr.Val, 64)
		if err != nil {
			return model.MetricJSON{}, fmt.Errorf("parseFloat: %w", err)
		}

		return model.MetricJSON{ID: metStr.Name, MType: metStr.MType, Value: &val}, nil
	default:
		return model.MetricJSON{}, model.ErrTypeNotSupport
	}
}
