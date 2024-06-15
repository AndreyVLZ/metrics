package inmemory

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/AndreyVLZ/metrics/internal/model"
)

const NameConst = "in memory"

var errNotFind = errors.New("not find")

type Storager interface {
	Get(ctx context.Context, mInfo model.Info) (model.Metric, error)
	Update(ctx context.Context, met model.Metric) (model.Metric, error)
	List(ctx context.Context) ([]model.Metric, error)
	AddBatch(ctx context.Context, arr []model.Metric) error
}

type MemStore struct {
	mu    sync.Mutex
	store map[model.Info]model.Value
}

func New() *MemStore {
	return &MemStore{
		store: make(map[model.Info]model.Value),
	}
}

func (s *MemStore) Start(_ context.Context) error { return nil }
func (s *MemStore) Stop(_ context.Context) error  { return nil }
func (s *MemStore) Name() string                  { return NameConst }

func (s *MemStore) List(_ context.Context) ([]model.Metric, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	arr := make([]model.Metric, 0, len(s.store))
	for mInfo, mVal := range s.store {
		arr = append(arr, model.Metric{Info: mInfo, Value: mVal})
	}

	return arr, nil
}

func (s *MemStore) AddBatch(_ context.Context, arr []model.Metric) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range arr {
		if _, err := s.update(arr[i]); err != nil {
			return err
		}
	}

	return nil
}

func (s *MemStore) Get(_ context.Context, mInfo model.Info) (model.Metric, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	met, ok := s.get(mInfo)
	if !ok {
		return model.Metric{}, errNotFind
	}

	return met, nil
}

func (s *MemStore) Update(_ context.Context, met model.Metric) (model.Metric, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.update(met)
}

func (s *MemStore) get(mInfo model.Info) (model.Metric, bool) {
	val, ok := s.store[mInfo]
	if !ok {
		return model.Metric{}, false
	}

	return model.Metric{Info: mInfo, Value: val}, true
}

func (s *MemStore) update(met model.Metric) (model.Metric, error) {
	mDB, ok := s.get(met.Info)
	if !ok {
		return s.set(met)
	}

	if err := mDB.Update(met.Value); err != nil {
		return model.Metric{}, fmt.Errorf("%w", err)
	}

	return s.set(mDB)
}

func (s *MemStore) set(met model.Metric) (model.Metric, error) {
	s.store[met.Info] = met.Value

	return met, nil
}
