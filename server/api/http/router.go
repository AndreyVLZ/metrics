package http

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/AndreyVLZ/metrics/internal/model"
	"github.com/AndreyVLZ/metrics/server/api/http/handler"
	m "github.com/AndreyVLZ/metrics/server/api/http/middleware"
	"github.com/go-chi/chi/v5"
)

// Шаблон html.
const tpls = `{{define "List"}}
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>RunTime Metrics</title>
	</head>
<body>
	<ol type="1">
	{{ range . }}
		<li><strong>{{ .ID }}</strong>[{{.MType}}]:{{.String}}</li>
	{{ end }}
	</ol>
</body>
</html>{{end}}`

// Интерфейс service.
type iService interface {
	Ping() error
	Update(ctx context.Context, metJSON model.MetricJSON) (model.MetricJSON, error)
	Get(ctx context.Context, metInfo model.Info) (model.MetricJSON, error)
	List(ctx context.Context) ([]model.MetricJSON, error)
	AddBatch(ctx context.Context, arr []model.MetricJSON) error
}

func NewRoute(srv iService, log *slog.Logger) http.Handler {
	return initChiRouter(srv, log)
}

// Инициализация chi роутера.
func initChiRouter(srv iService, log *slog.Logger) *chi.Mux {
	const (
		typeChiConst  = "typeStr"
		nameChiConst  = "name"
		valueChiConst = "val"
	)

	fnValueParam := metInfoFromReq([2]string{typeChiConst, nameChiConst})
	fnUpdateParam := metFromReq([3]string{typeChiConst, nameChiConst, valueChiConst})

	updateEndPoint := fmt.Sprintf(
		"/{%s}/{%s}/{%s}",
		typeChiConst, nameChiConst, valueChiConst,
	)
	valueEndPoint := fmt.Sprintf(
		"/{%s}/{%s}",
		typeChiConst, nameChiConst,
	)

	tmpl := template.Must(template.New("metrics").Parse(tpls))
	route := chi.NewRouter()

	route.Route("/", func(r chi.Router) {
		r.Get("/", handler.ListHandle(srv, tmpl, log).ServeHTTP)
		r.Get("/ping", handler.PingHandler(srv, log).ServeHTTP)
		r.Post("/updates/",
			m.AppJSON()(handler.PostUpdatesHandler(srv, log)).ServeHTTP,
		)
		r.Route("/update", func(r chi.Router) {
			r.Post("/",
				m.AppJSON()(handler.PostJSONUpdateHandle(srv, log)).ServeHTTP,
			)
			r.Post(updateEndPoint,
				handler.PostUpdateHandle(srv, log, fnUpdateParam).ServeHTTP,
			)
		})
		r.Route("/value", func(r chi.Router) {
			r.Get(valueEndPoint,
				handler.GetValueHandle(srv, log, fnValueParam).ServeHTTP,
			)
			r.Post("/",
				handler.PostValueHandle(srv, log).ServeHTTP,
			)
		})
		/*
			r.Get("/debug/pprof/", http.HandlerFunc(pprof.Index))
			r.Get("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
			r.Get("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
			r.Get("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
			r.Get("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
		*/
	})

	return route
}

// Парсинг url [/counter/Name/Value].
func metFromReq(args [3]string) func(req *http.Request) model.MetricStr {
	return func(req *http.Request) model.MetricStr {
		return model.MetricStr{
			InfoStr: metInfoFromReq([2]string{args[0], args[1]})(req),
			Val:     chi.URLParam(req, args[2]),
		}
	}
}

// Парсинг url [/counter/Name].
func metInfoFromReq(args [2]string) func(req *http.Request) model.InfoStr {
	return func(req *http.Request) model.InfoStr {
		return model.InfoStr{
			MType: chi.URLParam(req, args[0]),
			Name:  chi.URLParam(req, args[1]),
		}
	}
}
