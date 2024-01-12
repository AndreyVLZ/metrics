package metric

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

/*
type Namer interface {
	Name() string
}
*/

type Typer interface {
	Type() string
}

type ToReader interface {
	ReadTo(any) error
}

type Valuer interface {
	//Namer
	Typer
	ToReader
	fmt.Stringer
}

type MetricDB struct {
	name string
	Valuer
}

func NewMetricDB(name string, val Valuer) MetricDB {
	return MetricDB{
		name:   name,
		Valuer: val,
	}
}

/*
func NewGetMetrciDB(typeStr string, name string) (MetricDB, error) {
	var (
		val Valuer
	)

	switch typeStr {
	case CounterType.String():
		val = new(Counter)
	case GaugeType.String():
		val = new(Gauge)
	default:
		return MetricDB{}, errors.New("xz")
	}

	return NewMetricDB(name, val), nil
}

func NewUpdateMetricDB(typeStr string, name string, valStr string) (MetricDB, error) {
	var (
		val Valuer
		err error
	)

	switch typeStr {
	case CounterType.String():
		val, err = NewCounter(valStr)
		if err != nil {
			return MetricDB{}, err
		}
	case GaugeType.String():
		val, err = NewGauge(valStr)
		if err != nil {
			return MetricDB{}, err
		}
	default:
		return MetricDB{}, errors.New("xz")
	}

	return NewMetricDB(name, val), nil
}
*/

func (m MetricDB) Name() string { return m.name }

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

var _ Valuer = new(Counter)

type Counter int64

func (c Counter) ReadTo(val any) error {
	return readTo(
		reflect.ValueOf(val),
		reflect.Int64,
		func(rv reflect.Value) { rv.SetInt(int64(c)) },
	)
}

// NewCounter Возвращает Counter распарсенной из строки valStr
// и ошибку если valStr не валиден
func NewCounter(valStr string) (*Counter, error) {
	newInt, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		return nil, ErrStringIsNotValid
	}
	c := Counter(newInt)

	return &c, nil
}

/*
func NewCounter(valStr string) (Counter, error) {
	c := Counter(0)
	err := c.Set(valStr)

	return c, err
}
*/

// Set Установка нового значения из строки valStr для Counter
// Возвращает ошибку если valStr не валиден
/*
func (c *Counter) Set(valStr string) error {
	newInt, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		return ErrStringIsNotValid
	}

	c.SetVal(Counter(newInt))

	return nil
}
*/

// SetVal Установка нового значения newVal для текущего Counter
//func (c *Counter) SetVal(newVal Counter) { *c = *c + newVal }

// Type Возвращает константу CounterType типа metricType
func (c Counter) Type() string { return CounterType.String() }

// String возвращает строковое представление типа Counter
func (c Counter) String() string { return strconv.FormatInt(int64(c), 10) }

var _ Valuer = new(Gauge)

type Gauge float64

func (g Gauge) ReadTo(val any) error {
	return readTo(
		reflect.ValueOf(val),
		reflect.Float64,
		func(rv reflect.Value) { rv.SetFloat(float64(g)) },
	)
}

// NewGauge Возвращает Gauge распарсенной из строки valStr
// и ошибку если valStr не валиден
func NewGauge(valStr string) (*Gauge, error) {
	newGauge, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return nil, ErrStringIsNotValid
	}

	g := Gauge(newGauge)

	return &g, nil
}

/*
func NewGauge(valStr string) (Gauge, error) {
	g := Gauge(0)
	err := g.Set(valStr)

	return g, err
}
*/

// Set Установка нового значения из строки valStr для Gauge
// Возвращает ошибку если valStr не валиден
/*
func (g *Gauge) Set(valStr string) error {
	newGauge, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return ErrStringIsNotValid
	}

	g.SetVal(Gauge(newGauge))

	return nil
}
*/

// SetVal Установка нового значения newVal для текущего Gauge
//func (g *Gauge) SetVal(newVal Gauge) { *g = newVal }

// Type Возвращает константу GaugeType типа metricType
func (g Gauge) Type() string { return GaugeType.String() }

// String возвращает строковое представление типа Gauge
func (g Gauge) String() string { return strconv.FormatFloat(float64(g), 'f', -1, 64) }

// NOTE: доделать
func check(rv reflect.Value, rk reflect.Kind) error {
	if rv.Kind() != rk {
		return fmt.Errorf("value is not as %v", rk.String())
	}

	return nil
}

func readTo(rv reflect.Value, rk reflect.Kind, fn func(reflect.Value)) error {
	if err := check(rv, reflect.Pointer); err != nil {
		return err
	}

	el := rv.Elem()
	if err := check(el, rk); err != nil {
		return err
	}

	fn(el)

	return nil
}

func newValuer(typeStr, val string) (Valuer, error) {
	switch typeStr {
	case CounterType.String():
		return NewCounter(val)
	case GaugeType.String():
		return NewGauge(val)
	default:
		return nil, errors.New("typeStr not support")
	}
}

func URLParse(typeStr, name, valStr string) (MetricDB, error) {
	if name == "" {
		return MetricDB{}, errors.New("name not correct")
	}

	val, err := newValuer(typeStr, valStr)
	if err != nil {
		return MetricDB{}, err
	}

	metricDB := NewMetricDB(name, val)

	return metricDB, nil
}
