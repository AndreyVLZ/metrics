package adapter

import (
	"context"

	"github.com/AndreyVLZ/metrics/internal/model"
)

type storage interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Get(ctx context.Context, mInfo model.Info) (model.Metric, error)
	Update(ctx context.Context, met model.Metric) (model.Metric, error)
	List(ctx context.Context) ([]model.Metric, error)
	AddBatch(ctx context.Context, arr []model.Metric) error
}

type PingAdapter struct {
	storage
}

func Ping(store storage) PingAdapter {
	return PingAdapter{
		storage: store,
	}
}

func (pa PingAdapter) Ping() error { return nil }
