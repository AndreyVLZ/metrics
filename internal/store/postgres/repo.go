package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AndreyVLZ/metrics/internal/model"
)

const (
	keyTxBegin        = "txBegin"
	keyPrepareGetStmt = "prepare getStmt"
)

type customError struct {
	error
	key string
}

func (err customError) Error() string {
	return fmt.Sprintf("%s: %s", err.key, err.error)
}

var errNotFind = errors.New("not find")

type metricRepo[VT model.ValueType] struct {
	name  string
	mType model.TypeMetric
	val   VT
}

type Repo[VT model.ValueType] struct {
	db        *sql.DB
	tableName string
}

func newRepo[VT model.ValueType](tableName string) Repo[VT] {
	return Repo[VT]{tableName: tableName}
}

func (r Repo[VT]) Get(ctx context.Context, name string) (model.MetricRepo[VT], error) {
	getStmt, err := r.db.PrepareContext(ctx,
		fmt.Sprintf(getSQL, r.tableName),
	)
	if err != nil {
		return model.MetricRepo[VT]{}, fmt.Errorf("prepare getStmt: %w", err)
	}

	defer getStmt.Close()

	return r.get(ctx, getStmt, name)
}

func (r *Repo[VT]) update(
	ctx context.Context,
	dbtx DBTX,
	met model.MetricRepo[VT],
) (model.MetricRepo[VT], error) {
	getStmt, err := dbtx.PrepareContext(ctx, fmt.Sprintf(getSQL, r.tableName))
	if err != nil {
		return model.MetricRepo[VT]{},
			customError{key: keyPrepareGetStmt, error: err}
	}

	metDB, err := r.get(ctx, getStmt, met.Name())
	if err != nil && !errors.Is(err, errNotFind) {
		return model.MetricRepo[VT]{}, fmt.Errorf("getter: %w", err)
	}

	if !errors.Is(errNotFind, err) {
		metDB.Update(met.Value())

		updStmt, err := dbtx.PrepareContext(ctx,
			fmt.Sprintf(updateSQL, r.tableName),
		)
		if err != nil {
			return model.MetricRepo[VT]{}, fmt.Errorf("prepare updStmt: %w", err)
		}
		defer updStmt.Close()

		args := []any{
			metDB.Name(),
			metDB.Value(),
		}

		if _, err := updStmt.ExecContext(ctx, args...); err != nil {
			return model.MetricRepo[VT]{}, fmt.Errorf("updStmt-exec: %w", err)
		}

		return metDB, nil
	}

	setStmt, err := dbtx.PrepareContext(ctx,
		fmt.Sprintf(setSQL, r.tableName),
	)
	if err != nil {
		return model.MetricRepo[VT]{}, fmt.Errorf("prepare setStmt: %w", err)
	}
	defer setStmt.Close()

	args := []any{
		met.Name(),
		met.Value(),
	}

	if _, err := setStmt.ExecContext(ctx, args...); err != nil {
		return model.MetricRepo[VT]{}, fmt.Errorf("setStmt-exec: %w", err)
	}

	return met, nil
}

func (r *Repo[VT]) Update(ctx context.Context, met model.MetricRepo[VT]) (model.MetricRepo[VT], error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.MetricRepo[VT]{}, customError{
			key:   keyTxBegin,
			error: err,
		}
	}

	metDB, err := r.update(ctx, tx, met)
	if err != nil {
		if errRoll := tx.Rollback(); errRoll != nil {
			err = errors.Join(err, fmt.Errorf("txRollback: %w", errRoll))
		}

		return model.MetricRepo[VT]{}, fmt.Errorf("updater: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return model.MetricRepo[VT]{}, fmt.Errorf("txCommit: %w", err)
	}

	return metDB, nil
}

func (r *Repo[VT]) list(ctx context.Context, dbtx DBTX) ([]model.MetricRepo[VT], error) {
	listStmt, err := dbtx.PrepareContext(ctx, fmt.Sprintf(listSQL, r.tableName))
	if err != nil {
		return nil, fmt.Errorf("prepare listStmt: %w", err)
	}
	defer listStmt.Close()

	rows, err := listStmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("stmt exec: %w", err)
	}
	defer rows.Close()

	arr := make([]model.MetricRepo[VT], 0)

	for rows.Next() {
		var met metricRepo[VT]
		desk := []any{
			&met.name,
			&met.mType,
			&met.val,
		}

		if err := rows.Scan(desk...); err != nil {
			return nil, fmt.Errorf("row scan: %w", err)
		}

		arr = append(arr, model.NewMetricRepo(met.name, met.mType, met.val))
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rowsErr: %w", err)
	}

	return arr, nil
}

func (r *Repo[VT]) addList(ctx context.Context, dbtx DBTX, arr []model.MetricRepo[VT]) error {
	for i := range arr {
		if _, err := r.update(ctx, dbtx, arr[i]); err != nil {
			return fmt.Errorf("updater: %w", err)
		}
	}

	return nil
}

func (r Repo[VT]) get(ctx context.Context, getStmt *sql.Stmt, name string) (model.MetricRepo[VT], error) {
	var met metricRepo[VT]

	args := []any{name}
	dest := []any{
		&met.name,
		&met.mType,
		&met.val,
	}

	if err := getStmt.QueryRowContext(ctx, args...).Scan(dest...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.MetricRepo[VT]{}, errNotFind
		}

		return model.MetricRepo[VT]{}, fmt.Errorf("stmt exec: %w", err)
	}

	mRepo := model.NewMetricRepo(met.name, met.mType, met.val)

	return mRepo, nil
}
