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

func (r *counterRepo) Get(name string) (metric.Valuer, error) {
	r.m.RLock()
	defer r.m.RUnlock()

	valCounter, ok := r.repo[name]
	if !ok {
		return nil, ErrValueByNameNotFound
	}

	return valCounter, nil
}

func (r *counterRepo) List() map[string]metric.Valuer {
	m := make(map[string]metric.Valuer, len(r.repo))
	for i := range r.repo {
		m[i] = r.repo[i]
	}

	return m
}

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

func (r *gaugeRepo) Get(name string) (metric.Valuer, error) {
	r.m.RLock()
	defer r.m.RUnlock()

	valGauge, ok := r.repo[name]
	if !ok {
		return nil, ErrValueByNameNotFound
	}

	return valGauge, nil
}

func (r *gaugeRepo) List() map[string]metric.Valuer {
	m := make(map[string]metric.Valuer, len(r.repo))
	for i := range r.repo {
		m[i] = r.repo[i]
	}

	return m
}
