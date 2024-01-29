package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/AndreyVLZ/metrics/internal/metric"
)

const listMetricSQL string = `
SELECT 
	m.type_name,
	m.name,
	m.cval,
	m.gval
FROM metric m
ORDER BY m.name
`

const getMetricSQL string = `
SELECT
	m.type_name,
	m.name,
	m.cval,
	m.gval
FROM metric m
WHERE m.name=$1 AND m.type_name=$2 
`

const updateMetricSQL string = `
INSERT INTO metric (type_name,name,cval,gval) VALUES
($1,$2,$3,$4)
ON CONFLICT (name,type_name) DO UPDATE
	SET (cval,gval) = (EXCLUDED.cval+metric.cval,EXCLUDED.gval)
RETURNING type_name,name,cval,gval
`

const setMetricSQL string = `
INSERT INTO metric AS met (type_name,name,cval,gval) VALUES 
($1,$2,$3,$4)
ON CONFLICT (name,type_name) DO UPDATE
	SET (cval,gval) = (met.cval,met.gval)
RETURNING type_name,name,cval,gval
`

const createTablesSQL string = `
DROP TABLE IF EXISTS metric;
CREATE TABLE IF NOT EXISTS metric (
	type_name varchar(10) NOT NULL,
	name varchar(100) NOT NULL,
	cval bigint,
	gval double precision,
	UNIQUE (name,type_name),
	primary key (name,type_name)
)`

type DBTX interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)

	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

type Query struct {
	db DBTX
}

func NewQuery(db *sql.DB) *Query {
	return &Query{
		db: db,
	}
}

func NewQueryWithTX(db *sql.Tx) *Query {
	return &Query{
		db: db,
	}
}
func (store *Postgres) exec(ctx context.Context, fn func(q *Query) error) error {
	return fn(NewQuery(store.db))
}

func (store *Postgres) execTx(ctx context.Context, fn func(q *Query) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := NewQueryWithTX(tx)

	err = fn(q)
	if err != nil {
		fmt.Printf("Отмена транзакции %v\n", err)
		return tx.Rollback()
	}

	fmt.Println("Коммит транзакции")
	return tx.Commit()
}

func (store *Postgres) createTable() error {
	ctx := context.Background()
	_, err := store.db.ExecContext(ctx, createTablesSQL)
	if err != nil {
		return err
	}

	return nil
}

func (store *Postgres) Get(ctx context.Context, metricDB metric.MetricDB) (metric.MetricDB, error) {
	var metricResponse postMetric
	err := store.exec(ctx, func(q *Query) error {
		// Подготавливаем запрос
		getMetricStmt, err := q.db.PrepareContext(ctx, getMetricSQL)
		if err != nil {
			return err
		}

		// Получаем метрику из базы по имени и типу
		if err := getMetricStmt.QueryRowContext(ctx,
			metricDB.Name(),
			metricDB.Type(),
		).Scan(
			&metricResponse.mType,
			&metricResponse.id,
			&metricResponse.cVal,
			&metricResponse.gVal,
		); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return metric.MetricDB{}, err
	}

	metricSend, err := NewMetricDBFromPostMetric(metricResponse)
	if err != nil {
		return metric.MetricDB{}, err
	}

	return metricSend, nil
}

func (store *Postgres) Set(ctx context.Context, metricDB metric.MetricDB) (metric.MetricDB, error) {
	err := store.exec(ctx, func(q *Query) error {
		setMetricStmt, err := q.db.PrepareContext(ctx, setMetricSQL)
		if err != nil {
			return err
		}

		pMetric, err := newPostMetricFromMetricDB(metricDB)
		if err != nil {
			return err
		}

		err = setMetricStmt.QueryRowContext(ctx,
			pMetric.mType,
			pMetric.id,
			pMetric.cVal,
			pMetric.gVal,
		).Scan(
			&pMetric.cVal,
			&pMetric.gVal,
		)

		if err != nil {
			return err
		}

		metricDB, err = NewMetricDBFromPostMetric(pMetric)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return metric.MetricDB{}, err
	}

	return metricDB, nil
}

func (store *Postgres) Update(ctx context.Context, metricDB metric.MetricDB) (metric.MetricDB, error) {
	err := store.execTx(ctx, func(q *Query) error {
		updateMetricStmt, err := q.db.PrepareContext(ctx, updateMetricSQL)
		if err != nil {
			return err
		}

		pMetric, err := newPostMetricFromMetricDB(metricDB)
		if err != nil {
			return err
		}

		err = updateMetricStmt.QueryRowContext(ctx,
			pMetric.mType,
			pMetric.id,
			pMetric.cVal,
			pMetric.gVal,
		).Scan(
			&pMetric.cVal,
			&pMetric.gVal,
		)

		if err != nil {
			return err
		}

		metricDB, err = NewMetricDBFromPostMetric(pMetric)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return metric.MetricDB{}, err
	}

	return metricDB, nil
}

func (store *Postgres) List(ctx context.Context) []metric.MetricDB {
	arr := []metric.MetricDB{}
	err := store.exec(ctx, func(q *Query) error {
		// Подготавливаем запрос
		listMetricStmt, err := q.db.PrepareContext(ctx, listMetricSQL)
		if err != nil {
			return err
		}

		rows, err := listMetricStmt.QueryContext(ctx)
		if err != nil {
			return err
		}

		for rows.Next() {
			var pMetric postMetric
			if err := rows.Scan(
				&pMetric.mType,
				&pMetric.id,
				&pMetric.cVal,
				&pMetric.gVal,
			); err != nil {
				return err
			}

			metricDB, err := NewMetricDBFromPostMetric(pMetric)
			if err != nil {
				return err
			}

			arr = append(arr, metricDB)
		}

		err = rows.Err()
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return []metric.MetricDB{}
	}

	return arr
}

func (store *Postgres) SetBatch(ctx context.Context, arr []metric.MetricDB) error {
	err := store.execTx(ctx, func(q *Query) error {
		setMetricStmt, err := q.db.PrepareContext(ctx, setMetricSQL)
		if err != nil {
			return err
		}

		for _, metricDB := range arr {
			var err error
			pMetric, err := newPostMetricFromMetricDB(metricDB)
			if err != nil {
				return err
			}
			_, err = setMetricStmt.ExecContext(ctx,
				pMetric.mType,
				pMetric.id,
				pMetric.cVal,
				pMetric.gVal,
			)

			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (store *Postgres) UpdateBatch(ctx context.Context, arr []metric.MetricDB) error {
	err := store.execTx(ctx, func(q *Query) error {
		updateMetricStmt, err := q.db.PrepareContext(ctx, updateMetricSQL)
		if err != nil {
			return err
		}

		for _, metricDB := range arr {
			var err error
			pMetric, err := newPostMetricFromMetricDB(metricDB)
			if err != nil {
				return err
			}
			_, err = updateMetricStmt.ExecContext(ctx,
				pMetric.mType,
				pMetric.id,
				pMetric.cVal,
				pMetric.gVal,
			)

			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
