// Формирует данные для запроса к store и для последующего ответа handler'у.
package service

import (
	"context"
	"fmt"

	"github.com/AndreyVLZ/metrics/internal/model"
)

// Интерфейс хранилища.
type store interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Get(ctx context.Context, mInfo model.Info) (model.Metric, error)
	Update(ctx context.Context, met model.Metric) (model.Metric, error)
	List(ctx context.Context) ([]model.Metric, error)
	AddBatch(ctx context.Context, arr []model.Metric) error
	Ping() error
}

// Сервис.
type Service struct {
	store store
}

func New(store store) Service {
	return Service{
		store: store,
	}
}

func (srv Service) Ping() error { return srv.store.Ping() }

func (srv Service) AddBatch(ctx context.Context, list []model.MetricJSON) error {
	arr, err := model.BuildArrMetric(list)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return srv.store.AddBatch(ctx, arr)
}

func (srv Service) List(ctx context.Context) ([]model.MetricJSON, error) {
	list, err := srv.store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return model.BuildArrMetricJSON(list), nil
}

func (srv Service) Update(ctx context.Context, metJSON model.MetricJSON) (model.MetricJSON, error) {
	met, err := model.ParseMetric(metJSON)
	if err != nil {
		return model.MetricJSON{}, fmt.Errorf("%w", err)
	}

	metDB, err := srv.store.Update(ctx, met)
	if err != nil {
		return model.MetricJSON{}, fmt.Errorf("%w", err)
	}

	return model.BuildMetricJSON(metDB), nil
}

func (srv Service) Get(ctx context.Context, metInfo model.Info) (model.MetricJSON, error) {
	metDB, err := srv.store.Get(ctx, metInfo)
	if err != nil {
		return model.MetricJSON{}, fmt.Errorf("%w", err)
	}

	return model.BuildMetricJSON(metDB), nil
}
