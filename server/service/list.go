package service

import (
	"context"
	"fmt"

	_ "net/http/pprof"

	"github.com/AndreyVLZ/metrics/internal/model"
)

func (srv Service) AddBatch(ctx context.Context, list []model.MetricJSON) error {
	fmt.Printf("\nsBatch::%v\n\n", list)
	cList := make([]model.MetricRepo[int64], 0, len(list))
	gList := make([]model.MetricRepo[float64], 0, len(list))

	for i := range list {
		switch list[i].MType {
		case model.TypeCountConst.String():
			cList = append(cList, model.NewMetricRepo(list[i].ID, model.TypeCountConst, *list[i].Delta))
		case model.TypeGaugeConst.String():
			gList = append(gList, model.NewMetricRepo(list[i].ID, model.TypeGaugeConst, *list[i].Value))
		}
	}

	return srv.store.AddBatch(ctx, model.Batch{CList: cList, GList: gList})
}

func (srv Service) List(ctx context.Context) ([]model.MetricJSON, error) {
	batch, err := srv.store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	arr := make([]model.MetricJSON, 0, len(batch.CList)+len(batch.GList))

	list(&arr, batch.CList, func(m *model.MetricJSON, val int64) { m.Delta = &val })
	list(&arr, batch.GList, func(m *model.MetricJSON, val float64) { m.Value = &val })

	return arr, nil
}

func list[VT model.ValueType](
	arr *[]model.MetricJSON,
	list []model.MetricRepo[VT],
	fnSet func(m *model.MetricJSON, val VT),
) {
	for i := range list {
		metJSON := model.MetricJSON{ID: list[i].Name(), MType: list[i].Type()}
		fnSet(&metJSON, list[i].Value())
		*arr = append(*arr, metJSON)
	}
}
