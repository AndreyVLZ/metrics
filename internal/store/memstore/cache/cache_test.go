package cache

import (
	"testing"

	"github.com/AndreyVLZ/metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

type data[VT model.ValueType] struct {
	name string
	val  VT
	isOK bool
}

type test[VT model.ValueType] struct {
	name     string
	mtype    model.TypeMetric
	ch       *Cache[VT]
	init     []data[VT]
	check    []data[VT]
	totalLen int
}

type tCase struct {
	tInt   test[int64]
	tFloat test[float64]
}

func TestSet(t *testing.T) {
	tCase := tCase{
		tInt: test[int64]{
			name:  "counter",
			mtype: model.TypeCountConst,
			ch:    New[int64](),
			init: []data[int64]{
				{
					name: "Count-1",
					val:  10,
				},
				{
					name: "Count-2",
					val:  20,
				},
				{
					name: "Count-2",
					val:  30,
				},
			},
			check: []data[int64]{
				{
					name: "Count-1",
					val:  10,
					isOK: true,
				},
				{
					name: "Count-2",
					val:  30,
					isOK: true,
				},
				{
					name: "Count-0",
					val:  30,
					isOK: false,
				},
			},
			totalLen: 2,
		},

		tFloat: test[float64]{
			name:  "gauge",
			mtype: model.TypeGaugeConst,
			ch:    New[float64](),
			init: []data[float64]{
				{
					name: "Gauge-1",
					val:  10.01,
				},
				{
					name: "Gauge-2",
					val:  20.02,
				},
				{
					name: "Gauge-2",
					val:  30.03,
				},
			},
			check: []data[float64]{
				{
					name: "Gauge-1",
					val:  10.01,
					isOK: true,
				},
				{
					name: "Gauge-2",
					val:  30.03,
					isOK: true,
				},
				{
					name: "Gauge-0",
					val:  0,
					isOK: false,
				},
			},
			totalLen: 2,
		},
	}

	fCheck(t, tCase.tInt)
	fCheck(t, tCase.tFloat)
}

func fCheck[VT model.ValueType](t *testing.T, test test[VT]) {
	t.Run(test.name, func(t *testing.T) {
		cache := test.ch
		mtype := test.mtype

		for i := range test.init {
			data := test.init[i]
			cache.Set(model.NewMetricRepo(data.name, mtype, data.val))
		}

		for i := range test.check {
			act, ok := cache.Get(test.check[i].name)
			assert.Equal(t, test.check[i].isOK, ok)

			if test.check[i].isOK {
				assert.Equal(t, test.check[i].val, act.Value())
			}
		}

		list := cache.List()
		assert.Equal(t, int64(len(list)), int64(test.totalLen))
	})
}
