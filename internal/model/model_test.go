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
		val := NewCounterValue(updVal)

		err := met.Update(val)

		if assert.NoError(t, err) {
			assert.Equal(t, valWant, met.GetDelta())
		}
	})

	t.Run("update gauge", func(t *testing.T) {
		var (
			initVal = 10.01
			updVal  = 20.02
			valWant = 20.02
		)

		met := NewGaugeMetric("Gauge-1", initVal)
		val := NewGaugeValue(updVal)

		err := met.Update(val)

		if assert.NoError(t, err) {
			assert.Equal(t, valWant, met.GetValue())
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

func TestParseMetricJSON(t *testing.T) {
	t.Run("parse counter ok", func(t *testing.T) {
		var delta int64 = 10

		wantJSON := MetricJSON{ID: "Counter-1", MType: "counter", Delta: &delta}
		metStr := MetricStr{
			InfoStr: InfoStr{Name: "Counter-1", MType: "counter"},
			Val:     "10",
		}

		metJSON, err := ParseMetricJSON(metStr)
		assert.NoError(t, err)

		assert.Equal(t, wantJSON, metJSON)
	})

	t.Run("parse gauge ok", func(t *testing.T) {
		var val = 10.01

		wantJSON := MetricJSON{ID: "Gauge-1", MType: "gauge", Value: &val}
		metStr := MetricStr{
			InfoStr: InfoStr{Name: "Gauge-1", MType: "gauge"},
			Val:     "10.01",
		}

		metJSON, err := ParseMetricJSON(metStr)
		assert.NoError(t, err)

		assert.Equal(t, wantJSON, metJSON)
	})

	t.Run("parse err type not support", func(t *testing.T) {
		metStr := MetricStr{
			InfoStr: InfoStr{Name: "Counter-1", MType: "c"},
			Val:     "10",
		}

		_, err := ParseMetricJSON(metStr)
		if err == nil {
			t.Error("want err")
		}
	})

	t.Run("parse err not correct delta", func(t *testing.T) {
		metStr := MetricStr{
			InfoStr: InfoStr{Name: "Counter-1", MType: "counter"},
			Val:     "c",
		}

		_, err := ParseMetricJSON(metStr)
		if err == nil {
			t.Error("want err")
		}
	})

	t.Run("parse err not correct value", func(t *testing.T) {
		metStr := MetricStr{
			InfoStr: InfoStr{Name: "Gauge-1", MType: "gauge"},
			Val:     "c",
		}

		_, err := ParseMetricJSON(metStr)
		if err == nil {
			t.Error("want err")
		}
	})
}

func TestParseMetric(t *testing.T) {
	t.Run("parse counter ok", func(t *testing.T) {
		var delta int64 = 10
		metJSON := MetricJSON{ID: "Counter-1", MType: "counter", Delta: &delta}

		wanMet := NewCounterMetric("Counter-1", delta)

		met, err := ParseMetric(metJSON)

		assert.NoError(t, err)
		assert.Equal(t, wanMet, met)
	})

	t.Run("parse gauge ok", func(t *testing.T) {
		var val = 10.01
		metJSON := MetricJSON{ID: "Gauge-1", MType: "gauge", Value: &val}

		wanMet := NewGaugeMetric("Gauge-1", val)

		met, err := ParseMetric(metJSON)

		assert.NoError(t, err)
		assert.Equal(t, wanMet, met)
	})

	t.Run("parse err parseInfo ok", func(t *testing.T) {
		var delta int64 = 10
		metJSON := MetricJSON{ID: "", MType: "counter", Delta: &delta}
		_, err := ParseMetric(metJSON)
		if err == nil {
			t.Error("want err")
		}
	})
}

func TestBuildArrMetric(t *testing.T) {
	t.Run("build arr ok", func(t *testing.T) {
		var delta int64 = 10
		var val = 10.01

		arrMetJSON := []MetricJSON{
			{ID: "Counter-1", MType: "counter", Delta: &delta},
			{ID: "Gauge-1", MType: "gauge", Value: &val},
		}

		wantArr := []Metric{
			NewCounterMetric("Counter-1", 10),
			NewGaugeMetric("Gauge-1", 10.01),
		}

		arr, err := BuildArrMetric(arrMetJSON)
		assert.NoError(t, err)
		assert.Equal(t, wantArr, arr)
	})

	t.Run("build arr err parse", func(t *testing.T) {
		var delta int64 = 10
		var val = 10.01

		arrMetJSON := []MetricJSON{
			{ID: "Counter-1", MType: "counter", Delta: &delta},
			{ID: "", MType: "gauge", Value: &val},
		}

		_, err := BuildArrMetric(arrMetJSON)
		if err == nil {
			t.Error("want err")
		}
	})
}

func BenchmarkMetricUpdateCounter(b *testing.B) {
	met := NewCounterMetric("Counter-1", 100)
	val := NewCounterValue(11)

	for i := 0; i < b.N; i++ {
		met.Update(val)
	}
}

func BenchmarkMetricUpdateGauge(b *testing.B) {
	met := NewGaugeMetric("Counter-1", 10.01)
	val := NewGaugeValue(20.02)

	for i := 0; i < b.N; i++ {
		met.Update(val)
	}
}
