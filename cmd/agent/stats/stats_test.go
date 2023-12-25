package stats

import (
	"testing"

	"github.com/AndreyVLZ/metrics/internal/metric"
	"github.com/stretchr/testify/assert"
)

type mConst struct {
	mConst metricConst
	want   string
}

var allMetric = [27]mConst{
	{
		mConst: Alloc,
		want:   "Alloc",
	},
	{
		mConst: BuckHashSys,
		want:   "BuckHashSys",
	},
	{
		mConst: Frees,
		want:   "Frees",
	},
	{
		mConst: GCSys,
		want:   "GCSys",
	},
	{
		mConst: HeapAlloc,
		want:   "HeapAlloc",
	},
	{
		mConst: HeapIdle,
		want:   "HeapIdle",
	},
	{
		mConst: HeapInuse,
		want:   "HeapInuse",
	},
	{
		mConst: HeapObjects,
		want:   "HeapObjects",
	},
	{
		mConst: HeapReleased,
		want:   "HeapReleased",
	},
	{
		mConst: HeapSys,
		want:   "HeapSys",
	},
	{
		mConst: LastGC,
		want:   "LastGC",
	},
	{
		mConst: Lookups,
		want:   "Lookups",
	},
	{
		mConst: MCacheInuse,
		want:   "MCacheInuse",
	},
	{
		mConst: MCacheSys,
		want:   "MCacheSys",
	},
	{
		mConst: MSpanInuse,
		want:   "MSpanInuse",
	},
	{
		mConst: MSpanSys,
		want:   "MSpanSys",
	},
	{
		mConst: Mallocs,
		want:   "Mallocs",
	},
	{
		mConst: NextGC,
		want:   "NextGC",
	},
	{
		mConst: OtherSys,
		want:   "OtherSys",
	},
	{
		mConst: PauseTotalNs,
		want:   "PauseTotalNs",
	},
	{
		mConst: StackInuse,
		want:   "StackInuse",
	},
	{
		mConst: StackSys,
		want:   "StackSys",
	},
	{
		mConst: Sys,
		want:   "Sys",
	},
	{
		mConst: TotalAlloc,
		want:   "TotalAlloc",
	},
	{
		mConst: NumForcedGC,
		want:   "NumForcedGC",
	},
	{
		mConst: NumGC,
		want:   "NumGC",
	},
	{
		mConst: GCCPUFraction,
		want:   "GCCPUFraction",
	},
}

func TestMetricConstString(t *testing.T) {
	for _, m := range allMetric {
		assert.Equal(t, m.mConst.String(), m.want)
	}
}

func TestRead(t *testing.T) {
	stats := NewStats()
	for _, m := range allMetric {
		stats.Read(m.mConst)
	}
	assert.Equal(t, len(allMetric), int(stats.Total()))
}

/*
func TestAll(t *testing.T) {
	stats := NewStats()
	repo := stats.All()
	assert.Equal(t, len(metricConst(0).supportName()), len(repo))
}

func TestAllByRepo(t *testing.T) {
	stats := NewStats()
	repo := stats.AllByRepo()
	assert.Equal(t, len(metricConst(0).supportName()), len(repo.List()))
}
*/

func TestTotal(t *testing.T) {
	type a struct {
		start int
		end   int
		total int
	}

	tests := []a{
		{
			start: 0,
			end:   26,
			total: 26,
		},
		{
			start: 1,
			end:   11,
			total: 10,
		},
		{
			start: 10,
			end:   12,
			total: 2,
		},
		{
			start: 25,
			end:   26,
			total: 1,
		},
	}

	for _, test := range tests {
		stats := NewStats()
		for _, m := range allMetric[test.start:test.end] {
			stats.Read(m.mConst)
		}
		assert.Equal(t, int(stats.Total()), test.total)
	}
}

func TestTypeName(t *testing.T) {
	stats := NewStats()
	assert.Equal(t, metric.GaugeType.String(), stats.TypeName())
}
