package service

import (
	"context"
	"errors"
	"testing"

	"github.com/AndreyVLZ/metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func (fs *fakeStore) UpdateCounter(_ context.Context, metRepo model.MetricRepo[int64]) (model.MetricRepo[int64], error) {
	if fs.err != nil {
		return model.MetricRepo[int64]{}, fs.err
	}

	return fs.mCRepo, nil
}

func (fs *fakeStore) UpdateGauge(_ context.Context, metRepo model.MetricRepo[float64]) (model.MetricRepo[float64], error) {
	if fs.err != nil {
		return model.MetricRepo[float64]{}, nil
	}

	return fs.mGRepo, nil
}

func TestUpdate(t *testing.T) {
	type testCase struct {
		name   string
		isErr  bool
		fStore *fakeStore
		fnWant func() model.MetricJSON
		fnSet  func() model.MetricJSON
	}

	tc := []testCase{
		{
			name:  "counter ok",
			isErr: false,
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
			fnSet: func() model.MetricJSON {
				val := int64(1)
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
			fnSet: func() model.MetricJSON {
				val := float64(0.1)
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
			fStore: &fakeStore{
				err: errors.New("err server"),
			},
			fnSet: func() model.MetricJSON {
				val := int64(1)
				return model.MetricJSON{
					ID:    "Counter-1",
					MType: "counter",
					Delta: &val,
				}
			},
			fnWant: func() model.MetricJSON {
				return model.MetricJSON{}
			},
		},
		{
			name:   "err type not support",
			isErr:  true,
			fStore: &fakeStore{},
			fnSet: func() model.MetricJSON {
				val := int64(1)
				return model.MetricJSON{
					ID:    "Counter-1",
					MType: "c",
					Delta: &val,
				}
			},
			fnWant: func() model.MetricJSON {
				return model.MetricJSON{}
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			srv := New(test.fStore)

			actMetJSON, err := srv.Update(ctx, test.fnSet())
			if test.isErr && err == nil {
				t.Fatalf("wanted err:%v", err)
			}

			assert.Equal(t, test.fnWant(), actMetJSON)
		})
	}
}
