package restoreservice

import (
	"context"

	"github.com/AndreyVLZ/metrics/cmd/server/consumer"
	"github.com/AndreyVLZ/metrics/internal/storage"
)

type restoreService struct {
	store    storage.Storage
	consumer *consumer.Consumer
}

func New(store storage.Storage, fileName string) *restoreService {
	return &restoreService{
		store:    store,
		consumer: consumer.New(fileName),
	}
}

func (rs *restoreService) Name() string { return "restoreService" }
func (rs *restoreService) Start() error {
	err := rs.consumer.Open()
	if err != nil {
		return err
	}
	arr, err := rs.consumer.ReadMetric()
	if err != nil {
		return err
	}
	return rs.store.SetBatch(context.Background(), arr)
}

func (rs *restoreService) Stop() error {
	return rs.consumer.Close()
}
