package service

import (
	"context"
	"errors"
	_ "net/http/pprof"

	m "github.com/AndreyVLZ/metrics/internal/model"
)

type Store interface {
	UpdateCounter(ctx context.Context, met m.MetricRepo[int64]) (m.MetricRepo[int64], error)
	UpdateGauge(ctx context.Context, met m.MetricRepo[float64]) (m.MetricRepo[float64], error)
	GetCounter(ctx context.Context, name string) (m.MetricRepo[int64], error)
	GetGauge(ctx context.Context, name string) (m.MetricRepo[float64], error)
	List(ctx context.Context) (m.Batch, error)
	AddBatch(ctx context.Context, batch m.Batch) error
	Ping() error
}

var errTypeNotSupport = errors.New("type not support")

type Service struct {
	store Store
}

func New(store Store) Service {
	return Service{
		store: store,
	}
}

func (srv Service) Ping() error { return srv.store.Ping() }
