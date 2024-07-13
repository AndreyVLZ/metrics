// Формирует данные для запроса к store и для последующего ответа handler'у.
package service

import (
	"context"
	"fmt"

	"github.com/AndreyVLZ/metrics/internal/model"
)

// Интерфейс хранилища.
type store interface {
	Ping() error
	Get(ctx context.Context, mInfo model.Info) (model.Metric, error)
	Update(ctx context.Context, met model.Metric) (model.Metric, error)
	List(ctx context.Context) ([]model.Metric, error)
	AddBatch(ctx context.Context, arr []model.Metric) error
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

// Ping.
func (srv Service) Ping() error { return srv.store.Ping() }

// Добавление списка метрик.
func (srv Service) AddBatch(ctx context.Context, list []model.MetricJSON) error {
	arr, err := buildArrMetric(list)
	if err != nil {
		return fmt.Errorf("buildArrMetric: %w", err)
	}

	if err := srv.store.AddBatch(ctx, arr); err != nil {
		return fmt.Errorf("store.AddBatch: %w", err)
	}

	return nil
}

// Список метрик.
func (srv Service) List(ctx context.Context) ([]model.MetricJSON, error) {
	list, err := srv.store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("store.List: %w", err)
	}

	return model.BuildArrMetricJSON(list), nil
}

// Обновление метрики.
func (srv Service) Update(ctx context.Context, metJSON model.MetricJSON) (model.MetricJSON, error) {
	met, err := parseMetric(metJSON)
	if err != nil {
		return model.MetricJSON{}, fmt.Errorf("parseMetric: %w", err)
	}

	metDB, err := srv.store.Update(ctx, met)
	if err != nil {
		return model.MetricJSON{}, fmt.Errorf("store.Update: %w", err)
	}

	return model.BuildMetricJSON(metDB), nil
}

// Получение метрики.
func (srv Service) Get(ctx context.Context, metInfo model.Info) (model.MetricJSON, error) {
	metDB, err := srv.store.Get(ctx, metInfo)
	if err != nil {
		return model.MetricJSON{}, fmt.Errorf("store.Get: %w", err)
	}

	return model.BuildMetricJSON(metDB), nil
}

// Возвращает массив MetricJSON из массива Metric.
func buildArrMetric(arr []model.MetricJSON) ([]model.Metric, error) {
	res := make([]model.Metric, len(arr))

	for i := range arr {
		met, err := parseMetric(arr[i])
		if err != nil {
			return nil, fmt.Errorf("parseMetric: %w", err)
		}

		res[i] = met
	}

	return res, nil
}

func parseMetric(met model.MetricJSON) (model.Metric, error) {
	var val model.Value

	info, err := model.ParseInfo(met.ID, met.MType)
	if err != nil {
		return model.Metric{}, fmt.Errorf("parseInfo: %w", err)
	}

	switch info.MType {
	case model.TypeCountConst:
		val = model.Value{Delta: met.Delta, Val: nil}
	case model.TypeGaugeConst:
		val = model.Value{Delta: nil, Val: met.Value}
	default:
		return model.Metric{}, model.ErrTypeNotSupport
	}

	return model.Metric{Info: info, Value: val}, nil
}
