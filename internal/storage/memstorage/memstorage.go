package memstorage

import (
	"context"
	"errors"
	"sync"

	"github.com/AndreyVLZ/metrics/internal/metric"
)

type MemStorageConfig struct {
}

var (
	ErrNotSupportedType    error = errors.New("not supported type")
	ErrValueByNameNotFound error = errors.New("value by name not found")
)

type Repository interface {
	Get(string) (metric.Valuer, error)
	Set(string, metric.Valuer) error
	Update(string, metric.Valuer) error
	List() map[string]metric.Valuer
}

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

func (s *MemStorage) Open() error { return nil }

func (s *MemStorage) SetBatch(ctx context.Context, arr []metric.MetricDB) error {
	for i := range arr {
		_, err := s.Set(ctx, arr[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *MemStorage) UpdateBatch(ctx context.Context, arr []metric.MetricDB) error {
	for i := range arr {
		_, err := s.Update(ctx, arr[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (s MemStorage) Update(ctx context.Context, metricNew metric.MetricDB) (metric.MetricDB, error) {
	switch metricNew.Type() {
	case metric.GaugeType.String():
		if err := s.gRepo.Update(metricNew.Name(), metricNew.Valuer); err != nil {
			return metric.MetricDB{}, err
		}
		return s.Get(ctx, metricNew)
	case metric.CounterType.String():
		if err := s.cRepo.Update(metricNew.Name(), metricNew.Valuer); err != nil {
			return metric.MetricDB{}, err
		}
		return s.Get(ctx, metricNew)
	default:
		return metric.MetricDB{}, ErrNotSupportedType
	}
}

func (s MemStorage) Set(ctx context.Context, metricNew metric.MetricDB) (metric.MetricDB, error) {
	switch metricNew.Type() {
	case metric.GaugeType.String():
		if err := s.gRepo.Set(metricNew.Name(), metricNew.Valuer); err != nil {
			return metric.MetricDB{}, err
		}
		return s.Get(ctx, metricNew)
	case metric.CounterType.String():
		if err := s.cRepo.Set(metricNew.Name(), metricNew.Valuer); err != nil {
			return metric.MetricDB{}, err
		}
		return s.Get(ctx, metricNew)
	default:
		return metric.MetricDB{}, ErrNotSupportedType
	}
}

func (s MemStorage) set(ctx context.Context, metricNew metric.MetricDB, upSet bool) (metric.MetricDB, error) {
	switch metricNew.Type() {
	case metric.GaugeType.String():
		if err := s.gRepo.Set(metricNew.Name(), metricNew.Valuer); err != nil {
			return metric.MetricDB{}, err
		}
		return s.Get(ctx, metricNew)
	case metric.CounterType.String():
		if err := s.cRepo.Set(metricNew.Name(), metricNew.Valuer); err != nil {
			return metric.MetricDB{}, err
		}
		return s.Get(ctx, metricNew)
	default:
		return metric.MetricDB{}, ErrNotSupportedType
	}
}

func (s MemStorage) Get(ctx context.Context, metricNew metric.MetricDB) (metric.MetricDB, error) {
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

func (s MemStorage) List(context.Context) []metric.MetricDB {
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

func (s MemStorage) Ping() error { return nil }

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

func (r *counterRepo) Update(name string, val metric.Valuer) error {
	var valCounter metric.Counter
	err := val.ReadTo(&valCounter)
	if err != nil {
		return err
	}

	r.set(name, valCounter, true)

	return nil
}

func (r *counterRepo) Set(name string, val metric.Valuer) error {
	var valCounter metric.Counter
	err := val.ReadTo(&valCounter)
	if err != nil {
		return err
	}

	r.set(name, valCounter, false)
	return nil
}

func (r *counterRepo) set(name string, val metric.Counter, upSet bool) {
	r.m.Lock()
	defer r.m.Unlock()

	_, ok := r.repo[name]
	if ok && upSet {
		r.repo[name] += val
		return
	}

	r.repo[name] = val
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

func (r *gaugeRepo) Update(name string, val metric.Valuer) error {
	return r.Set(name, val)
}

func (r *gaugeRepo) Set(name string, val metric.Valuer) error {
	var valGauge metric.Gauge
	err := val.ReadTo(&valGauge)
	if err != nil {
		return err
	}

	r.set(name, valGauge)
	return nil
}

func (r *gaugeRepo) set(name string, val metric.Gauge) {
	r.m.Lock()
	defer r.m.Unlock()

	r.repo[name] = val
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
