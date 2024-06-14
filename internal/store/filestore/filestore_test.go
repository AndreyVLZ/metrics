package filestore

import (
	"context"
	"sync"
	"testing"

	"github.com/AndreyVLZ/metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

type spyStore struct {
	storage
	err error
	arr []model.Metric
}

func newSpyStore() *spyStore {
	return &spyStore{}
}

func (ss *spyStore) Start(_ context.Context) error {
	return nil
}

func (ss *spyStore) Stop(_ context.Context) error {
	return nil
}

func (ss *spyStore) AddBatch(_ context.Context, batch []model.Metric) error {
	ss.arr = batch
	return ss.err
}

func (ss *spyStore) List(_ context.Context) ([]model.Metric, error) {
	return ss.arr, ss.err
}

type spyFile struct {
	err error
	met model.Metric
	arr []model.Metric
	mu  sync.Mutex
}

func (sf *spyFile) WriteMetric(met model.Metric) error {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	sf.met = met
	return sf.err
}

func (sf *spyFile) WriteBatch(arr []model.Metric) error {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	sf.arr = arr

	return nil
}

func (sf *spyFile) ReadBatch() ([]model.Metric, error) {
	return sf.arr, sf.err
}

func (sf *spyFile) Open() error {
	return sf.err
}

func (sf *spyFile) Close() error {
	return sf.err
}

func TestFileStoreStart(t *testing.T) {
	ctx := context.Background()

	t.Run("start store as synchro ok", func(t *testing.T) {
		wantIsDeamon := false
		cfg := Config{
			StorePath: "testFileStore.json",
			IsRestore: false,
			StoreInt:  0,
		}
		spyFile := &spyFile{}
		spyStore := &spyStore{}

		fileStore := New(cfg, spyStore)
		fileStore.file = spyFile

		err := fileStore.Start(ctx)
		assert.NoError(t, err)

		assert.Equal(t, wantIsDeamon, fileStore.isDeamon)
	})

	/*
		t.Run("start store as deamon ok", func(t *testing.T) {
			wantIsDeamon := true
			wantArr := []model.Metric{
				model.NewCounterMetric("Counter-1", 1000),
			}
			cfg := Config{
				StorePath: "testFileStore.json",
				IsRestore: false,
				StoreInt:  1,
			}
			spyFile := &spyFile{}
			spyStore := &spyStore{arr: wantArr}

			fileStore := New(cfg, spyStore)
			fileStore.file = spyFile

			err := fileStore.Start(ctx)
			assert.NoError(t, err)

			time.Sleep(1200 * time.Millisecond)

			assert.Equal(t, wantIsDeamon, fileStore.isDeamon)
			assert.Equal(t, wantArr, spyFile.arr)
		})
			t.Run("start store and restore", func(t *testing.T) {
				wantArr := []model.Metric{
					model.NewCounterMetric("Counter-1", 1000),
				}

				cfg := Config{
					StorePath: "testFileStore.json",
					IsRestore: true,
					StoreInt:  0,
				}

				spyFile := &spyFile{arr: wantArr}
				spyStore := &spyStore{}

				fileStore := New(cfg, spyStore)
				fileStore.file = spyFile

				err := fileStore.Start(ctx)
				assert.NoError(t, err)
				assert.Equal(t, wantArr, spyStore.arr)
			})
	*/
}

func TestFileStoreStop(t *testing.T) {
	ctx := context.Background()

	wantArr := []model.Metric{
		model.NewCounterMetric("Counter-1", 1000),
	}

	cfg := Config{
		StorePath: "testFileStore.json",
		IsRestore: false,
		StoreInt:  1,
	}

	spyFile := &spyFile{}
	spyStore := &spyStore{arr: wantArr}
	fileStore := New(cfg, spyStore)
	fileStore.file = spyFile

	err := fileStore.Stop(ctx)
	assert.NoError(t, err)
	assert.Equal(t, wantArr, spyFile.arr)
}
