package server

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/AndreyVLZ/metrics/internal/model"
	"github.com/AndreyVLZ/metrics/internal/store"
	api "github.com/AndreyVLZ/metrics/server/http"
	"github.com/AndreyVLZ/metrics/server/http/handler"
	m "github.com/AndreyVLZ/metrics/server/http/middleware"
	"github.com/AndreyVLZ/metrics/server/service"
	"github.com/go-chi/chi/v5"
)

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
		<li><strong>{{ .ID }}</strong>[{{.MType}}]: {{.Delta}} :: {{.Value}}</li>
	{{ end }}
	</ol>
</body>
</html>{{end}}`

const stopTimeout = 5

type iAPI interface {
	Start() error
	Stop(ctx context.Context) error
}

type IService interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type Server struct {
	cfg      *Config
	api      iAPI
	log      *slog.Logger
	services []IService
}

func New(cfg *Config, log *slog.Logger) Server {
	store := store.New(
		store.Config{
			ConnDB:    cfg.dbDNS,
			StorePath: cfg.storePath,
			IsRestore: cfg.isRestore,
			StoreInt:  cfg.storeInt,
		},
	)

	srv := service.New(store)
	mux := initChiRouter(srv, log)

	handler := m.Logging(log,
		m.Gzip(
			m.Hash(
				cfg.key,
				mux,
			),
		),
	)

	httpServer := api.NewServer(
		api.Config{
			Addr: cfg.addr,
		},
		handler,
	)

	return Server{
		cfg:      cfg,
		api:      httpServer,
		services: []IService{store},
		log:      log,
	}
}

func (srv *Server) Start(ctx context.Context) error {
	srv.log.LogAttrs(ctx,
		slog.LevelInfo, "start server",
		slog.String("addr", srv.cfg.addr),
		slog.Group("flags",
			slog.Int("storeInterval", srv.cfg.storeInt),
			slog.String("storePath", srv.cfg.storePath),
			slog.Bool("restore", srv.cfg.isRestore),
			slog.String("dbDNS", srv.cfg.dbDNS),
			slog.String("key", srv.cfg.key),
		),
	)

	for i := range srv.services {
		if err := srv.services[i].Start(ctx); err != nil {
			return fmt.Errorf("%w", err)
		}

		srv.log.Info("services Start", "name", srv.services[i].Name())
	}

	return srv.api.Start()
}

func (srv *Server) Stop(ctx context.Context) error {
	ctxTimeout, stopTimeout := context.WithTimeout(ctx, stopTimeout*time.Second)
	defer stopTimeout()

	errs := make([]error, 0, len(srv.services)+1)
	if err := srv.api.Stop(ctxTimeout); err != nil {
		errs = append(errs, err)
	}

	for _, srv := range srv.services {
		if err := srv.Stop(ctxTimeout); err != nil {
			errs = append(errs, fmt.Errorf("service [%s] err: %w", srv.Name(), err))
		}
	}

	return errors.Join(errs...)
}

func initChiRouter(srv service.Service, log *slog.Logger) *chi.Mux {
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
		r.Get("/debug/pprof/", http.HandlerFunc(pprof.Index))
		r.Get("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		r.Get("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		r.Get("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		r.Get("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	})

	return route
}

func metFromReq(args [3]string) func(req *http.Request) model.MetricStr {
	return func(req *http.Request) model.MetricStr {
		return model.MetricStr{
			Info: metInfoFromReq([2]string{args[0], args[1]})(req),
			Val:  chi.URLParam(req, args[2]),
		}
	}
}

func metInfoFromReq(args [2]string) func(req *http.Request) model.Info {
	return func(req *http.Request) model.Info {
		return model.Info{
			MType: chi.URLParam(req, args[0]),
			Name:  chi.URLParam(req, args[1]),
		}
	}
}
