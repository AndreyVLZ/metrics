package filestore

import (
	"os"
	"testing"

	"github.com/AndreyVLZ/metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestBuildMetric(t *testing.T) {
	metInt := model.NewMetricRepo[int64]("Counter1", model.TypeCountConst, 1)
	delta := int64(1)
	wantInt := metric{Name: "Counter1", MType: model.TypeCountConst.String(), Delta: &delta}
	metFloat := model.NewMetricRepo[float64]("Gauge1", model.TypeGaugeConst, 1.1)
	val := float64(1.1)
	wantFloat := metric{Name: "Gauge1", MType: model.TypeGaugeConst.String(), Val: &val}

	assert.Equal(t, wantInt, buildMetric[int64](metInt))
	assert.Equal(t, wantFloat, buildMetric[float64](metFloat))
}

func TestFile(t *testing.T) {
	fileName := "temp.json"

	cList := []model.MetricRepo[int64]{
		model.NewMetricRepo[int64]("Counter-1", model.TypeCountConst, 1),
		model.NewMetricRepo[int64]("Counter-2", model.TypeCountConst, 2),
	}

	gList := []model.MetricRepo[float64]{
		model.NewMetricRepo[float64]("Gauge-1", model.TypeGaugeConst, 1.1),
		model.NewMetricRepo[float64]("Gauge-2", model.TypeGaugeConst, 2.2),
	}

	file := NewFile(fileName)
	if err := file.Open(); err != nil {
		t.Errorf("err open file %v\n", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("err close file: %v\n", err)
		}

		if err := os.Remove(fileName); err != nil {
			t.Errorf("err remove file: %v\n", err)
		}
	}()

	fnWriteMetric(t, "count", file, cList)
	fnWriteMetric(t, "gauge", file, gList)

	batch, err := file.ReadBatch()
	if err != nil {
		t.Errorf("err read batch: %v\n", err)
	}

	b := model.Batch{CList: cList, GList: gList}

	assert.Equal(t, b, batch)

	if err := file.WriteBatch(b); err != nil {
		t.Errorf("err write batch: %v\n", err)
	}
}

func fnReadBatch[VT model.ValueType](t *testing.T, name string, file *File, list []model.MetricRepo[VT]) {
	t.Run("read batch "+name, func(t *testing.T) {
		batch, err := file.ReadBatch()
		if err != nil {
			t.Errorf("err read batch: %v\n", err)
		}
		assert.Equal(t, list, batch)
	})
}

func fnWriteMetric[VT model.ValueType](t *testing.T, name string, file *File, list []model.MetricRepo[VT]) {
	t.Run("write metric "+name, func(t *testing.T) {
		for _, met := range list {
			if err := file.WriteMetric(buildMetric(met)); err != nil {
				t.Errorf("err write metric: %v\n", err)
			}
		}
	})
}
