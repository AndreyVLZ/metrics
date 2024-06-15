// Определяет для интерфейса storage
// Дополнительный метод 'заглушку' Ping.
package adapter

import (
	"context"

	"github.com/AndreyVLZ/metrics/internal/model"
)

// storage интерфейс хранилища.
type storage interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Get(ctx context.Context, mInfo model.Info) (model.Metric, error)
	Update(ctx context.Context, met model.Metric) (model.Metric, error)
	List(ctx context.Context) ([]model.Metric, error)
	AddBatch(ctx context.Context, arr []model.Metric) error
}

// PingAdapter хранит интерфейс хранилища.
// Через 'встраивание' storage определяем для
// PingAdapter все методы интeрфейса.
type PingAdapter struct {
	storage
}

// Ping возвращает PingAdapter для хранилища store.
func Ping(store storage) PingAdapter {
	return PingAdapter{
		storage: store,
	}
}

// Ping возращает нулевую ошибку.
func (pa PingAdapter) Ping() error { return nil }
