package store

import (
	"context"

	"github.com/AndreyVLZ/metrics/internal/model"
	"github.com/AndreyVLZ/metrics/internal/store/adapter"
	"github.com/AndreyVLZ/metrics/internal/store/filestore"
	"github.com/AndreyVLZ/metrics/internal/store/inmemory"
	"github.com/AndreyVLZ/metrics/internal/store/postgres"
)

type Storage interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Ping() error
	Get(ctx context.Context, mInfo model.Info) (model.Metric, error)
	Update(ctx context.Context, met model.Metric) (model.Metric, error)
	List(ctx context.Context) ([]model.Metric, error)
	AddBatch(ctx context.Context, arr []model.Metric) error
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
		filestore := filestore.New(
			filestore.Config{
				StorePath: cfg.StorePath,
				IsRestore: cfg.IsRestore,
				StoreInt:  cfg.StoreInt,
			}, inmemory.New())

		return adapter.Ping(filestore)
	default:
		inmem := inmemory.New()

		return adapter.Ping(inmem)
	}
}
