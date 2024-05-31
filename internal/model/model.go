package model

import (
	"errors"
	"fmt"
	"strconv"
)

var ErrNameEmpty = errors.New("name empty")

var errTypeNotSupport = errors.New("type not support")

// Структура метрики для http запросов и ответов.
type MetricJSON struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (m MetricJSON) String() string {
	switch m.MType {
	case TypeCountConst.String():
		return strconv.FormatInt(*m.Delta, 10)
	default:
		return strconv.FormatFloat(*m.Value, 'f', -1, 64)
	}
}

type Info struct {
	MName string
	MType Type
}

func ParseInfo(nameStr, typeStr string) (Info, error) {
	if nameStr == "" {
		return Info{}, ErrNameEmpty
	}

	mType, err := ParseType(typeStr)
	if err != nil {
		return Info{}, fmt.Errorf("%w", err)
	}

	return Info{MName: nameStr, MType: mType}, nil
}

func NewCounterInfo(mName string) Info { return Info{MName: mName, MType: TypeCountConst} }
func NewGaugeInfo(mName string) Info   { return Info{MName: mName, MType: TypeGaugeConst} }

type Value struct {
	Delta *int64
	Val   *float64
}

func NewCounterValue(delta int64) Value { return Value{Delta: &delta, Val: nil} }
func NewGaugeValue(val float64) Value   { return Value{Delta: nil, Val: &val} }

type Metric struct {
	Info
	Value
}

func NewMetric(info Info, value Value) Metric { return Metric{Info: info, Value: value} }

func NewCounterMetric(mName string, delta int64) Metric {
	return NewMetric(NewCounterInfo(mName), NewCounterValue(delta))
}

func NewGaugeMetric(mName string, val float64) Metric {
	return NewMetric(NewGaugeInfo(mName), NewGaugeValue(val))
}

func (m *Metric) GetValue() float64 { return *m.Val }
func (m *Metric) GetDelta() int64   { return *m.Delta }

func (m *Metric) Update(newVal Value) error {
	switch m.MType {
	case TypeCountConst:
		return m.updateDelta(newVal.Delta)
	case TypeGaugeConst:
		return m.updateValue(newVal.Val)
	}

	return errTypeNotSupport
}

func (m *Metric) updateDelta(newDelta *int64) error {
	if newDelta != nil {
		*m.Delta = *newDelta + *m.Delta

		return nil
	}

	return errors.New("delta is nil")
}

func (m *Metric) updateValue(newValue *float64) error {
	if newValue != nil {
		m.Val = newValue

		return nil
	}

	return errors.New("value is nil")
}

type InfoStr struct {
	Name  string
	MType string
}

/*
func (info InfoStr) Valid() error {
	if info.MType == "" {
		return errors.New("info.MType empty")
	}

	if info.Name == "" {
		return errors.New("info.Name empty")
	}

	return nil
}
*/

// Представление метрики в виде строк.
type MetricStr struct {
	InfoStr
	Val string
}

func BuildMetricJSON(met Metric) MetricJSON {
	return MetricJSON{
		ID:    met.MName,
		MType: met.MType.String(),
		Value: met.Val,
		Delta: met.Delta,
	}
}

func BuildArrMetricJSON(arrMet []Metric) []MetricJSON {
	arrMetJSON := make([]MetricJSON, len(arrMet))

	for i := range arrMet {
		arrMetJSON[i] = BuildMetricJSON(arrMet[i])
	}

	return arrMetJSON
}

func ParseMetricJSON(metStr MetricStr) (MetricJSON, error) {
	switch metStr.MType {
	case TypeCountConst.String():
		val, err := strconv.ParseInt(metStr.Val, 10, 64)
		if err != nil {
			return MetricJSON{}, fmt.Errorf("%w", err)
		}

		return MetricJSON{ID: metStr.Name, MType: metStr.MType, Delta: &val}, nil
	case TypeGaugeConst.String():
		val, err := strconv.ParseFloat(metStr.Val, 64)
		if err != nil {
			return MetricJSON{}, fmt.Errorf("%w", err)
		}

		return MetricJSON{ID: metStr.Name, MType: metStr.MType, Value: &val}, nil
	default:
		return MetricJSON{}, errTypeNotSupport
	}
}

func ParseMetric(met MetricJSON) (Metric, error) {
	var val Value

	info, err := ParseInfo(met.ID, met.MType)
	if err != nil {
		return Metric{}, fmt.Errorf("%w", err)
	}

	switch info.MType {
	case TypeCountConst:
		val = NewCounterValue(*met.Delta)
	default:
		val = NewGaugeValue(*met.Value)
	}

	return Metric{Info: info, Value: val}, nil
}

func BuildArrMetric(arr []MetricJSON) ([]Metric, error) {
	res := make([]Metric, len(arr))

	for i := range arr {
		met, err := ParseMetric(arr[i])
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		res[i] = met
	}

	return res, nil
}
