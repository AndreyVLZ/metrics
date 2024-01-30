package storage

import (
	"context"
	"errors"

	"github.com/AndreyVLZ/metrics/internal/metric"
	"github.com/AndreyVLZ/metrics/internal/storage/memstorage"
	"github.com/AndreyVLZ/metrics/internal/storage/postgres"
)

var ErrTypeStorageNotSupport = errors.New("type store not support")

type Storage interface {
	Open() error
	Ping() error
	List(context.Context) []metric.MetricDB
	Get(context.Context, metric.MetricDB) (metric.MetricDB, error)
	// добавляем если нет
	Set(context.Context, metric.MetricDB) (metric.MetricDB, error)
	SetBatch(context.Context, []metric.MetricDB) error
	// добавляем/обновляем если нет
	Update(context.Context, metric.MetricDB) (metric.MetricDB, error)
	UpdateBatch(context.Context, []metric.MetricDB) error
}

type StorageType string

const (
	StorageTypePostgres StorageType = "pg"
	StorageTypeInmemory StorageType = "mem"
)

type Config struct {
	StorageType StorageType
	ConnDB      string
}

func New(cfg Config) (Storage, error) {
	switch cfg.StorageType {
	case StorageTypeInmemory:
		return memstorage.New(), nil
	case StorageTypePostgres:
		return postgres.New(postgres.PostgresConfig{ConnDB: cfg.ConnDB}), nil
	default:
		return nil, ErrTypeStorageNotSupport
	}
}
