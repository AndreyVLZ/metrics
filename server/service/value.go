package service

import (
	"context"

	_ "net/http/pprof"

	"github.com/AndreyVLZ/metrics/internal/model"
)

func (srv Service) Get(ctx context.Context, metInfo model.Info) (model.MetricJSON, error) {
	switch metInfo.MType {
	case model.TypeCountConst.String():
		return get(ctx, metInfo.Name,
			srv.store.GetCounter,
			func(m *model.MetricJSON, val int64) { m.Delta = &val },
		)
	case model.TypeGaugeConst.String():
		return get(ctx, metInfo.Name,
			srv.store.GetGauge,
			func(m *model.MetricJSON, val float64) { m.Value = &val },
		)
	default:
		return model.MetricJSON{}, errTypeNotSupport
	}
}

func get[VT model.ValueType](
	ctx context.Context,
	name string,
	fnGet func(ctx context.Context, name string) (model.MetricRepo[VT], error),
	fnParse func(m *model.MetricJSON, val VT),
) (model.MetricJSON, error) {
	mDB, err := fnGet(ctx, name)
	if err != nil {
		return model.MetricJSON{}, err
	}

	mJSON := model.MetricJSON{ID: mDB.Name(), MType: mDB.Type(), Delta: nil, Value: nil}
	fnParse(&mJSON, mDB.Value())

	return mJSON, nil
}
