package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStats(t *testing.T) {
	stats := New()
	if err := stats.Init(); err != nil {
		t.Fatal(err)
	}

	rtList := stats.RuntimeList()
	t.Run("len count arr", func(t *testing.T) {
		assert.Equal(t, 29, len(rtList))
	})

	uList := stats.UtilList()
	t.Run("len gauge arr", func(t *testing.T) {
		assert.Equal(t, 3, len(uList))
	})
}

func TestSupportName(t *testing.T) {
	arrWantMet := []Name{
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
		RandomValue,
		PollCount,
		TotalMemory,
		FreeMemory,
		CPUutilization1,
	}

	names := supportName()
	assert.Equal(t, len(arrWantMet), len(names))
	for i, met := range arrWantMet {
		assert.Equal(t, met.String(), names[i])
	}
}
