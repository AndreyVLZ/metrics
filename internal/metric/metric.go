package metric

import (
	"errors"
	"strconv"
)

var ErrStringIsNotValid error = errors.New("string is not valid")

// metricType Тип метрики
type metricType uint8

// Константы поддерживаемых типов метрик
const (
	GaugeType metricType = iota
	CounterType
)

// String Возвращает строковое представление типов
func (m metricType) String() string {
	return []string{"gauge", "counter"}[m]
}

type Counter uint64

// NewCounter Возвращает Counter распарсенной из строки valStr
// и ошибку если valStr не валиден
func NewCounter(valStr string) (Counter, error) {
	c := Counter(0)
	err := c.Set(valStr)

	return c, err
}

// Set Установка нового значения из строки valStr для Counter
// Возвращает ошибку если valStr не валиден
func (c *Counter) Set(valStr string) error {
	newInt, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		return ErrStringIsNotValid
	}

	c.SetVal(Counter(newInt))

	return nil
}

// SetVal Установка нового значения newVal для текущего Counter
func (c *Counter) SetVal(newVal Counter) { *c = *c + newVal }

// Type Возвращает константу CounterType типа metricType
func (c Counter) Type() metricType { return CounterType }

// String возвращает строковое представление типа Counter
func (c Counter) String() string { return strconv.FormatInt(int64(c), 10) }

type Gauge float64

// NewGauge Возвращает Gauge распарсенной из строки valStr
// и ошибку если valStr не валиден
func NewGauge(valStr string) (Gauge, error) {
	g := Gauge(0)
	err := g.Set(valStr)

	return g, err
}

// Set Установка нового значения из строки valStr для Gauge
// Возвращает ошибку если valStr не валиден
func (g *Gauge) Set(valStr string) error {
	newGauge, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return ErrStringIsNotValid
	}

	g.SetVal(Gauge(newGauge))

	return nil
}

// SetVal Установка нового значения newVal для текущего Gauge
func (g *Gauge) SetVal(newVal Gauge) { *g = newVal }

// Type Возвращает константу GaugeType типа metricType
func (g Gauge) Type() metricType { return GaugeType }

// String возвращает строковое представление типа Gauge
func (g Gauge) String() string { return strconv.FormatFloat(float64(g), 'f', -1, 64) }
