package wrapstore

import (
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

func (ws WrapStore) Set(m metric.MetricDB) error {
	fmt.Println("SET BY WrapStore ")
	err := ws.producer.WriteMetric(&m)
	if err != nil {
		log.Printf("ERR %v\n", err)
	}

	return ws.Storage.Set(m)
}
