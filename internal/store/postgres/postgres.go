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
	NameConst = "postgres store"
)

var (
	errNotFind        = errors.New("not find")
	errDeltaNotValid  = errors.New("delta not valid")
	errValueNotValid  = errors.New("value not valid")
	errTypeNotSupport = errors.New("type not support")
)

const (
	getSQL          = "SELECT type_id,mname,delta,val FROM metric WHERE type_id=$1 AND mname=$2"
	setSQL          = "INSERT INTO metric (type_id,mname,delta,val) VALUES ($1,$2,$3,$4)"
	updSQL          = "UPDATE metric SET delta=$3, val=$4 WHERE type_id=$1 AND mname=$2"
	listSQL         = "SELECT type_id,mname,delta,val FROM metric"
	createTablesSQL = `
CREATE TABLE mettype (
	type_id integer,
	mtype varchar(20),
	PRIMARY KEY (type_id),
	UNIQUE(mtype)
);
CREATE TABLE metric (
	type_id integer NOT NULL REFERENCES mettype(type_id),
	mname varchar(50),
	delta bigint,
	val double precision,
	PRIMARY KEY (type_id,mname),
	UNIQUE(type_id,mname)
);`
)

// metricDB структура для сканирования из postgres.
type metricDB struct {
	Info  model.Info
	Delta sql.NullInt64
	Value sql.NullFloat64
}

// buildMetric возвращает модель метрики.
// Вовращает ошибку если:
// delta == nil,
// value == nil,
// тип не подерживается.
func (m metricDB) buildMetric() (model.Metric, error) {
	switch m.Info.MType {
	case model.TypeCountConst:
		if !m.Delta.Valid {
			return model.Metric{}, errDeltaNotValid
		}

		return model.NewCounterMetric(m.Info.MName, m.Delta.Int64), nil
	case model.TypeGaugeConst:
		if !m.Value.Valid {
			return model.Metric{}, errValueNotValid
		}

		return model.NewGaugeMetric(m.Info.MName, m.Value.Float64), nil
	default:
		return model.Metric{}, errTypeNotSupport
	}
}

type Config struct {
	ConnDB string
}

type Postgres struct {
	db  *sql.DB
	cfg Config
}

func New(cfg Config) *Postgres {
	return &Postgres{cfg: cfg}
}

func (s *Postgres) Name() string                 { return NameConst }
func (s *Postgres) Ping() error                  { return s.db.Ping() }
func (s *Postgres) Stop(_ context.Context) error { return s.db.Close() }

func (s *Postgres) Start(ctx context.Context) error {
	database, err := sql.Open("postgres", s.cfg.ConnDB)
	if err != nil {
		return fmt.Errorf("openDB [%s]: %w", s.cfg.ConnDB, err)
	}

	s.db = database

	if err := s.createTable(ctx); err != nil {
		log.Printf("create tables err: %v\n", err)
	}

	return nil
}

// Возвращает срез всех метрик из базы.
func (s *Postgres) List(ctx context.Context) ([]model.Metric, error) {
	var metDB metricDB

	arr := make([]model.Metric, 0)

	listStmt, err := s.db.PrepareContext(ctx, listSQL)
	if err != nil {
		return nil, fmt.Errorf("prepare listSQL: %w", err)
	}
	defer listStmt.Close()

	rows, err := listStmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("rowsErr: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		dest := []any{
			&metDB.Info.MType,
			&metDB.Info.MName,
			&metDB.Delta,
			&metDB.Value,
		}

		if errScan := rows.Scan(dest...); errScan != nil {
			return nil, fmt.Errorf("row scan: %w", errScan)
		}

		met, errBuild := metDB.buildMetric()
		if errBuild != nil {
			return nil, fmt.Errorf("%w", errBuild)
		}

		arr = append(arr, met)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rowsErr: %w", err)
	}

	return arr, nil
}

// AddBatch добавлеяет срез Metric в базу.
// Реализация в одной транзакции.
func (s *Postgres) AddBatch(ctx context.Context, arr []model.Metric) error {
	transaction, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := s.addBatchTx(ctx, transaction, arr); err != nil {
		if errRoll := transaction.Rollback(); errRoll != nil {
			err = errors.Join(err, fmt.Errorf("txRollback: %w", errRoll))
		}

		return fmt.Errorf("%w", err)
	}

	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("txCommit: %w", err)
	}

	return nil
}

func (s *Postgres) Get(ctx context.Context, mInfo model.Info) (model.Metric, error) {
	getStmt, err := s.db.PrepareContext(ctx, getSQL)
	if err != nil {
		return model.Metric{}, fmt.Errorf("prepare getSQL: %w", err)
	}

	return get(ctx, getStmt, mInfo)
}

func (s *Postgres) Update(ctx context.Context, met model.Metric) (model.Metric, error) {
	getStmt, err := s.db.PrepareContext(ctx, getSQL)
	if err != nil {
		return model.Metric{}, fmt.Errorf("prepare getSQL: %w", err)
	}

	setStmt, err := s.db.PrepareContext(ctx, setSQL)
	if err != nil {
		return model.Metric{}, fmt.Errorf("prepare setSQL: %w", err)
	}

	updStmt, err := s.db.PrepareContext(ctx, updSQL)
	if err != nil {
		return model.Metric{}, fmt.Errorf("prepare updSQL: %w", err)
	}

	return update(ctx, getStmt, setStmt, updStmt, met)
}

func (s *Postgres) addBatchTx(ctx context.Context, tx *sql.Tx, arr []model.Metric) error {
	getStmt, err := tx.PrepareContext(ctx, getSQL)
	if err != nil {
		return fmt.Errorf("prepare getSQL: %w", err)
	}
	defer getStmt.Close()

	setStmt, err := tx.PrepareContext(ctx, setSQL)
	if err != nil {
		return fmt.Errorf("preapare setSQL: %w", err)
	}
	defer setStmt.Close()

	updStmt, err := tx.PrepareContext(ctx, updSQL)
	if err != nil {
		return fmt.Errorf("prepare updSQL: %w", err)
	}
	defer updStmt.Close()

	for i := range arr {
		if _, err := update(ctx, getStmt, setStmt, updStmt, arr[i]); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	return nil
}

// Вызывает подготовленный запрос на получение метрики,
// если метрика существует в базе - обновляет метрику и вызывает поготовленный запрос на обновление метрики в базе,
// иначе вызывает подготовленный запрос на сохранение новой метрики.
func update(ctx context.Context, getStmt, setStmt, updStmt *sql.Stmt, met model.Metric) (model.Metric, error) {
	metRes, err := get(ctx, getStmt, met.Info)
	if err != nil {
		if errors.Is(err, errNotFind) { // set
			return upset(ctx, setStmt, met)
		}

		return model.Metric{}, fmt.Errorf("getErr: %w", err)
	}

	// update && set
	if err := metRes.Update(met.Value); err != nil {
		return model.Metric{}, fmt.Errorf("updErr: %w", err)
	}

	return upset(ctx, updStmt, metRes)
}

func upset(ctx context.Context, upsetStmt *sql.Stmt, met model.Metric) (model.Metric, error) {
	args := []any{
		met.MType,
		met.MName,
		met.Delta,
		met.Val,
	}

	if _, err := upsetStmt.ExecContext(ctx, args...); err != nil {
		return model.Metric{}, fmt.Errorf("upsetErr: %w", err)
	}

	return met, nil
}

func get(ctx context.Context, getStmt *sql.Stmt, mInfo model.Info) (model.Metric, error) {
	var metDB metricDB

	args := []any{
		mInfo.MType,
		mInfo.MName,
	}

	dest := []any{
		&metDB.Info.MType,
		&metDB.Info.MName,
		&metDB.Delta,
		&metDB.Value,
	}

	if err := getStmt.QueryRowContext(ctx, args...).Scan(dest...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Metric{}, errNotFind
		}

		return model.Metric{}, fmt.Errorf("%w", err)
	}

	met, err := metDB.buildMetric()
	if err != nil {
		return model.Metric{}, fmt.Errorf("%w", err)
	}

	return met, nil
}

// Создает необходимые таблицы в базе.
func (s *Postgres) createTable(ctx context.Context) error {
	isExist, err := s.checkTable(ctx, "metric")
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if isExist {
		return nil
	}

	if _, err := s.db.ExecContext(ctx, createTablesSQL); err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := s.addTypes(ctx); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

// Проверяет наличие таблицы nameTable.
func (s *Postgres) checkTable(ctx context.Context, nameTable string) (bool, error) {
	var n int64

	sqlCheckSQL := "select 1 from information_schema.tables where table_name =$1"

	if err := s.db.QueryRowContext(ctx, sqlCheckSQL, nameTable).Scan(&n); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("%w", err)
	}

	return true, nil
}

// Добавляет в таблицу поддерживамые типы метрик.
func (s *Postgres) addTypes(ctx context.Context) error {
	sqlAddTypeSQL := "INSERT INTO mettype (type_id,mtype) VALUES ($1,$2)"
	for mtype := model.TypeCountConst; mtype <= model.TypeGaugeConst; mtype++ {
		if _, err := s.db.ExecContext(ctx, sqlAddTypeSQL, mtype, mtype.String()); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	return nil
}
