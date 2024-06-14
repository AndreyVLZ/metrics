package model

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrNameEmpty      = errors.New("name empty")
	ErrTypeNotSupport = errors.New("type not support")
	errDeltaNil       = errors.New("delta is nil")
	errValueNil       = errors.New("value is nil")
)

// MetricJSON структура метрики для http запросов и ответов.
type MetricJSON struct {
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
}

// String возвращает троковое представления значения метрики.
func (m MetricJSON) String() string {
	switch m.MType {
	case TypeCountConst.String():
		return strconv.FormatInt(*m.Delta, 10)
	case TypeGaugeConst.String():
		return strconv.FormatFloat(*m.Value, 'f', -1, 64)
	default:
		return ErrTypeNotSupport.Error()
	}
}

// Info хранит строку с именем метрики и идентификатор типа метрики.
type Info struct {
	MName string
	MType Type
}

// ParseInfo парсинг имени и типа метрики.
// Ошибка если имя =="" или тип не поддерживается.
func ParseInfo(nameStr, typeStr string) (Info, error) {
	if nameStr == "" {
		return Info{}, ErrNameEmpty
	}

	mType, err := ParseType(typeStr)
	if err != nil {
		return Info{}, fmt.Errorf("parseType: %w", err)
	}

	return Info{MName: nameStr, MType: mType}, nil
}

// Value хранит значения для метрики.
type Value struct {
	Delta *int64
	Val   *float64
}

// updateDelta обновляет Delta новым значение newDelta.
func (v *Value) updateDelta(newDelta *int64) error {
	if newDelta != nil {
		*v.Delta = *newDelta + *v.Delta

		return nil
	}

	return errDeltaNil
}

// updateValue устанавливает новое значением newValue для Value.
func (v *Value) updateValue(newValue *float64) error {
	if newValue != nil {
		v.Val = newValue

		return nil
	}

	return errValueNil
}

// Metric хранит Info и Value метрики.
type Metric struct {
	Value
	Info
}

// NewMetric возвращает новую метрику из структур Info и Value.
func NewMetric(info Info, value Value) Metric { return Metric{Info: info, Value: value} }

// NewCounterMetric возвращает метрику с имене mName и значением delta типа counter.
func NewCounterMetric(mName string, delta int64) Metric {
	return NewMetric(Info{MName: mName, MType: TypeCountConst}, Value{Delta: &delta, Val: nil})
}

// NewGaugeMetric возвращает метрику с имене mName и значением val типа gauge.
func NewGaugeMetric(mName string, val float64) Metric {
	return NewMetric(Info{MName: mName, MType: TypeGaugeConst}, Value{Delta: nil, Val: &val})
}

// Обновляет метрику новым значением newVal.
// Ошибка если delta или value == nil.
func (m *Metric) Update(newVal Value) error {
	switch m.MType {
	case TypeCountConst:
		return m.Value.updateDelta(newVal.Delta)
	case TypeGaugeConst:
		return m.Value.updateValue(newVal.Val)
	}

	return ErrTypeNotSupport
}

// InfoStr хранит строки с именем и типом метрики.
type InfoStr struct {
	Name  string
	MType string
}

// MetricStr хранит InfoStr и строку со значением метрики.
type MetricStr struct {
	InfoStr
	Val string
}

// BuildMetricJSON MetricJSON из структуры Metric.
func BuildMetricJSON(met Metric) MetricJSON {
	return MetricJSON{
		ID:    met.MName,
		MType: met.MType.String(),
		Value: met.Val,
		Delta: met.Delta,
	}
}

// BuildArrMetricJSON возвращает массив MetricJSON из массива Metric.
func BuildArrMetricJSON(arrMet []Metric) []MetricJSON {
	arrMetJSON := make([]MetricJSON, len(arrMet))

	for i := range arrMet {
		arrMetJSON[i] = BuildMetricJSON(arrMet[i])
	}

	return arrMetJSON
}
