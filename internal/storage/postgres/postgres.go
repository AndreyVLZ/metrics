package postgres

import (
	"database/sql"

	"github.com/AndreyVLZ/metrics/internal/metric"
	_ "github.com/lib/pq"
)

type PostgresStore struct {
	psqlConn string
	db       *sql.DB
}

func New(psqlConn string) *PostgresStore {
	return &PostgresStore{
		psqlConn: psqlConn,
	}

}

func (s *PostgresStore) Start() error {
	return s.open()
}

func (s *PostgresStore) Stop() error {
	return s.db.Close()
}

func (s *PostgresStore) open() error {
	db, err := sql.Open("postgres", s.psqlConn)
	if err != nil {
		return err
	}
	s.db = db

	return nil
}

func (s *PostgresStore) Ping() error {
	err := s.open()
	if err != nil {
		return err
	}
	return s.db.Ping()
}

func (s *PostgresStore) Get(m metric.MetricDB) (metric.MetricDB, error) {
	return metric.MetricDB{}, nil
}

func (s *PostgresStore) Set(m metric.MetricDB) error {
	return nil
}

func (s *PostgresStore) List() []metric.MetricDB {
	return nil
}
