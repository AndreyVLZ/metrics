package wrapstore

import (
	"context"
	"fmt"
	"log"

	"github.com/AndreyVLZ/metrics/cmd/server/producer"
	"github.com/AndreyVLZ/metrics/internal/metric"
	"github.com/AndreyVLZ/metrics/internal/storage"
)

type WrapStore struct {
	storage.Storage
	producer *producer.Producer
}

func NewWrapStore(store storage.Storage, produce *producer.Producer) WrapStore {
	return WrapStore{
		Storage:  store,
		producer: produce,
	}
}

func (ws WrapStore) Set(ctx context.Context, m metric.MetricDB) (metric.MetricDB, error) {
	err := ws.producer.WriteMetric(&m)
	fmt.Printf("Wrap write %v\n", m)
	if err != nil {
		log.Printf("err set wrap store %v\n", err)
	}

	return ws.Storage.Set(ctx, m)
}
