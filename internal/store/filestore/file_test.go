package filestore

import (
	"os"
	"testing"

	"github.com/AndreyVLZ/metrics/internal/model"
)

func TestFile(t *testing.T) {
	fileName := "temp.json"

	file := NewFile(fileName)

	if err := file.Open(); err != nil {
		t.Errorf("file open err %v\n", err)
	}

	t.Cleanup(func() {
		if err := file.Close(); err != nil {
			t.Errorf("err close file: %v\n", err)
		}

		if err := os.Remove(fileName); err != nil {
			t.Errorf("err remove file: %v\n", err)
		}
	})

	t.Run("write metric", func(t *testing.T) {
		met := model.NewCounterMetric("Counter-1", 10)
		if err := file.WriteMetric(met); err != nil {
			t.Errorf("write metric err %v\n", err)
		}
	})

	var (
		batch []model.Metric
		err   error
	)

	t.Run("read batch", func(t *testing.T) {
		batch, err = file.ReadBatch()
		if err != nil {
			t.Errorf("read batch err %v\n", err)
		}
	})

	t.Run("write batch", func(t *testing.T) {
		if err := file.WriteBatch(batch); err != nil {
			t.Errorf("write batch err %v\n", err)
		}
	})
}
