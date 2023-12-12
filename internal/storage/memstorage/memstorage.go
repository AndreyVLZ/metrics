package memstorage

import (
	"errors"
	"fmt"
	"sync"

	"github.com/AndreyVLZ/metrics/internal/metric"
	"github.com/AndreyVLZ/metrics/internal/storage"
)

var (
	ErrNotSupportedType    error = errors.New("not supported type")
	ErrValueByNameNotFound error = errors.New("value by name not found")
)

type MemStorage struct {
	gaugeRepo storage.Repository
	countRepo storage.Repository
}

func New(gRepo, cRepo storage.Repository) *MemStorage {
	return &MemStorage{
		gaugeRepo: gRepo,
		countRepo: cRepo,
	}
}

// Set Cохранение строкового значения метрики valStr по имени метрики name
// Возвращает ошибку если тип метрики typeStr не поддерживается
func (m *MemStorage) Set(typeStr, name, valStr string) error {
	switch typeStr {
	case metric.GaugeType.String():
		return m.setGauge(name, valStr)
	case metric.CounterType.String():
		return m.setCounter(name, valStr)
	default:
		return ErrNotSupportedType
	}
}
func (m *MemStorage) Get(typeStr, name string) (string, error) {
	switch typeStr {
	case metric.GaugeType.String():
		return m.getGauge(name)
	case metric.CounterType.String():
		return m.getCounter(name)
	default:
		return "", ErrNotSupportedType
	}
}

/*
func (m *MemStorage) List() map[string]string {
	gRepo:= m.gaugeRepo.List()
	cRepo:= m.countRepo.List()
	list:= make(map[string]string,len(gRepo)+len(cRepo))
	for k,v:=range gRepo{
		lis
	}
	return m.gaugeRepo
}
*/
// GaugeRepo Возвращает репозиторий для Gauge типа
func (m *MemStorage) GaugeRepo() storage.Repository { return m.gaugeRepo }

// CounterRepo Возвращает репозиторий для Counter типа
func (m *MemStorage) CounterRepo() storage.Repository { return m.countRepo }

// setGauge Cохранение строкового значения метрики valStr по имени метрики name для типа Gauge
func (m *MemStorage) setGauge(name, valStr string) error {
	return m.gaugeRepo.Set(name, valStr)
}

func (m *MemStorage) getGauge(name string) (string, error) {
	return m.gaugeRepo.Get(name)
}

// setCounter Cохранение строкового значения метрики valStr по имени метрики name для типа Counter
func (m *MemStorage) setCounter(name, valStr string) error {
	return m.countRepo.Set(name, valStr)
}
func (m *MemStorage) getCounter(name string) (string, error) {
	return m.countRepo.Get(name)
}

/*
func (m *MemStorage) Repo(mType metric.MetricType) (storage.Repository, error) {
	switch mType {
	case metric.CounterType:
		return m.countRepo, nil
	case metric.GaugeType:
		return m.gaugeRepo, nil
	default:
		return nil, errors.New("no Repo by mType")
	}
}
*/

// CounterRepo
type counterRepo struct {
	m     sync.Mutex
	store map[string]*metric.Counter
}

func NewCounterRepo() *counterRepo {
	return &counterRepo{
		store: make(map[string]*metric.Counter),
	}
}

// Set Сохранение строкового значения valStr по имени name
func (cr *counterRepo) Set(name, valStr string) error {
	c, err := metric.NewCounter(valStr)
	if err != nil {
		return err
	}

	cr.SetVal(name, c)

	return nil
}

// SetVal Сохранение значения val типа Counter по имени name
func (cr *counterRepo) SetVal(name string, valC metric.Counter) {
	cr.m.Lock()
	defer cr.m.Unlock()

	val, ok := cr.store[name]
	if ok {
		val.SetVal(valC)
		return
	}

	cr.store[name] = &valC
}

// Get Возвращает значение типа string по имени name и
// ошибку если значение отсутствует
func (cr *counterRepo) Get(name string) (string, error) {
	if el, ok := cr.store[name]; ok {
		return el.String(), nil
	}

	return "", ErrValueByNameNotFound
}

// List Возвращает набор сохраненых значений
func (cr *counterRepo) List() map[string]string {
	m := make(map[string]string, len(cr.store))

	for k, v := range cr.store {
		m[k] = v.String()
	}

	return m
}

// GaugeRepo
type gaugeRepo struct {
	m     sync.Mutex
	store map[string]*metric.Gauge
}

func NewGaugeRepo() *gaugeRepo {
	return &gaugeRepo{
		store: make(map[string]*metric.Gauge),
	}
}

// Set Сохранение строкового значения valStr по имени name
func (gr *gaugeRepo) Set(name, valStr string) error {
	g, err := metric.NewGauge(valStr)
	if err != nil {
		return err
	}

	gr.SetVal(name, g)

	return nil
}

// SetVal Сохранение значения val типа Gauge по имени name
func (gr *gaugeRepo) SetVal(name string, val metric.Gauge) {
	gr.m.Lock()
	defer gr.m.Unlock()

	valOld, ok := gr.store[name]
	if ok {
		valOld.SetVal(val)
		//gr.store[name] = val
		return
	}

	gr.store[name] = &val
}

// Get Возвращает значение типа string по имени name и
// ошибку если значение отсутствует
func (gr *gaugeRepo) Get(name string) (string, error) {
	if el, ok := gr.store[name]; ok {
		return el.String(), nil
	}

	return "", ErrValueByNameNotFound
}

// List Возвращает набор сохраненых значений
func (gr *gaugeRepo) List() map[string]string {
	arr := make(map[string]string, len(gr.store))

	for k, v := range gr.store {
		arr[k] = v.String()
	}

	return arr
}

// NOTE: DELETE
func (gr *gaugeRepo) Range() {
	for name, g := range gr.store {
		fmt.Printf("name: %s val: %s [%s]\n", name, g.String(), g.Type().String())
	}
}

func (cr *counterRepo) Range() {
	for name, g := range cr.store {
		fmt.Printf("name: %s val: %s [%s]\n", name, g.String(), g.Type().String())
	}
}
