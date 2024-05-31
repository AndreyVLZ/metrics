package filestore

import (
	"context"
	"fmt"
	"log"

	"github.com/AndreyVLZ/metrics/internal/model"
)

type wrapStore struct {
	file iFile
	storage
}

func newWrapStore(file iFile, s storage) *wrapStore {
	return &wrapStore{
		file:    file,
		storage: s,
	}
}

func (ws *wrapStore) Update(ctx context.Context, met model.Metric) (model.Metric, error) {
	metDB, err := ws.storage.Update(ctx, met)
	if err != nil {
		return model.Metric{}, fmt.Errorf("%w", err)
	}

	if err := ws.file.WriteMetric(metDB); err != nil {
		log.Printf("err write metric in file: %v\n", err)
	}

	return metDB, nil
}

func (ws *wrapStore) AddBatch(ctx context.Context, arr []model.Metric) error {
	if err := ws.storage.AddBatch(ctx, arr); err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := ws.file.WriteBatch(arr); err != nil {
		log.Printf("err writeBatch in file: %v\n", err)
	}

	return nil
}
