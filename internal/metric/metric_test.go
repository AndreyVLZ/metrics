package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	tc := []struct {
		mType metricType
		want  string
	}{
		{
			mType: GaugeType,
			want:  "gauge",
		}, {
			mType: CounterType,
			want:  "counter",
		},
	}

	for _, test := range tc {
		assert.Equal(t, test.mType.String(), test.want)
	}
}

// Counter
func TestNewCounter(t *testing.T) {
	tc := []struct {
		valStr string
		want   Counter
		err    error
	}{
		{
			valStr: "123",
			want:   Counter(123),
			err:    nil,
		},
		/*
			{
				valStr: "123-",
				want:   Counter(0),
				err:    ErrStringIsNotValid,
			},
		*/
	}

	for _, test := range tc {
		ex, err := NewCounter(test.valStr)
		assert.Equal(t, err, test.err)
		assert.Equal(t, ex, &test.want)
	}
}

/*
func TestSetCounter(t *testing.T) {
	tc := []struct {
		counter Counter
		valStr  string
		want    Counter
		err     error
	}{
		{
			counter: Counter(5),
			valStr:  "5",
			want:    Counter(10),
			err:     nil,
		},
		{
			counter: Counter(5),
			valStr:  "5-",
			want:    Counter(0),
			err:     ErrStringIsNotValid,
		},
	}

	for _, test := range tc {

		err := test.counter.Set(test.valStr)
		assert.Equal(t, err, test.err)
	}
}

func TestSetValCounter(t *testing.T) {
	tc := []struct {
		counter Counter
		val     Counter
		want    Counter
	}{
		{
			counter: Counter(5),
			val:     Counter(5),
			want:    Counter(10),
		},
		{
			counter: Counter(5),
			val:     Counter(10),
			want:    Counter(15),
		},
	}

	for _, test := range tc {
		test.counter.SetVal(test.val)
		assert.Equal(t, test.counter, test.want)
	}
}
*/

func TestTypeCounter(t *testing.T) {
	assert.Equal(t, Counter(5).Type(), CounterType.String())
}

func TestStringCounter(t *testing.T) {
	tc := []struct {
		counter Counter
		want    string
	}{
		{
			counter: Counter(5),
			want:    "5",
		},
		{
			counter: Counter(15),
			want:    "15",
		},
	}

	for _, test := range tc {
		assert.Equal(t, test.counter.String(), test.want)
	}
}

// Gauge
func TestNewGauge(t *testing.T) {
	tc := []struct {
		valStr string
		want   Gauge
		err    error
	}{
		{
			valStr: "12.3",
			want:   Gauge(12.3),
			err:    nil,
		},
		/*
			{
				valStr: "12.3-",
				want:   Gauge(0),
				err:    ErrStringIsNotValid,
			},
		*/
	}

	for _, test := range tc {

		ex, err := NewGauge(test.valStr)

		assert.Equal(t, err, test.err)
		assert.Equal(t, ex, &test.want)
	}
}

/*
func TestSetGauge(t *testing.T) {
	tc := []struct {
		gauge  Gauge
		valStr string
		want   Gauge
		err    error
	}{
		{
			gauge:  Gauge(5.5),
			valStr: "10.10",
			want:   Gauge(10.10),
			err:    nil,
		},
		{
			gauge:  Gauge(5.5),
			valStr: "5-",
			want:   Gauge(0),
			err:    ErrStringIsNotValid,
		},
	}

	for _, test := range tc {
		err := test.gauge.Set(test.valStr)
		assert.Equal(t, err, test.err)
	}
}

func TestSetValGauge(t *testing.T) {
	tc := []struct {
		gauge Gauge
		val   Gauge
		want  Gauge
	}{
		{
			gauge: Gauge(5.5),
			val:   Gauge(5.10),
			want:  Gauge(5.10),
		},
		{
			gauge: Gauge(50),
			val:   Gauge(10),
			want:  Gauge(10),
		},
	}

	for _, test := range tc {
		test.gauge.SetVal(test.val)
		assert.Equal(t, test.gauge, test.want)
	}
}
*/

func TestTypeGauge(t *testing.T) {
	assert.Equal(t, Gauge(5.6).Type(), GaugeType.String())
}

func TestStringGauge(t *testing.T) {
	tc := []struct {
		gauge Gauge
		want  string
	}{
		{
			gauge: Gauge(5.5),
			want:  "5.5",
		},
		{
			gauge: Gauge(15.9),
			want:  "15.9",
		},
	}

	for _, test := range tc {
		assert.Equal(t, test.gauge.String(), test.want)
	}
}
