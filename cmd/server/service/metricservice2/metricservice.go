package metricservice

import "context"

type Storage interface {
	Get(context.Context) error
	Set(context.Context) error
	List(context.Context) error
}

type metricService struct {
	storage Storage
}

func New(storage Storage) *metricService {
	return &metricService{storage: storage}
}

func (ms *metricService) Get(ctx context.Context) error {
	return ms.storage.Get(ctx)
}

// UpSet
func (ms *metricService) Set(ctx context.Context) error {
	return ms.storage.Set(ctx)
}

func (ms *metricService) List(ctx context.Context) error {
	return ms.storage.Set(ctx)
}
