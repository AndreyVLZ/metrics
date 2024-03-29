package metric

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

type Typer interface {
	Type() string
}

type ToReader interface {
	ReadTo(any) error
}

type Valuer interface {
	Typer
	ToReader
	fmt.Stringer
}

var _ json.Marshaler = MetricDB{}

type MetricDB struct {
	name string
	Valuer
}

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (m MetricDB) MarshalJSON() ([]byte, error) {
	type metricAlias MetricDB
	metricJSON := Metric{
		ID:    m.Name(),
		MType: m.Type(),
	}

	if m.Type() == CounterType.String() {
		val := new(int64)
		err := m.ReadTo(val)
		if err != nil {
			return []byte{}, err
		}
		metricJSON.Delta = val
	}

	if m.Type() == GaugeType.String() {
		val := new(float64)
		err := m.ReadTo(val)
		if err != nil {
			return []byte{}, err
		}
		metricJSON.Value = val
	}

	return json.Marshal(metricJSON)
}

func (m *MetricDB) UnmarshalJSON(data []byte) error {
	type metricAlias MetricDB
	metricJSON := Metric{}

	err := json.Unmarshal(data, &metricJSON)
	if err != nil {
		return err
	}

	m.name = metricJSON.ID
	if metricJSON.MType == CounterType.String() && metricJSON.Delta != nil {
		cVal := Counter(*metricJSON.Delta)
		m.Valuer = cVal
	}
	if metricJSON.MType == GaugeType.String() && metricJSON.Value != nil {
		fVal := Gauge(*metricJSON.Value)
		m.Valuer = fVal
	}

	return nil
}

func NewMetricDB(name string, val Valuer) MetricDB {
	return MetricDB{
		name:   name,
		Valuer: val,
	}
}

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
