package service

import (
	"context"
	"errors"
	"testing"

	"github.com/AndreyVLZ/metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

type fakeStore struct {
	Store
	err    error
	batch  model.Batch
	mCRepo model.MetricRepo[int64]
	mGRepo model.MetricRepo[float64]
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		batch: model.Batch{},
	}
}

func (fs *fakeStore) List(ctx context.Context) (model.Batch, error) {
	if fs.err != nil {
		return model.Batch{}, fs.err
	}

	return fs.batch, nil
}

func (fs *fakeStore) AddBatch(ctx context.Context, batch model.Batch) error {
	if fs.err != nil {
		return fs.err
	}

	fs.batch = batch

	return nil
}

func TestList(t *testing.T) {

	type exec struct {
		isErr bool
		fnExp func() []model.MetricJSON
	}
	type testCase struct {
		name   string
		fStore *fakeStore
		exec   exec
	}

	tc := []testCase{
		{
			name: "ok",
			fStore: &fakeStore{
				batch: exBatch,
			},

			exec: exec{
				isErr: false,
				fnExp: fnList,
			},
		},

		{
			name: "err store",
			fStore: &fakeStore{
				err: errors.New("err store.List"),
			},
			exec: exec{
				isErr: true,
				fnExp: func() []model.MetricJSON {
					return nil
				},
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			srv := New(test.fStore)

			actArr, err := srv.List(ctx)
			if test.exec.isErr && err == nil {
				t.Fatalf("srv list return not err %v\n", err)
			}

			assert.Equal(t, test.exec.fnExp(), actArr)
		})
	}
}

var fnList = func() []model.MetricJSON {
	d1 := int64(1)
	d2 := int64(2)
	v1 := float64(1.1)
	return []model.MetricJSON{
		{
			ID:    "Counter-1",
			MType: model.TypeCountConst.String(),
			Delta: &d1,
		},
		{
			ID:    "Counter-2",
			MType: model.TypeCountConst.String(),
			Delta: &d2,
		},
		{
			ID:    "Gauge-1",
			MType: model.TypeGaugeConst.String(),
			Value: &v1,
		},
	}
}

var exBatch = model.Batch{
	CList: []model.MetricRepo[int64]{
		model.NewMetricRepo[int64](
			"Counter-1",
			model.TypeCountConst,
			1,
		),
		model.NewMetricRepo[int64](
			"Counter-2",
			model.TypeCountConst,
			2,
		),
	},
	GList: []model.MetricRepo[float64]{
		model.NewMetricRepo[float64](
			"Gauge-1",
			model.TypeGaugeConst,
			1.1,
		),
	},
}

func TestAddBatch(t *testing.T) {
	type exec struct {
		isErr bool
		fnExp func() model.Batch
	}

	type testCase struct {
		name   string
		fStore *fakeStore
		list   []model.MetricJSON
		exec   exec
	}

	tc := []testCase{
		{
			name:   "ok",
			fStore: newFakeStore(),
			list:   fnList(),
			exec: exec{
				isErr: false,
				fnExp: func() model.Batch {
					return exBatch
				},
			},
		},

		{
			name:   "err store.AddBatch",
			fStore: &fakeStore{batch: model.Batch{}, err: errors.New("err store.AddBatch")},
			list:   fnList(),
			exec: exec{
				isErr: true,
				fnExp: func() model.Batch {
					return model.Batch{}
				},
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			srv := New(test.fStore)

			err := srv.AddBatch(ctx, test.list)
			if test.exec.isErr && err == nil {
				t.Fatalf("srv list return not err %v\n", err)
			}

			assert.Equal(t, test.exec.fnExp(), test.fStore.batch)
		})
	}
}
