package memstore

import (
	"context"
	"errors"
	_ "net/http/pprof"

	"github.com/AndreyVLZ/metrics/internal/model"
	"github.com/AndreyVLZ/metrics/internal/store/memstore/cache"
)

const NameConst = "MEM STORE"

var (
	errNotFind = errors.New("not find")
)

type MemStore struct {
	cRepo *memRepo[int64]
	gRepo *memRepo[float64]
}

func New() *MemStore {
	return &MemStore{
		cRepo: newMemRepo[int64](),
		gRepo: newMemRepo[float64](),
	}
}

func (s *MemStore) Name() string                  { return NameConst }
func (s *MemStore) Start(_ context.Context) error { return nil }
func (s *MemStore) Stop(_ context.Context) error  { return nil }
func (s *MemStore) Ping() error                   { return nil }

func (s *MemStore) UpdateCounter(ctx context.Context, met model.MetricRepo[int64]) (model.MetricRepo[int64], error) {
	return s.cRepo.update(ctx, met), nil
}

func (s *MemStore) UpdateGauge(ctx context.Context, met model.MetricRepo[float64]) (model.MetricRepo[float64], error) {
	return s.gRepo.update(ctx, met), nil
}

func (s *MemStore) GetCounter(ctx context.Context, name string) (model.MetricRepo[int64], error) {
	return s.cRepo.get(ctx, name)
}

func (s *MemStore) GetGauge(ctx context.Context, name string) (model.MetricRepo[float64], error) {
	return s.gRepo.get(ctx, name)
}

func (s *MemStore) List(ctx context.Context) (model.Batch, error) {
	return model.Batch{
		CList: s.cRepo.list(ctx),
		GList: s.gRepo.list(ctx),
	}, nil
}

func (s *MemStore) AddBatch(ctx context.Context, batch model.Batch) error {
	s.cRepo.addList(ctx, batch.CList)
	s.gRepo.addList(ctx, batch.GList)

	return nil
}

type memRepo[VT model.ValueType] struct {
	cache *cache.Cache[VT]
}

func newMemRepo[VT model.ValueType]() *memRepo[VT] {
	return &memRepo[VT]{
		cache: cache.New[VT](),
	}
}

func (r memRepo[VT]) get(_ context.Context, name string) (model.MetricRepo[VT], error) {
	met, ok := r.cache.Get(name)
	if !ok {
		return model.MetricRepo[VT]{}, errNotFind
	}

	return met, nil
}

func (r *memRepo[VT]) update(_ context.Context, met model.MetricRepo[VT]) model.MetricRepo[VT] {
	if metCache, ok := r.cache.Get(met.Name()); ok {
		metCache.Update(met.Value())

		return r.cache.Set(metCache)
	}

	return r.cache.Set(met)
}

func (r memRepo[VT]) addList(_ context.Context, arr []model.MetricRepo[VT]) {
	for i := range arr {
		met := arr[i]

		metCache, ok := r.cache.Get(met.Name())
		if ok {
			metCache.Update(met.Value())
			r.cache.Set(metCache)

			continue
		}

		r.cache.Set(met)
	}
}

func (r memRepo[VT]) list(_ context.Context) []model.MetricRepo[VT] { return r.cache.List() }
