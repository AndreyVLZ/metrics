package memstorage

import (
	"context"
	"log"
	"testing"

	"github.com/AndreyVLZ/metrics/internal/metric"
	"github.com/stretchr/testify/assert"
)

type valByType struct {
	metric.Valuer
}

func (v valByType) Type() string {
	return "not counter or gayge"
}

func TestSetMemStore(t *testing.T) {

	type fnSet func(string) metric.Valuer

	type metricdb struct {
		name   string
		valStr string
		fn     fnSet
	}

	tc := []struct {
		tName  string
		metric metricdb
		err    error
	}{
		{
			tName: "positive #1",
			metric: metricdb{
				name:   "counter",
				valStr: "123",
				fn: func(str string) metric.Valuer {
					valuer, _ := metric.NewCounter(str)
					return valuer
				},
			},
			err: nil,
		},
		{
			tName: "positive #2",
			metric: metricdb{
				name:   "gauge",
				valStr: "123.123",
				fn: func(str string) metric.Valuer {
					valuer, _ := metric.NewGauge(str)
					return valuer
				},
			},
			err: nil,
		},
		{
			tName: "negative #3",
			metric: metricdb{
				name:   "not counter or gauge",
				valStr: "",
				fn: func(str string) metric.Valuer {
					return valByType{metric.Counter(0)}
				},
			},
			err: ErrNotSupportedType,
		},
	}

	for _, test := range tc {
		t.Run(test.tName, func(t *testing.T) {
			metricNew, err := New().Set(
				context.Background(),
				metric.NewMetricDB(
					test.metric.name,
					test.metric.fn(test.metric.valStr),
				),
			)
			assert.Equal(t, err, test.err)
			if test.err == nil {
				log.Printf("metricNew %v\n", metricNew)
				assert.Equal(t, metricNew.Name(), test.metric.name)
				assert.Equal(t, metricNew.String(), test.metric.valStr)
			}
		})
	}
}
