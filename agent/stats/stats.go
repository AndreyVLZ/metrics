package stats

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync/atomic"

	"github.com/AndreyVLZ/metrics/internal/model"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type Name uint

const totalRuntimeMetric = 29 // Кол-во метрик из пакета runtime.

const (
	TotalMetric int = 32 // Общее кол-во метрик.
)

// Константы поддерживаемых метрик:
// 29 метрик из пакета runtime:
// 28:[gauge], 1:[counter]
// 3 метрики из пакета goUtil:
// 3:[gauge]
const (
	Alloc Name = iota
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
	RandomValue
	PollCount // total runtime
	TotalMemory
	FreeMemory
	CPUutilization1
)

func supportName() [TotalMetric]string {
	return [TotalMetric]string{
		// [runtime] int64
		"Alloc", "BuckHashSys", "Frees", "GCSys", "HeapAlloc", "HeapIdle",
		"HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups",
		"MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC",
		"OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc",
		// int32
		"NumForcedGC", "NumGC",
		// float64
		"GCCPUFraction", "RandomValue",
		// int64
		"PollCount",
		// [util] uint64
		"TotalMemory", "FreeMemory",
		// int
		"CPUutilization1",
	}
}

// Возвращает имя метрики.
func (n Name) String() string { return supportName()[n] }

var (
	// Массив функций для чтения метрик из пакета gopsutil.
	arrFuncUtilRead = [3]func(*utilStats) float64{
		func(u *utilStats) float64 { return float64(u.memStats.Available) },
		func(u *utilStats) float64 { return float64(u.memStats.Free) },
		func(_ *utilStats) float64 { cpuCount, _ := cpu.Counts(true); return float64(cpuCount) },
	}

	// Массив функций для чтения метрик из пакета runtime.
	arrFuncRuntimeRead = [28]func(*runtimeStats) float64{
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.Alloc) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.BuckHashSys) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.Frees) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.GCSys) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.HeapAlloc) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.HeapIdle) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.HeapInuse) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.HeapObjects) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.HeapReleased) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.HeapSys) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.LastGC) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.Lookups) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.MCacheInuse) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.MCacheSys) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.MSpanInuse) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.MSpanSys) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.Mallocs) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.NextGC) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.OtherSys) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.PauseTotalNs) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.StackInuse) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.StackSys) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.Sys) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.TotalAlloc) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.NumForcedGC) },
		func(s *runtimeStats) float64 { s.total.Add(1); return float64(s.memStats.NumGC) },
		func(s *runtimeStats) float64 { s.total.Add(1); return s.memStats.GCCPUFraction },
		func(_ *runtimeStats) float64 { return rand.ExpFloat64() },
	}
)

// Статистика.
type Stats struct {
	rtStats   *runtimeStats
	utilStats utilStats
}

func New() *Stats {
	return &Stats{
		rtStats:   newRuntimeStats(),
		utilStats: utilStats{},
	}
}

// Инициализирует пакет gopsutil.
func (s *Stats) Init() error { return s.utilStats.Init() }

// Возвращает срез метрик, прочитанных из пакета gopsutil.
func (s *Stats) UtilList() []model.Metric {
	return s.readList(TotalMemory, CPUutilization1)
}

// Возвращает срез метрик, прочитанных из пакета runtime.
func (s *Stats) RuntimeList() []model.Metric {
	return s.readList(Alloc, PollCount)
}

func (s *Stats) readList(start, stop Name) []model.Metric {
	list := make([]model.Metric, 0, stop-start)

	for iName := start; iName <= stop; iName++ {
		list = append(list, s.readMetric(iName))
	}

	return list
}

func (s *Stats) readMetric(metName Name) model.Metric {
	switch {
	case metName >= Alloc && metName <= RandomValue:
		runtime.ReadMemStats(&s.rtStats.memStats)

		val := arrFuncRuntimeRead[metName]
		aval := val(s.rtStats)

		return model.NewGaugeMetric(metName.String(), aval)
	case metName >= TotalMemory && metName <= CPUutilization1:
		l := metName - totalRuntimeMetric
		val := arrFuncUtilRead[l]
		aval := val(&s.utilStats)

		return model.NewGaugeMetric(metName.String(), aval)
	default:
		return model.NewCounterMetric(metName.String(), s.rtStats.total.Load())
	}
}

// Статистика для runtime.
type runtimeStats struct {
	memStats runtime.MemStats
	total    atomic.Int64
}

func newRuntimeStats() *runtimeStats {
	var memStats runtime.MemStats

	return &runtimeStats{
		memStats: memStats,
		total:    atomic.Int64{},
	}
}

// Статистика для gopsutil.
type utilStats struct {
	memStats *mem.VirtualMemoryStat
}

// Инициализация gopsutil.
func (us *utilStats) Init() error {
	vmStats, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	us.memStats = vmStats

	return nil
}
