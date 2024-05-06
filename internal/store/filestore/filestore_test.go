package filestore

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/AndreyVLZ/metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

type fakeStore struct {
	name     string
	err      error
	metCount model.MetricRepo[int64]
	metFloat model.MetricRepo[float64]
	batch    model.Batch
}

func (fs *fakeStore) Name() string                  { return fs.name }
func (fs *fakeStore) Ping() error                   { return nil }
func (fs *fakeStore) Start(_ context.Context) error { return fs.err }
func (fs *fakeStore) Stop(_ context.Context) error  { return fs.err }

func (fs *fakeStore) UpdateCounter(_ context.Context, met model.MetricRepo[int64]) (model.MetricRepo[int64], error) {
	fs.metCount = met
	return fs.metCount, fs.err
}

func (fs *fakeStore) UpdateGauge(_ context.Context, met model.MetricRepo[float64]) (model.MetricRepo[float64], error) {
	return fs.metFloat, fs.err
}

func (fs *fakeStore) GetCounter(_ context.Context, name string) (model.MetricRepo[int64], error) {
	return fs.metCount, fs.err
}

func (fs *fakeStore) GetGauge(_ context.Context, name string) (model.MetricRepo[float64], error) {
	return fs.metFloat, fs.err
}

func (fs *fakeStore) AddBatch(_ context.Context, batch model.Batch) error { return fs.err }

func (fs *fakeStore) List(_ context.Context) (model.Batch, error) {
	return fs.batch, fs.err
}

func TestName(t *testing.T) {
	store := New(Config{}, &fakeStore{})
	assert.Equal(t, NameConst, store.Name())
}

func TestPing(t *testing.T) {
	store := New(Config{}, &fakeStore{})
	assert.Equal(t, nil, store.Ping())
}

func TestUpdateCounter(t *testing.T) {
	type testCase struct {
		name      string
		cfg       Config
		fakeStore *fakeStore
		metSet    model.MetricRepo[int64]
		isErr     bool
		wantMet   model.MetricRepo[int64]
		wantBatch model.Batch
	}

	tc := []testCase{
		{
			name: "ok countter",
			cfg: Config{
				StorePath: "tempFileStore.json",
				IsRestore: false,
				StoreInt:  0, // 0-wrap
			},
			fakeStore: &fakeStore{
				metCount: model.NewMetricRepo[int64](
					"Counter-1",
					model.TypeCountConst,
					10,
				),
			},
			metSet: model.NewMetricRepo[int64](
				"Counter-1",
				model.TypeCountConst,
				10,
			),
			wantMet: model.NewMetricRepo[int64](
				"Counter-1",
				model.TypeCountConst,
				10,
			),
			wantBatch: model.Batch{
				CList: []model.MetricRepo[int64]{
					model.NewMetricRepo[int64](
						"Counter-1",
						model.TypeCountConst,
						10,
					),
				},
				GList: []model.MetricRepo[float64]{},
			},
			isErr: false,
		},

		{
			name: "counter err store",
			cfg: Config{
				StorePath: "tempFileStore.json",
				IsRestore: false,
				StoreInt:  0, // 0-wrap
			},
			fakeStore: &fakeStore{
				err: errors.New("store err"),
			},
			metSet: model.NewMetricRepo[int64](
				"Counter-1",
				model.TypeCountConst,
				10,
			),
			wantMet: model.MetricRepo[int64]{},
			wantBatch: model.Batch{
				CList: []model.MetricRepo[int64]{},
				GList: []model.MetricRepo[float64]{},
			},
			isErr: true,
		},
	}

	ctx := context.Background()
	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {

			store := New(test.cfg, test.fakeStore)
			if err := store.Start(ctx); err != nil {
				t.Errorf("start store: %v\n", err)
			}

			defer func() {
				store.Stop(ctx)
				if err := os.Remove(test.cfg.StorePath); err != nil {
					t.Errorf("remove temp file: %v\n", err)
				}
			}()

			metDB, err := store.UpdateCounter(ctx, test.metSet)
			if test.isErr && err == nil {
				t.Error("want err")
			}

			assert.Equal(t, test.wantMet, metDB)

			batch, err := store.file.ReadBatch()
			if err != nil {
				t.Errorf("read batch: %v\n", err)
			}

			assert.Equal(t, test.wantBatch, batch)
		})
	}
}

func TestUpdateGauge(t *testing.T) {
	type testCase struct {
		name      string
		cfg       Config
		fakeStore *fakeStore
		metSet    model.MetricRepo[float64]
		isErr     bool
		wantMet   model.MetricRepo[float64]
		wantBatch model.Batch
	}

	tc := []testCase{
		{
			name: "ok gauge",
			cfg: Config{
				StorePath: "tempFileStore.json",
				IsRestore: false,
				StoreInt:  0, // 0-wrap
			},
			fakeStore: &fakeStore{
				metFloat: model.NewMetricRepo[float64](
					"Gauge-1",
					model.TypeGaugeConst,
					10.01,
				),
			},
			metSet: model.NewMetricRepo[float64](
				"Gauge-1",
				model.TypeGaugeConst,
				10.01,
			),
			wantMet: model.NewMetricRepo[float64](
				"Gauge-1",
				model.TypeGaugeConst,
				10.01,
			),
			isErr: false,
			wantBatch: model.Batch{
				CList: []model.MetricRepo[int64]{},
				GList: []model.MetricRepo[float64]{
					model.NewMetricRepo[float64](
						"Gauge-1",
						model.TypeGaugeConst,
						10.01,
					),
				},
			},
		},

		{
			name: "gauge store err",
			cfg: Config{
				StorePath: "tempFileStore.json",
				IsRestore: false,
				StoreInt:  0, // 0-wrap
			},
			fakeStore: &fakeStore{
				err: errors.New("store err"),
			},
			metSet: model.NewMetricRepo[float64](
				"Gauge-1",
				model.TypeGaugeConst,
				10.01,
			),
			wantMet:   model.MetricRepo[float64]{},
			wantBatch: model.Batch{CList: []model.MetricRepo[int64]{}, GList: []model.MetricRepo[float64]{}},
			isErr:     true,
		},
	}

	ctx := context.Background()

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {

			store := New(test.cfg, test.fakeStore)
			if err := store.Start(ctx); err != nil {
				t.Errorf("start store: %v\n", err)
			}

			defer func() {
				store.Stop(ctx)
				if err := os.Remove(test.cfg.StorePath); err != nil {
					t.Errorf("remove temp file: %v\n", err)
				}
			}()

			metDB, err := store.UpdateGauge(ctx, test.metSet)
			if test.isErr && err == nil {
				t.Error("want err")
			}

			assert.Equal(t, test.wantMet, metDB)

			batch, err := store.file.ReadBatch()
			if err != nil {
				t.Errorf("read batch: %v\n", err)
			}

			assert.Equal(t, test.wantBatch, batch)
		})
	}
}

func TestAddBatch(t *testing.T) {
	type testCase struct {
		name      string
		cfg       Config
		fakeStore *fakeStore
		batch     model.Batch
		wantBatch model.Batch
		isErr     bool
	}

	tc := []testCase{
		{
			name: "ok",
			cfg: Config{
				StorePath: "tempFileStore.json",
				IsRestore: false,
				StoreInt:  0, // 0-wrap
			},
			fakeStore: &fakeStore{},
			batch: model.Batch{
				CList: []model.MetricRepo[int64]{
					model.NewMetricRepo[int64](
						"Counter-1",
						model.TypeCountConst,
						1,
					),
				},
			},
			wantBatch: model.Batch{
				CList: []model.MetricRepo[int64]{
					model.NewMetricRepo[int64](
						"Counter-1",
						model.TypeCountConst,
						1,
					),
				},
				GList: []model.MetricRepo[float64]{},
			},
			isErr: false,
		},

		{
			name: "store addBatch err",
			cfg: Config{
				StorePath: "tempFileStore.json",
				IsRestore: false,
				StoreInt:  0, // 0-wrap
			},
			fakeStore: &fakeStore{err: errors.New("errStore addBatch")},
			isErr:     true,
			batch: model.Batch{
				CList: []model.MetricRepo[int64]{
					model.NewMetricRepo[int64](
						"Counter-1",
						model.TypeCountConst,
						1,
					),
				},
				GList: []model.MetricRepo[float64]{},
			},
			wantBatch: model.Batch{
				CList: []model.MetricRepo[int64]{},
				GList: []model.MetricRepo[float64]{},
			},
		},
	}

	ctx := context.Background()
	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			store := New(test.cfg, test.fakeStore)
			if err := store.Start(ctx); err != nil {
				t.Errorf("start store: %v\n", err)
			}

			defer func() {
				store.Stop(ctx)
				if err := os.Remove(test.cfg.StorePath); err != nil {
					t.Errorf("remove temp file: %v\n", err)
				}
			}()

			err := store.AddBatch(ctx, test.batch)
			if test.isErr && err == nil {
				t.Error("want err")
			}
			batch, err := store.file.ReadBatch()
			if err != nil {
				t.Errorf("read batch: %v\n", err)
			}

			assert.Equal(t, test.wantBatch, batch)
		})
	}
}
