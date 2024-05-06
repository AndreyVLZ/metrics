package store

import (
	"context"
	_ "net/http/pprof"

	"github.com/AndreyVLZ/metrics/internal/model"
	m "github.com/AndreyVLZ/metrics/internal/model"
	"github.com/AndreyVLZ/metrics/internal/store/filestore"
	"github.com/AndreyVLZ/metrics/internal/store/memstore"
	"github.com/AndreyVLZ/metrics/internal/store/postgres"
)

type Storage interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Ping() error
	UpdateCounter(ctx context.Context, met m.MetricRepo[int64]) (m.MetricRepo[int64], error)
	UpdateGauge(ctx context.Context, met m.MetricRepo[float64]) (m.MetricRepo[float64], error)
	GetCounter(ctx context.Context, name string) (m.MetricRepo[int64], error)
	GetGauge(ctx context.Context, name string) (m.MetricRepo[float64], error)
	List(ctx context.Context) (m.Batch, error)
	AddBatch(ctx context.Context, batch model.Batch) error
}

type StorageType string

const (
	StorageTypePostgres StorageType = "pg"
	StorageTypeInFile   StorageType = "file"
	StorageTypeInMemory StorageType = "mem"
)

type Config struct {
	ConnDB    string
	StorePath string
	IsRestore bool
	StoreInt  int
}

func New(cfg Config) Storage {
	storeType := StorageTypeInMemory

	if cfg.StorePath != "" {
		storeType = StorageTypeInFile
	}

	if cfg.ConnDB != "" {
		storeType = StorageTypePostgres
	}

	switch storeType {
	case StorageTypePostgres:
		return postgres.New(postgres.Config{ConnDB: cfg.ConnDB})
	case StorageTypeInFile:
		return filestore.New(
			filestore.Config{
				StorePath: cfg.StorePath,
				IsRestore: cfg.IsRestore,
				StoreInt:  cfg.StoreInt,
			}, memstore.New())
	default:
		return memstore.New()
	}
}
