// Реализация сохрениения метрик в файл синхронно либо по интервалу StoreInt для storage.
package filestore

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/AndreyVLZ/metrics/internal/model"
)

const NameConst = "file store"

type storage interface {
	Get(ctx context.Context, mInfo model.Info) (model.Metric, error)
	Update(ctx context.Context, met model.Metric) (model.Metric, error)
	List(ctx context.Context) ([]model.Metric, error)
	AddBatch(ctx context.Context, arr []model.Metric) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type Config struct {
	StorePath string
	IsRestore bool
	StoreInt  int
}

type iFile interface {
	WriteMetric(met model.Metric) error
	WriteBatch(arr []model.Metric) error
	ReadBatch() ([]model.Metric, error)
	Open() error
	Close() error
}

type FileStore struct {
	storage
	file     iFile
	exit     chan struct{}
	cfg      Config
	isDeamon bool
}

func New(cfg Config, store storage) *FileStore {
	return &FileStore{
		cfg:      cfg,
		file:     NewFile(cfg.StorePath),
		storage:  store,
		isDeamon: false,
		exit:     make(chan struct{}),
	}
}

func (fs FileStore) Name() string { return NameConst }

func (fs *FileStore) Start(ctx context.Context) error {
	if err := fs.file.Open(); err != nil {
		return fmt.Errorf("file Open: %w", err)
	}

	if err := fs.storage.Start(ctx); err != nil {
		return fmt.Errorf("store Start: %w", err)
	}

	if fs.cfg.IsRestore {
		batch, err := fs.file.ReadBatch()
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		if err := fs.storage.AddBatch(ctx, batch); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	if fs.cfg.StoreInt == 0 {
		fs.storage = newWrapStore(fs.file, fs.storage)

		fmt.Printf("run as synchro\n")

		return nil
	}

	fs.isDeamon = true
	go fs.run(ctx)

	fmt.Printf("run as deamon\n")

	return nil
}

func (fs *FileStore) Stop(ctx context.Context) error {
	if fs.isDeamon {
		<-fs.exit
	}

	if err := saved(ctx, fs.storage, fs.file); err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := fs.file.Close(); err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := fs.storage.Stop(ctx); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (fs *FileStore) run(ctx context.Context) {
	for {
		select {
		case <-time.After(time.Duration(fs.cfg.StoreInt) * time.Second):
			fmt.Println("FLUSH")

			if err := saved(ctx, fs.storage, fs.file); err != nil {
				log.Printf("err save metrics %v\n", err)
			}
		case <-ctx.Done():
			fs.exit <- struct{}{}

			return
		}
	}
}

func saved(ctx context.Context, store storage, file iFile) error {
	batch, err := store.List(ctx)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return file.WriteBatch(batch)
}
