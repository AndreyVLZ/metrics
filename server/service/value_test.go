package service

import (
	"context"
	"errors"
	"testing"

	"github.com/AndreyVLZ/metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func (fs *fakeStore) GetCounter(_ context.Context, name string) (model.MetricRepo[int64], error) {
	if fs.err != nil {
		return model.MetricRepo[int64]{}, fs.err
	}

	return fs.mCRepo, nil
}

func (fs *fakeStore) GetGauge(_ context.Context, name string) (model.MetricRepo[float64], error) {
	if fs.err != nil {
		return model.MetricRepo[float64]{}, fs.err
	}

	return fs.mGRepo, nil
}

func TestGet(t *testing.T) {
	type testCase struct {
		name   string
		isErr  bool
		fStore *fakeStore
		fnWant func() model.MetricJSON
		mInfo  model.Info
	}

	tc := []testCase{
		{
			name:  "counter ok",
			isErr: false,
			mInfo: model.Info{
				Name:  "Counter-1",
				MType: model.TypeCountConst.String(),
			},
			fStore: &fakeStore{
				mCRepo: model.NewMetricRepo[int64](
					"Counter-1",
					model.TypeCountConst,
					10,
				),
			},
			fnWant: func() model.MetricJSON {
				val := int64(10)
				return model.MetricJSON{
					ID:    "Counter-1",
					MType: "counter",
					Delta: &val,
				}
			},
		},

		{
			name:  "gauge ok",
			isErr: false,
			mInfo: model.Info{
				Name:  "Gauge-1",
				MType: model.TypeGaugeConst.String(),
			},
			fStore: &fakeStore{
				mGRepo: model.NewMetricRepo[float64](
					"Gauge-1",
					model.TypeGaugeConst,
					10.01,
				),
			},
			fnWant: func() model.MetricJSON {
				val := float64(10.01)
				return model.MetricJSON{
					ID:    "Gauge-1",
					MType: "gauge",
					Value: &val,
				}
			},
		},

		{
			name:  "err srv",
			isErr: true,
			mInfo: model.Info{
				Name:  "Gauge-1",
				MType: model.TypeGaugeConst.String(),
			},
			fStore: &fakeStore{
				err: errors.New("err server"),
			},
			fnWant: func() model.MetricJSON {
				return model.MetricJSON{}
			},
		},
		{
			name:   "err type not support",
			isErr:  true,
			mInfo:  model.Info{},
			fStore: &fakeStore{},
			fnWant: func() model.MetricJSON {
				return model.MetricJSON{}
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			srv := New(test.fStore)

			actMetJSON, err := srv.Get(ctx, test.mInfo)
			if test.isErr && err == nil {
				t.Fatalf("wanted err:%v", err)
			}

			assert.Equal(t, test.fnWant(), actMetJSON)
		})
	}
}
