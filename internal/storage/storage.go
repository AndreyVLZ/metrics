package storage

import (
	"github.com/AndreyVLZ/metrics/internal/metric"
	"github.com/AndreyVLZ/metrics/internal/storage/postgres"
)

type Storage interface {
	Get(metric.MetricDB) (metric.MetricDB, error)
	Set(metric.MetricDB) error
	List() []metric.MetricDB
}

type Store struct {
	s         Storage
	postgres  *postgres.PostgresStore
	ifConnect bool
}

func (s *Store) Ping() bool {
	if err := s.postgres.Ping(); err != nil {
		return false
	}
	s.ifConnect = true
	return s.ifConnect
}

func NewStore(dbDNS string, s Storage) *Store {
	postgres := postgres.New(dbDNS)
	return &Store{
		s:        s,
		postgres: postgres,
	}
}

func (s *Store) Get(m metric.MetricDB) (metric.MetricDB, error) {
	if s.ifConnect {
		return s.postgres.Get(m)
	}

	return s.s.Get(m)
}
func (s *Store) Set(m metric.MetricDB) error {
	if s.ifConnect {
		return s.postgres.Set(m)
	}

	return s.s.Set(m)
}

func (s *Store) List() []metric.MetricDB {
	if s.ifConnect {
		return s.postgres.List()
	}

	return s.s.List()
}
