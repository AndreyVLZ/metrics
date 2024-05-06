package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildBatch(t *testing.T) {
	stats := New()

	batch, err := stats.BuildBatch()
	if err != nil {
		t.Errorf("err buildBatch: %v\n", err)
	}

	t.Run("len count arr", func(t *testing.T) {
		assert.Equal(t, 1, len(batch.CList))
	})

	t.Run("len gauge arr", func(t *testing.T) {
		assert.Equal(t, 27, len(batch.GList))
	})

}

func TestSupportName(t *testing.T) {
	arrWant := []string{
		// int64
		"Alloc", "BuckHashSys", "Frees", "GCSys", "HeapAlloc", "HeapIdle",
		"HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups",
		"MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC",
		"OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc",
		// int32
		"NumForcedGC", "NumGC",
		// float64
		"GCCPUFraction",
	}
	arr := supportName()
	for i, name := range arr {
		assert.Equal(t, arrWant[i], name)
	}

	arrWantMet := []metricConst{
		Alloc,
		BuckHashSys,
		Frees,
		GCSys,
		HeapAlloc,
		HeapIdle,
		HeapInuse,
		HeapObjects,
		HeapReleased,
		HeapSys,
		LastGC,
		Lookups,
		MCacheInuse,
		MCacheSys,
		MSpanInuse,
		MSpanSys,
		Mallocs,
		NextGC,
		OtherSys,
		PauseTotalNs,
		StackInuse,
		StackSys,
		Sys,
		TotalAlloc,
		NumForcedGC,
		NumGC,
		GCCPUFraction,
	}
	for i, met := range arrWantMet {
		assert.Equal(t, arrWant[i], met.String())
	}
}
