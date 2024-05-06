package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/AndreyVLZ/metrics/internal/model"
	_ "github.com/lib/pq"
)

const (
	NameConst        = "postgres store"
	nameTableCounter = "counter"
	nameTableGauge   = "gauge"
)

const (
	getSQL       = "SELECT name,type_id,val FROM %s WHERE name=$1"
	setSQL       = "INSERT INTO %s (name,val) VALUES ($1,$2)"
	updateSQL    = "UPDATE %s SET val = $2 WHERE name = $1"
	listSQL      = "SELECT c.name,c.type_id,c.val FROM %s c LEFT JOIN mtype mt USING (type_id)"
	createTables = `
CREATE TABLE IF NOT EXISTS mtype (
	type_id integer,
	name_type varchar(10),
	PRIMARY KEY (type_id),
	UNIQUE(name_type)
);
CREATE TABLE IF NOT EXISTS counter (
	type_id integer NOT NULL DEFAULT 1 REFERENCES mtype(type_id),
	name varchar(50) NOT NULL,
	val bigint,
	UNIQUE (name),
	PRIMARY KEY (name)
);
CREATE TABLE IF NOT EXISTS gauge (
	type_id integer NOT NULL DEFAULT 2 REFERENCES mtype(type_id),
	name varchar(50) NOT NULL,
	val double precision,
	UNIQUE (name),
	PRIMARY KEY (name)
);`
)

type DBTX interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

type Config struct {
	ConnDB string
}

type Postgres struct {
	cfg   Config
	db    *sql.DB
	cRepo Repo[int64]
	gRepo Repo[float64]
}

func New(cfg Config) *Postgres {
	return &Postgres{
		cfg:   cfg,
		cRepo: newRepo[int64](nameTableCounter),
		gRepo: newRepo[float64](nameTableGauge),
		// getSQL:    fmt.Sprintf(getSQL, nameTableCounter),
		// setSQL:    fmt.Sprintf(setSQL, nameTableCounter),
		// updateSQL: fmt.Sprintf(updateSQL, nameTableCounter),
		// listSQL:   fmt.Sprintf(listSQL, nameTableCounter),
		//
		// getSQL:    fmt.Sprintf(getSQL, nameTableGauge),
		// setSQL:    fmt.Sprintf(setSQL, nameTableGauge),
		// updateSQL: fmt.Sprintf(updateSQL, nameTableGauge),
		// listSQL:   fmt.Sprintf(listSQL, nameTableGauge),
	}
}

func (s *Postgres) Name() string                 { return NameConst }
func (s *Postgres) Stop(_ context.Context) error { return s.db.Close() }
func (s *Postgres) Ping() error                  { return s.db.Ping() }

func (s *Postgres) setDatabase(db *sql.DB) {
	s.db = db
	s.cRepo.db = db
	s.gRepo.db = db
}

func (s *Postgres) Start(ctx context.Context) error {
	database, err := sql.Open("postgres", s.cfg.ConnDB)
	if err != nil {
		return fmt.Errorf("openDB [%s]: %w", s.cfg.ConnDB, err)
	}

	s.setDatabase(database)

	if err := s.createTable(ctx); err != nil {
		log.Printf("tables: %w", err)
	}

	return nil
}

func (s *Postgres) GetCounter(ctx context.Context, name string) (model.MetricRepo[int64], error) {
	return s.cRepo.Get(ctx, name)
}

func (s *Postgres) GetGauge(ctx context.Context, name string) (model.MetricRepo[float64], error) {
	return s.gRepo.Get(ctx, name)
}

func (s *Postgres) UpdateCounter(ctx context.Context, met model.MetricRepo[int64]) (model.MetricRepo[int64], error) {
	return s.cRepo.Update(ctx, met)
}

func (s *Postgres) UpdateGauge(ctx context.Context, met model.MetricRepo[float64]) (model.MetricRepo[float64], error) {
	return s.gRepo.Update(ctx, met)
}

func (s *Postgres) List(ctx context.Context) (model.Batch, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Batch{}, fmt.Errorf("txBegin: %w", err)
	}

	batch, err := s.listTx(ctx, tx)
	if err != nil {
		if errRoll := tx.Rollback(); errRoll != nil {
			err = errors.Join(err, fmt.Errorf("txRollback: %w", errRoll))
		}

		return model.Batch{}, fmt.Errorf("list: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return model.Batch{}, fmt.Errorf("txCommit: %w", err)
	}

	return batch, nil
}

func (s *Postgres) listTx(ctx context.Context, dbtx DBTX) (model.Batch, error) {
	cList, err := s.cRepo.list(ctx, dbtx)
	if err != nil {
		return model.Batch{}, fmt.Errorf("crepo list: %w", err)
	}

	gList, err := s.gRepo.list(ctx, dbtx)
	if err != nil {
		return model.Batch{}, fmt.Errorf("grepo list: %w", err)
	}

	return model.Batch{CList: cList, GList: gList}, nil
}

func (s *Postgres) AddBatch(ctx context.Context, batch model.Batch) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("tx: %w", err)
	}

	if err := s.addBatchTx(ctx, tx, batch); err != nil {
		if errRoll := tx.Rollback(); errRoll != nil {
			err = errors.Join(err, fmt.Errorf("txRollback: %w", errRoll))
		}

		return fmt.Errorf("adder: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("txCommit: %w", err)
	}

	return nil
}

func (s *Postgres) addBatchTx(ctx context.Context, dbtx DBTX, batch model.Batch) error {
	if err := s.cRepo.addList(ctx, dbtx, batch.CList); err != nil {
		return fmt.Errorf("crepo adder list: %w", err)
	}

	if err := s.gRepo.addList(ctx, dbtx, batch.GList); err != nil {
		return fmt.Errorf("grepo adder list: %w", err)
	}

	return nil
}

func (s *Postgres) createTable(ctx context.Context) error {
	setTypeSQL := "INSERT INTO mtype (type_id, name_type) VALUES ($1,$2) ON CONFLICT (name_type) DO NOTHING"

	if _, err := s.db.ExecContext(ctx, createTables); err != nil {
		return fmt.Errorf("create table: %w", err)
	}

	stmt, err := s.db.PrepareContext(ctx, setTypeSQL)
	if err != nil {
		return fmt.Errorf("prepare tableStmt: %w", err)
	}
	defer stmt.Close()

	arrType := []model.TypeMetric{model.TypeCountConst, model.TypeGaugeConst}
	for i := range arrType {
		args := []any{arrType[i], arrType[i].String()}
		if _, err := stmt.ExecContext(ctx, args...); err != nil {
			return fmt.Errorf("add type metric: %w", err)
		}
	}

	return nil
}
