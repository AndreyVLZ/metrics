package filestore

import (
	"context"
	"fmt"
	"log"
	_ "net/http/pprof"
	"time"

	"github.com/AndreyVLZ/metrics/internal/model"
)

const NameConst = "file store"

type storage interface {
	Name() string
	Ping() error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	UpdateCounter(ctx context.Context, met model.MetricRepo[int64]) (model.MetricRepo[int64], error)
	UpdateGauge(ctx context.Context, met model.MetricRepo[float64]) (model.MetricRepo[float64], error)
	GetCounter(ctx context.Context, name string) (model.MetricRepo[int64], error)
	GetGauge(ctx context.Context, name string) (model.MetricRepo[float64], error)
	AddBatch(ctx context.Context, batch model.Batch) error
	List(ctx context.Context) (model.Batch, error)
}

type Config struct {
	StorePath string
	IsRestore bool
	StoreInt  int
}

type FileStore struct {
	storage
	cfg      Config
	file     *File
	isDeamon bool
	exit     chan struct{}
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
func (fs FileStore) Ping() error  { return nil }

func (fs *FileStore) Start(ctx context.Context) error {
	if err := fs.file.Open(); err != nil {
		return fmt.Errorf("%w", err)
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

func saved(ctx context.Context, store storage, file *File) error {
	batch, err := store.List(ctx)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return file.WriteBatch(batch)
}

type wrapStore struct {
	file *File
	storage
}

func newWrapStore(file *File, s storage) *wrapStore {
	return &wrapStore{
		file:    file,
		storage: s,
	}
}

func (ws *wrapStore) UpdateCounter(ctx context.Context, met model.MetricRepo[int64]) (model.MetricRepo[int64], error) {
	met, err := ws.storage.UpdateCounter(ctx, met)
	if err != nil {
		return model.MetricRepo[int64]{}, fmt.Errorf("%w", err)
	}

	if err := ws.file.WriteMetric(buildMetric(met)); err != nil {
		log.Printf("file write count %s\n", err)
	}

	return met, nil
}

func (ws *wrapStore) UpdateGauge(ctx context.Context, met model.MetricRepo[float64]) (model.MetricRepo[float64], error) {
	met, err := ws.storage.UpdateGauge(ctx, met)
	if err != nil {
		return model.MetricRepo[float64]{}, fmt.Errorf("%w", err)
	}

	if err := ws.file.WriteMetric(buildMetric(met)); err != nil {
		log.Printf("file write gauge %s\n", err)
	}

	return met, nil
}

func (ws *wrapStore) AddBatch(ctx context.Context, batch model.Batch) error {
	if err := ws.storage.AddBatch(ctx, batch); err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := ws.file.WriteBatch(batch); err != nil {
		log.Printf("FileStore: %s\n", err)
	}

	return nil
}
