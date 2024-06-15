package service

import (
	"context"
	"errors"
	"testing"

	"github.com/AndreyVLZ/metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

type fakeStore struct {
	err error
	met model.Metric
	arr []model.Metric
}

func (fs *fakeStore) Start(_ context.Context) error {
	return fs.err
}

func (fs *fakeStore) Stop(_ context.Context) error {
	return fs.err
}

func (fs *fakeStore) Get(_ context.Context, mInfo model.Info) (model.Metric, error) {
	return fs.met, fs.err
}

func (fs *fakeStore) Update(_ context.Context, met model.Metric) (model.Metric, error) {
	return fs.met, fs.err
}

func (fs *fakeStore) List(_ context.Context) ([]model.Metric, error) {
	return fs.arr, fs.err
}

func (fs *fakeStore) AddBatch(_ context.Context, arr []model.Metric) error {
	return fs.err
}

func (fs *fakeStore) Ping() error {
	return fs.err
}

func TestPing(t *testing.T) {
	t.Run("ping ok", func(t *testing.T) {
		store := fakeStore{}
		srv := New(&store)
		err := srv.Ping()
		assert.NoError(t, err)
	})
}

func TestAddBatch(t *testing.T) {
	ctx := context.Background()
	t.Run("addBatch ok", func(t *testing.T) {
		var delta int64 = 100
		list := []model.MetricJSON{
			{
				ID:    "PollCount",
				MType: "counter",
				Delta: &delta,
			},
		}
		store := fakeStore{}
		srv := New(&store)
		err := srv.AddBatch(ctx, list)
		assert.NoError(t, err)
	})

	t.Run("addBatch err type no correct", func(t *testing.T) {
		var delta int64 = 100
		list := []model.MetricJSON{
			{
				ID:    "PollCount",
				MType: "counter1",
				Delta: &delta,
			},
		}
		store := fakeStore{}
		srv := New(&store)
		err := srv.AddBatch(ctx, list)
		if err == nil {
			t.Error("want err")
		}
	})

	t.Run("addBatch err name empty", func(t *testing.T) {
		var delta int64 = 100
		list := []model.MetricJSON{
			{
				ID:    "",
				MType: "counter",
				Delta: &delta,
			},
		}
		store := fakeStore{}
		srv := New(&store)
		err := srv.AddBatch(ctx, list)
		if err == nil {
			t.Error("want err")
		}
	})
}

func TestList(t *testing.T) {
	ctx := context.Background()
	t.Run("list ok", func(t *testing.T) {
		var delta int64 = 100
		want := []model.MetricJSON{
			{
				ID:    "Counter-1",
				MType: "counter",
				Delta: &delta,
			},
		}
		store := fakeStore{arr: []model.Metric{
			model.NewCounterMetric("Counter-1", 100),
		}}
		srv := New(&store)
		list, err := srv.List(ctx)
		assert.NoError(t, err)
		assert.Equal(t, want, list)
	})

	t.Run("list err", func(t *testing.T) {
		store := fakeStore{err: errors.New("list err")}
		srv := New(&store)
		_, err := srv.List(ctx)
		if err == nil {
			t.Error("want err")
		}
	})
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	t.Run("update ok", func(t *testing.T) {
		var delta int64 = 100
		met := model.MetricJSON{
			ID:    "PollCount",
			MType: "counter",
			Delta: &delta,
		}

		store := fakeStore{arr: []model.Metric{
			model.NewCounterMetric("Counter-1", 100),
		}}
		srv := New(&store)

		_, err := srv.Update(ctx, met)
		assert.NoError(t, err)
	})

	t.Run("update err no correct type", func(t *testing.T) {
		var delta int64 = 100
		met := model.MetricJSON{
			ID:    "PollCount",
			MType: "1",
			Delta: &delta,
		}

		store := fakeStore{}
		srv := New(&store)

		_, err := srv.Update(ctx, met)
		if err == nil {
			t.Error("want err")
		}
	})

	t.Run("update err no name empty", func(t *testing.T) {
		var delta int64 = 100
		met := model.MetricJSON{
			ID:    "",
			MType: "counter",
			Delta: &delta,
		}

		store := fakeStore{}
		srv := New(&store)

		_, err := srv.Update(ctx, met)
		if err == nil {
			t.Error("want err")
		}
	})

	t.Run("update store update err", func(t *testing.T) {
		var delta int64 = 100
		met := model.MetricJSON{
			ID:    "PollCount",
			MType: "counter",
			Delta: &delta,
		}

		store := fakeStore{err: errors.New("update err")}
		srv := New(&store)

		_, err := srv.Update(ctx, met)
		if err == nil {
			t.Error("want err")
		}
	})
}

func TestGet(t *testing.T) {
	ctx := context.Background()

	t.Run("get ok", func(t *testing.T) {
		minfo := model.Info{
			MName: "Counter-1",
			MType: model.TypeCountConst,
		}
		store := fakeStore{}
		srv := New(&store)
		_, err := srv.Get(ctx, minfo)
		assert.NoError(t, err)
	})

	t.Run("get store err", func(t *testing.T) {
		minfo := model.Info{
			MName: "Counter-1",
			MType: model.TypeCountConst,
		}
		store := fakeStore{err: errors.New("get err")}
		srv := New(&store)
		_, err := srv.Get(ctx, minfo)
		if err == nil {
			t.Error("want err")
		}
	})
}
