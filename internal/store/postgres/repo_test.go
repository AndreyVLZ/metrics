package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/AndreyVLZ/metrics/internal/model"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func newMock() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	return db, mock
}

// Not err.
func TestGet(t *testing.T) {
	database, mock := newMock()

	repo := Repo[int64]{
		db:        database,
		tableName: nameTableCounter,
	}
	ctx := context.Background()

	mockRows := sqlmock.NewRows(
		[]string{"name", "type_id", "val"}).
		AddRow("nameMetric", 1, 100)

	mock.
		ExpectPrepare("SELECT name,type_id,val FROM counter WHERE name=$1").
		ExpectQuery().
		WithArgs("nameMetric").
		WillReturnRows(mockRows)

	metRepo, err := repo.Get(ctx, "nameMetric")
	if err != nil {
		t.Errorf("ge met: %v\n", err)
	}

	assert.Equal(t,
		metRepo,
		model.NewMetricRepo[int64](
			"nameMetric",
			model.TypeCountConst,
			100,
		),
	)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there WERE unfulfilled expectations: %s", err)
	}
}

// Ошибка при выполнении запрос к postrgres.
func TestGetErrStmt(t *testing.T) {
	database, mock := newMock()

	repo := Repo[int64]{
		db:        database,
		tableName: nameTableCounter,
	}

	ctx := context.Background()
	wantErr := errors.New("some ERR")

	mock.
		ExpectPrepare("SELECT name,type_id,val FROM counter WHERE name=$1").
		ExpectQuery().
		WithArgs("nameMetric").
		WillReturnError(wantErr)

	metRepo, err := repo.Get(ctx, "nameMetric")
	assert.Equal(t, fmt.Errorf("stmt exec: %w", wantErr), err)

	assert.Equal(t,
		metRepo,
		model.MetricRepo[int64]{},
	)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there WERE unfulfilled expectations: %s", err)
	}
}

// Ошибка sql.ErrNoRows.
func TestGetErrNoRows(t *testing.T) {
	database, mock := newMock()

	repo := Repo[int64]{
		db:        database,
		tableName: nameTableCounter,
	}

	ctx := context.Background()
	wantErr := errNotFind
	mockRows := sqlmock.NewRows(
		[]string{"name", "type_id", "val"},
	)

	mock.
		ExpectPrepare("SELECT name,type_id,val FROM counter WHERE name=$1").
		ExpectQuery().
		WithArgs("nameMetric").
		WillReturnRows(mockRows)

	metRepo, err := repo.Get(ctx, "nameMetric")
	assert.Equal(t, wantErr, err)

	assert.Equal(t,
		metRepo,
		model.MetricRepo[int64]{},
	)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there WERE unfulfilled expectations: %s", err)
	}
}

// Ошибка подготовки запроса.
func TestGetErrPrepareStmt(t *testing.T) {
	database, mock := newMock()

	repo := Repo[int64]{
		db:        database,
		tableName: nameTableCounter,
	}

	ctx := context.Background()
	errPrepare := errors.New("prepare ERR")
	wantErr := fmt.Errorf("prepare getStmt: %w", errPrepare)

	mock.
		ExpectPrepare("SELECT name,type_id,val FROM counter WHERE name=$1").
		WillReturnError(errPrepare)

	metRepo, err := repo.Get(ctx, "nameMetric")
	assert.Equal(t, wantErr, err)

	assert.Equal(t,
		metRepo,
		model.MetricRepo[int64]{},
	)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there WERE unfulfilled expectations: %s", err)
	}
}

func TestUpdateRepo(t *testing.T) {
	type testCase struct {
		name    string
		setMet  model.MetricRepo[int64]
		isErr   bool
		wantMet model.MetricRepo[int64]
		fnMock  func(sqlmock.Sqlmock)
	}

	tc := []testCase{
		{
			name: "ok set",
			setMet: model.NewMetricRepo[int64](
				"Counter-1",
				model.TypeCountConst,
				11,
			),
			isErr: false,
			wantMet: model.NewMetricRepo[int64](
				"Counter-1",          // name
				model.TypeCountConst, // type
				11,                   // val
			),
			fnMock: func(mock sqlmock.Sqlmock) {
				mockRows := sqlmock.NewRows(
					[]string{"name", "val"})

				mock.ExpectBegin() // транзакция

				// Подготова всех запросов.
				mock.ExpectPrepare(
					"SELECT name,type_id,val FROM counter WHERE name=$1",
				). // getSQL
					ExpectQuery().
					WithArgs("Counter-1").
					WillReturnRows(mockRows)
				mock.ExpectPrepare(
					"INSERT INTO counter (name,val) VALUES ($1,$2)",
				). // setSQL
					ExpectExec().
					WithArgs("Counter-1", 11).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()
			},
		},

		{
			name: "ok update",
			setMet: model.NewMetricRepo[int64](
				"Counter-1",
				model.TypeCountConst,
				11,
			),
			isErr: false,
			wantMet: model.NewMetricRepo[int64](
				"Counter-1",          // name
				model.TypeCountConst, // type
				111,                  // val
			),
			fnMock: func(mock sqlmock.Sqlmock) {
				mockRows := sqlmock.NewRows(
					[]string{"name", "type_id", "val"}).
					AddRow("Counter-1", 1, 100) // значения в базе

				mock.ExpectBegin() // транзакция

				// Подготова всех запросов.
				mock.ExpectPrepare(
					"SELECT name,type_id,val FROM counter WHERE name=$1",
				). // getSQL
					ExpectQuery().
					WithArgs("Counter-1").
					WillReturnRows(mockRows)
				mock.ExpectPrepare(
					"UPDATE counter SET val = $2 WHERE name = $1",
				). // updSQL
					ExpectExec().
					WithArgs("Counter-1", 111).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()
			},
		},

		{
			name: "err start tx",
			setMet: model.NewMetricRepo[int64](
				"Counter-1",
				model.TypeCountConst,
				11,
			),
			isErr:   true,
			wantMet: model.MetricRepo[int64]{},
			fnMock: func(mock sqlmock.Sqlmock) {
				mock.
					ExpectBegin(). // транзакция
					WillReturnError(errors.New("TX-ERR"))
			},
		},

		{
			name: "err get",
			setMet: model.NewMetricRepo[int64](
				"Counter-1",
				model.TypeCountConst,
				11,
			),
			isErr:   true,
			wantMet: model.MetricRepo[int64]{},
			fnMock: func(mock sqlmock.Sqlmock) {
				mockRows := sqlmock.NewRows(
					[]string{"name", "type_id", "val"}).
					AddRow("Counter-1", 1, 100).RowError(0, errors.New("row scan")) // значения в базе
				mock.ExpectBegin() // транзакция

				// Подготова всех запросов.
				mock.ExpectPrepare(
					"SELECT name,type_id,val FROM counter WHERE name=$1",
				). // getSQL
					ExpectQuery().
					WithArgs("Counter-1").
					WillReturnRows(mockRows)

				mock.ExpectRollback()
			},
		},

		{
			name: "err exec setStmt",
			setMet: model.NewMetricRepo[int64](
				"Counter-1",
				model.TypeCountConst,
				11,
			),
			isErr:   true,
			wantMet: model.MetricRepo[int64]{},
			fnMock: func(mock sqlmock.Sqlmock) {
				mockRows := sqlmock.NewRows(
					[]string{"name", "val"})

				mock.ExpectBegin() // транзакция

				// Подготова всех запросов.
				mock.ExpectPrepare(
					"SELECT name,type_id,val FROM counter WHERE name=$1",
				). // getSQL
					ExpectQuery().
					WithArgs("Counter-1").
					WillReturnRows(mockRows)
				mock.ExpectPrepare(
					"INSERT INTO counter (name,val) VALUES ($1,$2)",
				). // setSQL
					ExpectExec().
					WithArgs("Counter-1", 11).
					WillReturnResult(sqlmock.NewResult(1, 1)).
					WillReturnError(errors.New("execStmt ERR"))

				mock.ExpectRollback()
			},
		},

		{
			name: "err txCommit",
			setMet: model.NewMetricRepo[int64](
				"Counter-1",
				model.TypeCountConst,
				11,
			),
			isErr:   true,
			wantMet: model.MetricRepo[int64]{},
			fnMock: func(mock sqlmock.Sqlmock) {
				mockRows := sqlmock.NewRows(
					[]string{"name", "val"})

				mock.ExpectBegin() // транзакция

				// Подготова всех запросов.
				mock.ExpectPrepare(
					"SELECT name,type_id,val FROM counter WHERE name=$1",
				). // getSQL
					ExpectQuery().
					WithArgs("Counter-1").
					WillReturnRows(mockRows)
				mock.ExpectPrepare(
					"INSERT INTO counter (name,val) VALUES ($1,$2)",
				). // setSQL
					ExpectExec().
					WithArgs("Counter-1", 11).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.
					ExpectCommit().
					WillReturnError(errors.New("txCommit ERR"))
			},
		},

		{
			name: "err update",
			setMet: model.NewMetricRepo[int64](
				"Counter-1",
				model.TypeCountConst,
				11,
			),
			isErr:   true,
			wantMet: model.MetricRepo[int64]{},
			fnMock: func(mock sqlmock.Sqlmock) {
				mockRows := sqlmock.NewRows(
					[]string{"name", "val"})

				mock.ExpectBegin() // транзакция

				// Подготова всех запросов.
				mock.ExpectPrepare(
					"SELECT name,type_id,val FROM counter WHERE name=$1",
				). // getSQL
					ExpectQuery().
					WithArgs("Counter-1").
					WillReturnRows(mockRows)
				mock.ExpectPrepare(
					"INSERT INTO counter (name,val) VALUES ($1,$2)",
				). // setSQL
					WillReturnError(errors.New("prepare ERR"))

				mock.ExpectRollback()
			},
		},

		{
			name: "err update && rollback",
			setMet: model.NewMetricRepo[int64](
				"Counter-1",
				model.TypeCountConst,
				11,
			),
			isErr:   true,
			wantMet: model.MetricRepo[int64]{},
			fnMock: func(mock sqlmock.Sqlmock) {
				mockRows := sqlmock.NewRows(
					[]string{"name", "val"})

				mock.ExpectBegin() // транзакция

				// Подготова всех запросов.
				mock.ExpectPrepare(
					"SELECT name,type_id,val FROM counter WHERE name=$1",
				). // getSQL
					ExpectQuery().
					WithArgs("Counter-1").
					WillReturnRows(mockRows)
				mock.ExpectPrepare(
					"INSERT INTO counter (name,val) VALUES ($1,$2)",
				). // setSQL
					WillReturnError(errors.New("prepare ERR"))

				mock.
					ExpectRollback().
					WillReturnError(errors.New("rollback ERR"))
			},
		},

		{
			name: "err execUpdStmt",
			setMet: model.NewMetricRepo[int64](
				"Counter-1",
				model.TypeCountConst,
				11,
			),
			isErr:   false,
			wantMet: model.MetricRepo[int64]{},
			fnMock: func(mock sqlmock.Sqlmock) {
				mockRows := sqlmock.NewRows(
					[]string{"name", "type_id", "val"}).
					AddRow("Counter-1", 1, 100) // значения в базе

				mock.ExpectBegin() // транзакция

				// Подготова всех запросов.
				mock.ExpectPrepare(
					"SELECT name,type_id,val FROM counter WHERE name=$1",
				). // getSQL
					ExpectQuery().
					WithArgs("Counter-1").
					WillReturnRows(mockRows)
				mock.ExpectPrepare(
					"UPDATE counter SET val = $2 WHERE name = $1",
				). // updSQL
					ExpectExec().
					WithArgs("Counter-1", 111).
					WillReturnError(errors.New("exec ERR"))

				mock.ExpectRollback()
			},
		},

		{
			name: "err prepare UpdStmt",
			setMet: model.NewMetricRepo[int64](
				"Counter-1",
				model.TypeCountConst,
				11,
			),
			isErr:   false,
			wantMet: model.MetricRepo[int64]{},
			fnMock: func(mock sqlmock.Sqlmock) {
				mockRows := sqlmock.NewRows(
					[]string{"name", "type_id", "val"}).
					AddRow("Counter-1", 1, 100) // значения в базе

				mock.ExpectBegin() // транзакция

				// Подготова всех запросов.
				mock.ExpectPrepare(
					"SELECT name,type_id,val FROM counter WHERE name=$1",
				). // getSQL
					ExpectQuery().
					WithArgs("Counter-1").
					WillReturnRows(mockRows)
				mock.ExpectPrepare(
					"UPDATE counter SET val = $2 WHERE name = $1",
				). // updSQL
					WillReturnError(errors.New("prepare ERR"))

				mock.ExpectRollback()
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()

			database, mock := newMock()
			repo := Repo[int64]{
				db:        database,
				tableName: nameTableCounter,
			}

			test.fnMock(mock)

			metNew, err := repo.Update(ctx, test.setMet)
			if test.isErr && err == nil {
				t.Error("want err")
			}

			assert.Equal(t, test.wantMet, metNew)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there WERE unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestListRepo(t *testing.T) {
	type testCase struct {
		name  string
		isErr bool
		wArr  []model.MetricRepo[int64]
		fnSet func(sqlmock.Sqlmock)
	}

	tc := []testCase{
		{
			name:  "ok",
			isErr: false,
			wArr: []model.MetricRepo[int64]{
				model.NewMetricRepo[int64](
					"Counter-1",
					model.TypeCountConst,
					100,
				),
			},
			fnSet: func(mock sqlmock.Sqlmock) {
				mockRows := sqlmock.NewRows(
					[]string{"name", "type_id", "val"}).
					AddRow("Counter-1", 1, 100)
				mock.ExpectPrepare(
					"SELECT c.name,c.type_id,c.val FROM counter c LEFT JOIN mtype mt USING (type_id)",
				). // listSQL
					ExpectQuery().
					WithoutArgs().
					WillReturnRows(mockRows)
			},
		},

		{
			name:  "err prepare stmt",
			isErr: true,
			wArr:  nil,
			fnSet: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare(
					"SELECT c.name,c.type_id,c.val FROM counter c LEFT JOIN mtype mt USING (type_id)",
				). // listSQL
					WillReturnError(errors.New("PRE ERR"))
			},
		},

		{
			name:  "err exe stmt",
			isErr: true,
			wArr:  nil,
			fnSet: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare(
					"SELECT c.name,c.type_id,c.val FROM counter c LEFT JOIN mtype mt USING (type_id)",
				). // listSQL
					ExpectQuery().
					WithoutArgs().
					WillReturnError(errors.New("exeERR"))
			},
		},

		{
			name:  "err rows",
			isErr: true,
			wArr:  nil,
			fnSet: func(mock sqlmock.Sqlmock) {
				mockRows := sqlmock.NewRows(
					[]string{"name", "type_id", "val"}).
					AddRow(nil, nil, nil)
				mock.ExpectPrepare(
					"SELECT c.name,c.type_id,c.val FROM counter c LEFT JOIN mtype mt USING (type_id)",
				). // listSQL
					ExpectQuery().
					WithoutArgs().
					WillReturnRows(mockRows)
			},
		},
		{
			name:  "err scan row",
			isErr: true,
			wArr:  nil,
			fnSet: func(mock sqlmock.Sqlmock) {
				mockRows := sqlmock.NewRows(
					[]string{"name", "type_id", "val"}).
					AddRow("Counter-1", 1, 100).
					AddRow("Counter-2", 1, 200).
					RowError(1, errors.New("row scan ERR"))
				mock.ExpectPrepare(
					"SELECT c.name,c.type_id,c.val FROM counter c LEFT JOIN mtype mt USING (type_id)",
				). // listSQL
					ExpectQuery().
					WithoutArgs().
					WillReturnRows(mockRows)
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()

			database, mock := newMock()
			repo := Repo[int64]{
				db:        database,
				tableName: nameTableCounter,
			}

			test.fnSet(mock)

			arrMet, err := repo.list(ctx, database)

			if test.isErr && err == nil {
				t.Error("want err")
			}

			assert.Equal(t, test.wArr, arrMet)
		})
	}
}

type fakeiDBTX struct {
	DBTX
	err error
}

func (f fakeiDBTX) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	if f.err != nil {
		return nil, f.err
	}

	return f.DBTX.PrepareContext(ctx, query)
}

func TestAddList(t *testing.T) {
	type testCase struct {
		name  string
		isErr bool
		arr   []model.MetricRepo[int64]
		fnSet func(sqlmock.Sqlmock)
	}

	tc := []testCase{
		{
			name:  "ok addList",
			isErr: false,
			arr: []model.MetricRepo[int64]{
				model.NewMetricRepo[int64](
					"Counter-1",
					model.TypeCountConst,
					100,
				),
			},
			fnSet: func(mock sqlmock.Sqlmock) {
				mockRows := sqlmock.NewRows(
					[]string{"name", "val"})

				// Подготова всех запросов.
				mock.ExpectPrepare(
					"SELECT name,type_id,val FROM counter WHERE name=$1",
				). // getSQL
					ExpectQuery().
					WithArgs("Counter-1").
					WillReturnRows(mockRows)
				mock.ExpectPrepare(
					"INSERT INTO counter (name,val) VALUES ($1,$2)",
				). // setSQL
					ExpectExec().
					WithArgs("Counter-1", 100).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},

		{
			name:  "err addList",
			isErr: true,
			arr: []model.MetricRepo[int64]{
				model.NewMetricRepo[int64](
					"Counter-1",
					model.TypeCountConst,
					100,
				),
			},
			fnSet: func(mock sqlmock.Sqlmock) {
				// mock.ExpectBegin() // транзакция

				// Подготова всех запросов.
				mock.ExpectPrepare(
					"SELECT name,type_id,val FROM counter WHERE name=$1",
				).WillReturnError(errors.New("errPrepare")) // getSQL
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()

			database, mock := newMock()
			repo := Repo[int64]{
				db:        database,
				tableName: nameTableCounter,
			}

			test.fnSet(mock)

			err := repo.addList(ctx, database, test.arr)
			fmt.Printf("ERR %v\n", err)
			if test.isErr && err == nil {
				t.Error("want err")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there WERE unfulfilled expectations: %s", err)
			}
		})
	}
}
