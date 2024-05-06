package stats

import (
	"runtime"

	"github.com/AndreyVLZ/metrics/internal/model"
)

// metricConst.
type metricConst uint

// Константы поддерживаемых метрик для пакета runtime.
const (
	Alloc metricConst = iota
	BuckHashSys
	Frees
	GCSys
	HeapAlloc
	HeapIdle
	HeapInuse
	HeapObjects
	HeapReleased
	HeapSys
	LastGC
	Lookups
	MCacheInuse
	MCacheSys
	MSpanInuse
	MSpanSys
	Mallocs
	NextGC
	OtherSys
	PauseTotalNs
	StackInuse
	StackSys
	Sys
	TotalAlloc
	NumForcedGC
	NumGC
	GCCPUFraction
)

// supportName Получение массива имен поддерживаемых метрик.
func supportName() []string {
	return []string{
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
}

// String Возвращает имя метрики.
func (mc metricConst) String() string {
	return supportName()[mc]
}

// stats.
type Stats struct {
	memStats runtime.MemStats
	total    int64
}

func New() *Stats {
	var memStats runtime.MemStats

	return &Stats{
		memStats: memStats,
		total:    0,
	}
}

func (s *Stats) BuildBatch() (model.Batch, error) {
	supportName := supportName()
	arrFuncRead := s.arrFuncRead()

	listGauge := make([]model.MetricRepo[float64], len(supportName))

	for i := range supportName {
		listGauge[i] = model.NewMetricRepo(
			supportName[i], model.TypeGaugeConst, arrFuncRead[i]())
	}

	listCount := []model.MetricRepo[int64]{
		model.NewMetricRepo("PollCount", model.TypeCountConst, s.total),
	}

	return model.Batch{CList: listCount, GList: listGauge}, nil
}

func (s *Stats) arrFuncRead() []func() float64 {
	runtime.ReadMemStats(&s.memStats)

	return []func() float64{
		func() float64 { s.total++; return float64(s.memStats.Alloc) },
		func() float64 { s.total++; return float64(s.memStats.BuckHashSys) },
		func() float64 { s.total++; return float64(s.memStats.Frees) },
		func() float64 { s.total++; return float64(s.memStats.GCSys) },
		func() float64 { s.total++; return float64(s.memStats.HeapAlloc) },
		func() float64 { s.total++; return float64(s.memStats.HeapIdle) },
		func() float64 { s.total++; return float64(s.memStats.HeapInuse) },
		func() float64 { s.total++; return float64(s.memStats.HeapObjects) },
		func() float64 { s.total++; return float64(s.memStats.HeapReleased) },
		func() float64 { s.total++; return float64(s.memStats.HeapSys) },
		func() float64 { s.total++; return float64(s.memStats.LastGC) },
		func() float64 { s.total++; return float64(s.memStats.Lookups) },
		func() float64 { s.total++; return float64(s.memStats.MCacheInuse) },
		func() float64 { s.total++; return float64(s.memStats.MCacheSys) },
		func() float64 { s.total++; return float64(s.memStats.MSpanInuse) },
		func() float64 { s.total++; return float64(s.memStats.MSpanSys) },
		func() float64 { s.total++; return float64(s.memStats.Mallocs) },
		func() float64 { s.total++; return float64(s.memStats.NextGC) },
		func() float64 { s.total++; return float64(s.memStats.OtherSys) },
		func() float64 { s.total++; return float64(s.memStats.PauseTotalNs) },
		func() float64 { s.total++; return float64(s.memStats.StackInuse) },
		func() float64 { s.total++; return float64(s.memStats.StackSys) },
		func() float64 { s.total++; return float64(s.memStats.Sys) },
		func() float64 { s.total++; return float64(s.memStats.TotalAlloc) },
		func() float64 { s.total++; return float64(s.memStats.NumForcedGC) },
		func() float64 { s.total++; return float64(s.memStats.NumGC) },
		func() float64 { s.total++; return s.memStats.GCCPUFraction },
	}
}
