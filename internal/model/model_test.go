package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricUpdate(t *testing.T) {
	t.Run("update counter", func(t *testing.T) {
		var (
			initVal int64 = 100
			updVal  int64 = 100
			valWant int64 = 200
		)

		met := NewCounterMetric("Counter-1", initVal)
		val := Value{Delta: &updVal}

		err := met.Update(val)

		if assert.NoError(t, err) {
			assert.Equal(t, valWant, *met.Delta)
		}
	})

	t.Run("update gauge", func(t *testing.T) {
		var (
			initVal = 10.01
			updVal  = 20.02
			valWant = 20.02
		)

		met := NewGaugeMetric("Gauge-1", initVal)
		val := Value{Val: &updVal}

		err := met.Update(val)

		if assert.NoError(t, err) {
			assert.Equal(t, valWant, *met.Val)
		}
	})

	t.Run("update err nil delta", func(t *testing.T) {
		var initVal int64 = 100

		met := NewCounterMetric("Counter-1", initVal)
		val := Value{}

		err := met.Update(val)
		if err == nil {
			t.Error("want err")
		}
	})

	t.Run("update err nil value", func(t *testing.T) {
		var initVal = 10.01

		met := NewGaugeMetric("Gauge-1", initVal)
		val := Value{}

		err := met.Update(val)
		if err == nil {
			t.Error("want err")
		}
	})

	t.Run("update err type not support", func(t *testing.T) {
		var initVal = 10.01

		met := NewMetric(Info{MName: "Type not support", MType: Type(2)}, Value{Val: &initVal})
		val := Value{}

		err := met.Update(val)
		if err == nil {
			t.Error("want err")
		}
	})
}

func TestMetricString(t *testing.T) {
	t.Run("counter string", func(t *testing.T) {
		var delta int64 = 10

		metCounter := MetricJSON{ID: "Counter-1", MType: "counter", Delta: &delta}
		assert.Equal(t, "10", metCounter.String())
	})

	t.Run("counter string", func(t *testing.T) {
		var val = 10.01

		metCounter := MetricJSON{ID: "Gauge-1", MType: "gauge", Value: &val}
		assert.Equal(t, "10.01", metCounter.String())
	})
}

func TestParseInfo(t *testing.T) {
	t.Run("parse counter", func(t *testing.T) {
		wantInfo := Info{MName: "Counter-1", MType: TypeCountConst}
		info, err := ParseInfo("Counter-1", "counter")
		assert.NoError(t, err)
		assert.Equal(t, wantInfo, info)
	})

	t.Run("parse gauge", func(t *testing.T) {
		wantInfo := Info{MName: "Gauge-1", MType: TypeGaugeConst}
		info, err := ParseInfo("Gauge-1", "gauge")
		assert.NoError(t, err)
		assert.Equal(t, wantInfo, info)
	})

	t.Run("parse err name empty", func(t *testing.T) {
		_, err := ParseInfo("", "counter")
		if err == nil {
			t.Error("want err")
		}
	})

	t.Run("parse err type not support", func(t *testing.T) {
		_, err := ParseInfo("Counter-1", "c")
		if err == nil {
			t.Error("want err")
		}
	})
}

func TestBuildMetricJSON(t *testing.T) {
	var delta int64 = 10

	wantMet := MetricJSON{ID: "Counter-1", MType: "counter", Delta: &delta}
	met := NewCounterMetric("Counter-1", 10)
	metJSON := BuildMetricJSON(met)

	assert.Equal(t, wantMet, metJSON)
}

func TestBuildArrMetricJSON(t *testing.T) {
	var delta int64 = 10
	var val = 10.01

	wantArr := []MetricJSON{
		{ID: "Counter-1", MType: "counter", Delta: &delta},
		{ID: "Gauge-1", MType: "gauge", Value: &val},
	}

	arrMet := []Metric{
		NewCounterMetric("Counter-1", 10),
		NewGaugeMetric("Gauge-1", 10.01),
	}

	arrMetJSON := BuildArrMetricJSON(arrMet)

	assert.Equal(t, wantArr, arrMetJSON)
}

func BenchmarkMetricUpdateCounter(b *testing.B) {
	met := NewCounterMetric("Counter-1", 100)
	var delta int64 = 11
	val := Value{Delta: &delta}

	for i := 0; i < b.N; i++ {
		met.Update(val)
	}
}

func BenchmarkMetricUpdateGauge(b *testing.B) {
	met := NewGaugeMetric("Counter-1", 10.01)
	value := 10.01
	val := Value{Val: &value}

	for i := 0; i < b.N; i++ {
		met.Update(val)
	}
}
