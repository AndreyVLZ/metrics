package service

import (
	"context"
	_ "net/http/pprof"

	"github.com/AndreyVLZ/metrics/internal/model"
)

func (srv Service) Update(ctx context.Context, metJSON model.MetricJSON) (model.MetricJSON, error) {
	switch metJSON.MType {
	case model.TypeCountConst.String():
		return upd(ctx,
			model.NewMetricRepo(metJSON.ID, model.TypeCountConst, *metJSON.Delta),
			srv.store.UpdateCounter,
			func(m *model.MetricJSON, val int64) { m.Delta = &val },
		)
	case model.TypeGaugeConst.String():
		return upd(ctx,
			model.NewMetricRepo(metJSON.ID, model.TypeGaugeConst, *metJSON.Value),
			srv.store.UpdateGauge,
			func(m *model.MetricJSON, val float64) { m.Value = &val },
		)
	default:
		return model.MetricJSON{}, errTypeNotSupport
	}
}

func upd[VT model.ValueType](ctx context.Context,
	met model.MetricRepo[VT],
	fnUpd func(context.Context, model.MetricRepo[VT]) (model.MetricRepo[VT], error),
	fnParse func(m *model.MetricJSON, val VT),
) (model.MetricJSON, error) {
	m1, err := fnUpd(ctx, met)
	if err != nil {
		return model.MetricJSON{}, err
	}

	mJSON := model.MetricJSON{ID: m1.Name(), MType: m1.Type()}
	fnParse(&mJSON, m1.Value())

	return mJSON, nil
}
