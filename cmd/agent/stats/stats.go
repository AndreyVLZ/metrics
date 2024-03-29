package stats

import (
	"context"
	"runtime"

	m "github.com/AndreyVLZ/metrics/internal/metric"
	"github.com/AndreyVLZ/metrics/internal/storage"
)

// metricConst
type metricConst uint

// Константы поддерживаемых метрик для пакета runtime
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

// supportName Получение массива имен поддерживаемых метрик
func (mc metricConst) supportName() []string {
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

// String Возвращает имя метрики
func (mc metricConst) String() string {
	return mc.supportName()[mc]
}

// stats
type Stats struct {
	memStats runtime.MemStats
	total    m.Counter
}

func NewStats() *Stats {
	var memStats runtime.MemStats
	return &Stats{
		memStats: memStats,
	}
}

// Read обновляет статистику [memory allocator statistics] из пакета runtime,
// увеличивает инкремент total, возвращает значение для t метрики,
// возвращаемое значение типа Gauge(float64)
// Ex: Read(stats.Alloc)
func (s *Stats) Read(t metricConst) m.Gauge {
	runtime.ReadMemStats(&s.memStats)
	return []func() m.Gauge{
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.Alloc) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.BuckHashSys) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.Frees) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.GCSys) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.HeapAlloc) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.HeapIdle) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.HeapInuse) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.HeapObjects) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.HeapReleased) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.HeapSys) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.LastGC) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.Lookups) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.MCacheInuse) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.MCacheSys) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.MSpanInuse) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.MSpanSys) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.Mallocs) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.NextGC) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.OtherSys) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.PauseTotalNs) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.StackInuse) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.StackSys) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.Sys) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.TotalAlloc) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.NumForcedGC) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.NumGC) },
		func() m.Gauge { s.total++; return m.Gauge(s.memStats.GCCPUFraction) },
	}[t]()
}

// TypeName Поучение строкового имени типа метрики
func (s *Stats) TypeName() string {
	return m.GaugeType.String()
}

// Total Возвращает значения инкрeмента total типа Counter
func (s *Stats) Total() m.Counter {
	return s.total
}

// ReadToStore Получение и сохранение всех поддерживаемых метрик в хранилище s
func (s *Stats) ReadToStore(store storage.Storage) error {
	supportName := metricConst(0).supportName()
	for i := range supportName {
		_, err := store.Set(
			context.Background(),
			m.NewMetricDB(
				supportName[i],
				s.Read(metricConst(i)),
			),
		)
		if err != nil {
			return err
		}
	}

	_, err := store.Set(context.Background(), m.NewMetricDB("PollCount", s.Total()))
	if err != nil {
		return err
	}

	return nil
}
