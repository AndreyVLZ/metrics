package memstorage

import (
	"errors"
	"sync"

	"github.com/AndreyVLZ/metrics/internal/metric"
	"github.com/AndreyVLZ/metrics/internal/storage"
)

var (
	ErrNotSupportedType    error = errors.New("not supported type")
	ErrValueByNameNotFound error = errors.New("value by name not found")
)

type Repository interface {
	Set(string, metric.Valuer) error
	Get(string) (metric.Valuer, error)
	List() map[string]metric.Valuer
}

var _ storage.Storage = MemStorage{}

type MemStorage struct {
	gRepo Repository
	cRepo Repository
}

func New() *MemStorage {
	return &MemStorage{
		gRepo: NewGaugeRepo(),
		cRepo: NewCounterRepo(),
	}
}

func (s MemStorage) Set(metricNew metric.MetricDB) error {
	switch metricNew.Type() {
	case metric.GaugeType.String():
		return s.gRepo.Set(metricNew.Name(), metricNew.Valuer)
	case metric.CounterType.String():
		return s.cRepo.Set(metricNew.Name(), metricNew.Valuer)
	default:
		return ErrNotSupportedType
	}
}

/*
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
*/

func (s MemStorage) Get(metricNew metric.MetricDB) (metric.MetricDB, error) {
	var (
		val metric.Valuer
		err error
	)

	switch metricNew.Type() {
	case metric.CounterType.String():
		val, err = s.cRepo.Get(metricNew.Name())
		if err != nil {
			return metric.MetricDB{}, err
		}
	case metric.GaugeType.String():
		val, err = s.gRepo.Get(metricNew.Name())
		if err != nil {
			return metric.MetricDB{}, err
		}
	default:
		return metric.MetricDB{}, ErrNotSupportedType
	}

	metricNew.Valuer = val

	return metricNew, nil
}

func (s MemStorage) List() []metric.MetricDB {
	cMap := s.cRepo.List()
	gMap := s.gRepo.List()

	arr := make([]metric.MetricDB, 0, len(cMap)+len(gMap))

	for i := range cMap {
		arr = append(arr, metric.NewMetricDB(i, cMap[i]))
	}

	for i := range gMap {
		arr = append(arr, metric.NewMetricDB(i, gMap[i]))
	}

	return arr
}

/*
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
*/

// GaugeRepo Возвращает репозиторий для Gauge типа
//func (m *MemStorage) GaugeRepo() storage.Repository { return m.gaugeRepo }

// CounterRepo Возвращает репозиторий для Counter типа
//func (m *MemStorage) CounterRepo() storage.Repository { return m.countRepo }

/*
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
*/

// CounterRepo

var _ Repository = &counterRepo{}

type counterRepo struct {
	m    sync.RWMutex
	repo map[string]metric.Counter
}

func NewCounterRepo() *counterRepo {
	return &counterRepo{
		repo: make(map[string]metric.Counter),
	}
}

func (r *counterRepo) Set(name string, val metric.Valuer) error {
	var valCounter metric.Counter
	err := val.ReadTo(&valCounter)
	if err != nil {
		return err
	}

	return r.set(name, valCounter)
}

func (r *counterRepo) set(name string, val metric.Counter) error {
	r.m.Lock()
	defer r.m.Unlock()

	valNew := val

	valOld, ok := r.repo[name]
	if ok {
		valNew += valOld
	}

	r.repo[name] = valNew

	return nil
}

/*
// Set Сохранение строкового значения valStr по имени name
func (cr *counterRepo) Set(name, valStr string) error {
	c, err := metric.NewCounter(valStr)
	if err != nil {
		return err
	}

	cr.SetVal(name, c)

	return nil
}
*/

/*
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
*/

func (r *counterRepo) Get(name string) (metric.Valuer, error) {
	r.m.RLock()
	defer r.m.RUnlock()

	valCounter, ok := r.repo[name]
	if !ok {
		return nil, ErrValueByNameNotFound
	}

	return valCounter, nil
}

/*
// Get Возвращает значение типа string по имени name и
// ошибку если значение отсутствует
func (cr *counterRepo) Get(name string) (string, error) {
	if el, ok := cr.store[name]; ok {
		return el.String(), nil
	}

	return "", ErrValueByNameNotFound
}
*/

func (r *counterRepo) List() map[string]metric.Valuer {
	m := make(map[string]metric.Valuer, len(r.repo))
	for i := range r.repo {
		m[i] = r.repo[i]
	}

	return m
}

/*
// List Возвращает набор сохраненых значений
func (cr *counterRepo) List() map[string]string {
	m := make(map[string]string, len(cr.store))

	for k, v := range cr.store {
		m[k] = v.String()
	}

	return m
}
*/

// GaugeRepo
var _ Repository = &gaugeRepo{}

type gaugeRepo struct {
	m    sync.RWMutex
	repo map[string]metric.Gauge
}

func NewGaugeRepo() *gaugeRepo {
	return &gaugeRepo{
		repo: make(map[string]metric.Gauge),
	}
}

func (r *gaugeRepo) Set(name string, val metric.Valuer) error {
	var valGauge metric.Gauge
	err := val.ReadTo(&valGauge)
	if err != nil {
		return err
	}

	return r.set(name, valGauge)
}

func (r *gaugeRepo) set(name string, val metric.Gauge) error {
	r.m.Lock()
	defer r.m.Unlock()

	r.repo[name] = val
	return nil
}

/*
// Set Сохранение строкового значения valStr по имени name
func (gr *gaugeRepo) Set(name, valStr string) error {
	g, err := metric.NewGauge(valStr)
	if err != nil {
		return err
	}

	gr.SetVal(name, g)

	return nil
}
*/

/*
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
*/

func (r *gaugeRepo) Get(name string) (metric.Valuer, error) {
	r.m.RLock()
	defer r.m.RUnlock()

	valGauge, ok := r.repo[name]
	if !ok {
		return nil, ErrValueByNameNotFound
	}

	return valGauge, nil
}

/*
// Get Возвращает значение типа string по имени name и
// ошибку если значение отсутствует
func (gr *gaugeRepo) Get(name string) (string, error) {
	if el, ok := gr.store[name]; ok {
		return el.String(), nil
	}

	return "", ErrValueByNameNotFound
}
*/

func (r *gaugeRepo) List() map[string]metric.Valuer {
	m := make(map[string]metric.Valuer, len(r.repo))
	for i := range r.repo {
		m[i] = r.repo[i]
	}

	return m
}

/*
// List Возвращает набор сохраненых значений
func (gr *gaugeRepo) List() map[string]string {
	arr := make(map[string]string, len(gr.store))

	for k, v := range gr.store {
		arr[k] = v.String()
	}

	return arr
}
*/
