package inmemory

import (
	"context"
	"testing"

	"github.com/AndreyVLZ/metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestMemStore(t *testing.T) {
	ctx := context.Background()
	mem := New()

	t.Run("update_counter_1", func(t *testing.T) {
		metInsert := model.NewCounterMetric("Counter-1", 100)

		metDB, err := mem.Update(ctx, metInsert)
		if assert.NoError(t, err) {
			assert.Equal(t, metInsert, metDB)
		}
	})

	t.Run("update_counter_2", func(t *testing.T) {
		metInsert := model.NewCounterMetric("Counter-1", 300)
		metWant := model.NewCounterMetric("Counter-1", 400)

		metDB, err := mem.Update(ctx, metInsert)
		if assert.NoError(t, err) {
			assert.Equal(t, metWant, metDB)
		}
	})

	t.Run("update_gauge_1", func(t *testing.T) {
		metInsert := model.NewGaugeMetric("Gauge-1", 10.01)

		metDB, err := mem.Update(ctx, metInsert)
		if assert.NoError(t, err) {
			assert.Equal(t, metInsert, metDB)
		}
	})

	t.Run("update_gauge_2", func(t *testing.T) {
		metInsert := model.NewGaugeMetric("Gauge-1", 20.02)
		metWant := model.NewGaugeMetric("Gauge-1", 20.02)

		metDB, err := mem.Update(ctx, metInsert)
		if assert.NoError(t, err) {
			assert.Equal(t, metWant, metDB)
		}
	})

	t.Run("get_counter", func(t *testing.T) {
		metWant := model.NewCounterMetric("Counter-1", 400)
		info := metWant.Info

		metDB, err := mem.Get(ctx, info)
		if assert.NoError(t, err) {
			assert.Equal(t, metWant, metDB)
		}
	})

	t.Run("get_gauge", func(t *testing.T) {
		metWant := model.NewGaugeMetric("Gauge-1", 20.02)
		info := metWant.Info

		metDB, err := mem.Get(ctx, info)
		if assert.NoError(t, err) {
			assert.Equal(t, metWant, metDB)
		}
	})

	t.Run("batch", func(t *testing.T) {
		mem := New()
		arr := []model.Metric{
			model.NewGaugeMetric("Gauge-1", 10.01),
			model.NewGaugeMetric("Gauge-2", 20.02),
			model.NewCounterMetric("Counter-1", 200),
		}

		err := mem.AddBatch(ctx, arr)
		assert.NoError(t, err)

		arrDB, err := mem.List(ctx)
		if assert.NoError(t, err) {
			assert.ElementsMatch(t, arr, arrDB)
		}
	})

	t.Run("get_errNotFind", func(t *testing.T) {
		mem := New()
		info := model.NewGaugeInfo("NON")

		_, err := mem.Get(ctx, info)
		assert.Equal(t, errNotFind, err)
	})
}

func BenchmarkMemStoreUpdate(b *testing.B) {
	ctx := context.Background()
	mem := New()
	metInsert := model.NewCounterMetric("Counter-1", 100)

	for i := 0; i < b.N; i++ {
		mem.Update(ctx, metInsert)
	}
}

func BenchmarkMemStoreGet(b *testing.B) {
	ctx := context.Background()
	mem := New()

	metInsert := model.NewCounterMetric("Counter-1", 100)
	mem.Update(ctx, metInsert)

	info := metInsert.Info

	for i := 0; i < b.N; i++ {
		mem.Get(ctx, info)
	}
}

func BenchmarkMemStoreAddBatch(b *testing.B) {
	ctx := context.Background()
	mem := New()

	arr := []model.Metric{
		model.NewGaugeMetric("Gauge-1", 10.01),
		model.NewGaugeMetric("Gauge-2", 20.02),
		model.NewCounterMetric("Counter-1", 200),
	}

	for i := 0; i < b.N; i++ {
		mem.AddBatch(ctx, arr)
	}
}

func BenchmarkMemStoreList(b *testing.B) {
	ctx := context.Background()
	mem := New()

	arr := []model.Metric{
		model.NewGaugeMetric("Gauge-1", 10.01),
		model.NewGaugeMetric("Gauge-2", 20.02),
		model.NewCounterMetric("Counter-1", 200),
	}
	mem.AddBatch(ctx, arr)

	for i := 0; i < b.N; i++ {
		mem.List(ctx)
	}
}
