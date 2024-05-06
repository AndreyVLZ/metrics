package model

import (
	"errors"
	"fmt"
	_ "net/http/pprof"
	"strconv"
)

var errTypeNotSupport = errors.New("type not support")

type TypeMetric int8

func (tm TypeMetric) String() string { return supportTypeMetric()[tm] }

const (
	TypeNoCorrect  TypeMetric = iota
	TypeCountConst            // 1
	TypeGaugeConst            // 2
)

func supportTypeMetric() [3]string {
	return [3]string{
		"no correct",
		"counter",
		"gauge",
	}
}

type ValueType interface {
	int64 | float64
}

type MetricRepo[VT ValueType] struct {
	name  string
	mType TypeMetric
	val   VT
}

func NewMetricRepo[VT ValueType](name string, mType TypeMetric, val VT) MetricRepo[VT] {
	return MetricRepo[VT]{name: name, mType: mType, val: val}
}

func (m MetricRepo[_]) Type() string { return m.mType.String() }
func (m MetricRepo[_]) Name() string { return m.name }
func (m MetricRepo[VT]) Value() VT   { return m.val }

func (m *MetricRepo[VT]) Update(newVal VT) {
	switch m.mType {
	case TypeCountConst:
		m.val += newVal
	case TypeGaugeConst:
		m.val = newVal
	}
}

// Структура метрики для http запросов и ответов.
type MetricJSON struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

// func (m MetricJSON) Valid() error { return nil }

func (m MetricJSON) String() string {
	switch m.MType {
	case TypeCountConst.String():
		return strconv.FormatInt(*m.Delta, 10)
	case TypeGaugeConst.String():
		return strconv.FormatFloat(*m.Value, 'f', -1, 64)
	default:
		return TypeNoCorrect.String()
	}
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

type Info struct {
	Name  string
	MType string
}

func (info Info) Valid() error {
	if info.MType == "" {
		return errors.New("info.MType empty")
	}

	if info.Name == "" {
		return errors.New("info.Name empty")
	}

	return nil
}

// Представление метрики в виде строк.
type MetricStr struct {
	Info
	Val string
}

type Batch struct {
	CList []MetricRepo[int64]
	GList []MetricRepo[float64]
}

func (b *Batch) ToArrMetricJSON() []MetricJSON {
	arr := make([]MetricJSON, 0, len(b.CList)+len(b.GList))

	for i := range b.CList {
		met := b.CList[i]
		arr = append(arr, MetricJSON{ID: met.name, MType: met.Type(), Delta: &met.val, Value: nil})
	}

	for i := range b.GList {
		met := b.GList[i]
		arr = append(arr, MetricJSON{ID: met.name, MType: met.Type(), Value: &met.val, Delta: nil})
	}

	return arr
}
