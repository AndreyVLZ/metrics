package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/AndreyVLZ/metrics/internal/model"
	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	store := Postgres{}
	assert.Equal(t, NameConst, store.Name())
}

func TestCreateTable(t *testing.T) {
	type testCase struct {
		name   string
		isErr  bool
		fnMock func(sqlmock.Sqlmock)
	}

	tc := []testCase{
		{
			name:  "ok",
			isErr: false,
			fnMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(createTables).
					WithoutArgs().
					WillReturnResult(sqlmock.NewResult(1, 1))

				pre1 := mock.ExpectPrepare("INSERT INTO mtype (type_id, name_type) VALUES ($1,$2) ON CONFLICT (name_type) DO NOTHING")
				pre1.ExpectExec().
					WithArgs(model.TypeCountConst, model.TypeCountConst.String()).
					WillReturnResult(sqlmock.NewResult(1, 1))

				pre1.ExpectExec().
					WithArgs(model.TypeGaugeConst, model.TypeGaugeConst.String()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},

		{
			name:  "err exec createTable",
			isErr: true,
			fnMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(createTables).
					WillReturnError(errors.New("err exec"))
			},
		},

		{
			name:  "err prepare setSql",
			isErr: true,
			fnMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(createTables).
					WithoutArgs().
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.
					ExpectPrepare("INSERT INTO mtype (type_id, name_type) VALUES ($1,$2) ON CONFLICT (name_type) DO NOTHING").
					WillReturnError(errors.New("err prepare"))
			},
		},

		{
			name:  "err exec stmtSetSql",
			isErr: false,
			fnMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(createTables).
					WithoutArgs().
					WillReturnResult(sqlmock.NewResult(1, 1))

				pre1 := mock.ExpectPrepare("INSERT INTO mtype (type_id, name_type) VALUES ($1,$2) ON CONFLICT (name_type) DO NOTHING")
				pre1.
					ExpectExec().
					WillReturnError(errors.New("err exec stmt"))
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			database, mock := newMock()
			store := Postgres{db: database}

			test.fnMock(mock)

			err := store.createTable(ctx)
			if test.isErr && err == nil {
				t.Error("want err")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there WERE unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestAddBatch(t *testing.T) {
	type testCase struct {
		name   string
		batch  model.Batch
		isErr  bool
		fnMock func(sqlmock.Sqlmock)
	}

	tc := []testCase{
		{
			name:  "ok",
			isErr: false,
			batch: model.Batch{
				CList: []model.MetricRepo[int64]{
					model.NewMetricRepo[int64](
						"Counter-1",
						model.TypeCountConst,
						10,
					),
				},
				GList: []model.MetricRepo[float64]{
					model.NewMetricRepo[float64](
						"Gauge-1",
						model.TypeGaugeConst,
						10.01,
					),
				},
			},
			fnMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
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
					WithArgs("Counter-1", 10).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectPrepare(
					"SELECT name,type_id,val FROM gauge WHERE name=$1",
				). // getSQL
					ExpectQuery().
					WithArgs("Gauge-1").
					WillReturnRows(mockRows)
				mock.ExpectPrepare(
					"INSERT INTO gauge (name,val) VALUES ($1,$2)",
				). // setSQL
					ExpectExec().
					WithArgs("Gauge-1", 10.01).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()
			},
		},

		{
			name:  "err tx commit",
			isErr: true,
			batch: model.Batch{
				CList: []model.MetricRepo[int64]{
					model.NewMetricRepo[int64](
						"Counter-1",
						model.TypeCountConst,
						10,
					),
				},
				GList: []model.MetricRepo[float64]{
					model.NewMetricRepo[float64](
						"Gauge-1",
						model.TypeGaugeConst,
						10.01,
					),
				},
			},
			fnMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
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
					WithArgs("Counter-1", 10).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectPrepare(
					"SELECT name,type_id,val FROM gauge WHERE name=$1",
				). // getSQL
					ExpectQuery().
					WithArgs("Gauge-1").
					WillReturnRows(mockRows)
				mock.ExpectPrepare(
					"INSERT INTO gauge (name,val) VALUES ($1,$2)",
				). // setSQL
					ExpectExec().
					WithArgs("Gauge-1", 10.01).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit().WillReturnError(errors.New("commit err"))
			},
		},

		{
			name:  "err addBatchTx && Rollback Clist",
			isErr: true,
			batch: model.Batch{
				CList: []model.MetricRepo[int64]{
					model.NewMetricRepo[int64](
						"Counter-1",
						model.TypeCountConst,
						10,
					),
				},
				GList: []model.MetricRepo[float64]{
					model.NewMetricRepo[float64](
						"Gauge-1",
						model.TypeGaugeConst,
						10.01,
					),
				},
			},
			fnMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				// Подготова всех запросов.
				mock.
					ExpectPrepare( // getSQL
						"SELECT name,type_id,val FROM counter WHERE name=$1",
					).
					WillReturnError(errors.New("err addBatch err"))

				mock.ExpectRollback().WillReturnError(errors.New("rollback err"))
			},
		},

		{
			name:  "err addBatchTx && Rollback Glist",
			isErr: true,
			batch: model.Batch{
				CList: []model.MetricRepo[int64]{
					model.NewMetricRepo[int64](
						"Counter-1",
						model.TypeCountConst,
						10,
					),
				},
				GList: []model.MetricRepo[float64]{
					model.NewMetricRepo[float64](
						"Gauge-1",
						model.TypeGaugeConst,
						10.01,
					),
				},
			},
			fnMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
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
					WithArgs("Counter-1", 10).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.
					ExpectPrepare( // getSQL
						"SELECT name,type_id,val FROM gauge WHERE name=$1",
					).
					WillReturnError(errors.New("err addBatch err"))

				mock.ExpectRollback().WillReturnError(errors.New("rollback err"))
			},
		},

		{
			name:  "err txBegin",
			isErr: true,
			batch: model.Batch{
				CList: []model.MetricRepo[int64]{
					model.NewMetricRepo[int64](
						"Counter-1",
						model.TypeCountConst,
						10,
					),
				},
				GList: []model.MetricRepo[float64]{
					model.NewMetricRepo[float64](
						"Gauge-1",
						model.TypeGaugeConst,
						10.01,
					),
				},
			},
			fnMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(errors.New("err tx begin"))
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			database, mock := newMock()

			store := Postgres{
				cRepo: Repo[int64]{tableName: nameTableCounter},
				gRepo: Repo[float64]{tableName: nameTableGauge},
			}
			store.setDatabase(database)

			test.fnMock(mock)

			err := store.AddBatch(ctx, test.batch)
			if test.isErr && err == nil {
				t.Error("want err")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there WERE unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestList(t *testing.T) {
	type testCase struct {
		name   string
		isErr  bool
		wBatch model.Batch
		fnMock func(sqlmock.Sqlmock)
	}

	tc := []testCase{
		{
			name:  "ok List",
			isErr: false,
			wBatch: model.Batch{
				CList: []model.MetricRepo[int64]{
					model.NewMetricRepo[int64](
						"Counter-1",
						model.TypeCountConst,
						10,
					),
				},
				GList: []model.MetricRepo[float64]{
					model.NewMetricRepo[float64](
						"Gauge-1",
						model.TypeGaugeConst,
						10.01,
					),
				},
			},

			fnMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mockCRows := sqlmock.NewRows(
					[]string{"name", "type_id", "val"}).
					AddRow("Counter-1", 1, 10)
				mockGRows := sqlmock.NewRows(
					[]string{"name", "type_id", "val"}).
					AddRow("Gauge-1", 2, 10.01)

				mock.ExpectPrepare( // ClistSQL
					"SELECT c.name,c.type_id,c.val FROM counter c LEFT JOIN mtype mt USING (type_id)",
				).ExpectQuery().
					WithoutArgs().
					WillReturnRows(mockCRows)

				mock.ExpectPrepare( // GlistSQL
					"SELECT c.name,c.type_id,c.val FROM gauge c LEFT JOIN mtype mt USING (type_id)",
				).ExpectQuery().
					WithoutArgs().
					WillReturnRows(mockGRows)

				mock.ExpectCommit()
			},
		},

		{
			name:   "err txCommit",
			isErr:  false,
			wBatch: model.Batch{},
			fnMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mockCRows := sqlmock.NewRows(
					[]string{"name", "type_id", "val"}).
					AddRow("Counter-1", 1, 10)
				mockGRows := sqlmock.NewRows(
					[]string{"name", "type_id", "val"}).
					AddRow("Gauge-1", 2, 10.01)

				mock.ExpectPrepare( // ClistSQL
					"SELECT c.name,c.type_id,c.val FROM counter c LEFT JOIN mtype mt USING (type_id)",
				).ExpectQuery().
					WithoutArgs().
					WillReturnRows(mockCRows)

				mock.ExpectPrepare( // GlistSQL
					"SELECT c.name,c.type_id,c.val FROM gauge c LEFT JOIN mtype mt USING (type_id)",
				).ExpectQuery().
					WithoutArgs().
					WillReturnRows(mockGRows)

				mock.
					ExpectCommit().
					WillReturnError(errors.New("txCommit ERR"))
			},
		},
		{
			name:  "err exeCStmt",
			isErr: true,
			wBatch: model.Batch{
				CList: nil,
				GList: nil,
			},

			fnMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mockCRows := sqlmock.NewRows(
					[]string{"name", "type_id", "val"}).
					AddRow("Counter-1", 1, 10)

				mock.ExpectPrepare( // ClistSQL
					"SELECT c.name,c.type_id,c.val FROM counter c LEFT JOIN mtype mt USING (type_id)",
				).ExpectQuery().
					WithoutArgs().
					WillReturnRows(mockCRows)

				mock.ExpectPrepare( // GlistSQL
					"SELECT c.name,c.type_id,c.val FROM gauge c LEFT JOIN mtype mt USING (type_id)",
				).WillReturnError(errors.New("exe ERR"))

				mock.ExpectRollback()
			},
		},

		{
			name:   "err exeGStmt",
			isErr:  true,
			wBatch: model.Batch{},

			fnMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectPrepare( // ClistSQL
					"SELECT c.name,c.type_id,c.val FROM counter c LEFT JOIN mtype mt USING (type_id)",
				).WillReturnError(errors.New("exe ERR"))

				mock.ExpectRollback()
			},
		},

		{
			name:   "err exeStmt && rollback",
			isErr:  true,
			wBatch: model.Batch{},

			fnMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectPrepare( // ClistSQL
					"SELECT c.name,c.type_id,c.val FROM counter c LEFT JOIN mtype mt USING (type_id)",
				).WillReturnError(errors.New("exe ERR"))

				mock.
					ExpectRollback().
					WillReturnError(errors.New("rollback ERR"))
			},
		},

		{
			name:   "err txBegin",
			isErr:  true,
			wBatch: model.Batch{},

			fnMock: func(mock sqlmock.Sqlmock) {
				mock.
					ExpectBegin().
					WillReturnError(errors.New("txBegin ERR"))
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			database, mock := newMock()

			store := Postgres{
				cRepo: Repo[int64]{tableName: nameTableCounter},
				gRepo: Repo[float64]{tableName: nameTableGauge},
			}
			store.setDatabase(database)

			test.fnMock(mock)

			batch, err := store.List(ctx)
			fmt.Printf("ERRR %v\n", err)
			if test.isErr && err == nil {
				t.Error("want err")
			}

			assert.Equal(t, test.wBatch, batch)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there WERE unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestGetCounter(t *testing.T) {
	type testCase[VT model.ValueType] struct {
		name    string
		metName string
		isErr   bool
		wMet    model.MetricRepo[VT]
		fnMock  func(sqlmock.Sqlmock)
	}

	tc := []testCase[int64]{
		{
			name:    "ok GetCounter",
			metName: "Counter-1",
			isErr:   false,
			wMet: model.NewMetricRepo[int64](
				"Counter-1",
				model.TypeCountConst,
				10,
			),
			fnMock: func(mock sqlmock.Sqlmock) {
				mockRows := sqlmock.NewRows(
					[]string{"name", "type_id", "val"}).
					AddRow("Counter-1", 1, 10)

				mock.
					ExpectPrepare("SELECT name,type_id,val FROM counter WHERE name=$1").
					ExpectQuery().
					WithArgs("Counter-1").
					WillReturnRows(mockRows)
			},
		},

		{
			name:    "err exeStmt",
			metName: "Counter-1",
			isErr:   true,
			wMet:    model.MetricRepo[int64]{},
			fnMock: func(mock sqlmock.Sqlmock) {
				mock.
					ExpectPrepare("SELECT name,type_id,val FROM counter WHERE name=$1").
					WillReturnError(errors.New("exec ERR"))
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			database, mock := newMock()

			store := Postgres{
				cRepo: Repo[int64]{tableName: nameTableCounter},
				gRepo: Repo[float64]{tableName: nameTableGauge},
			}
			store.setDatabase(database)

			test.fnMock(mock)

			metDB, err := store.GetCounter(ctx, test.metName)
			fmt.Printf("ERR %v\n", err)
			if test.isErr && err == nil {
				t.Error("want err")
			}

			assert.Equal(t, test.wMet, metDB)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there WERE unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestGetGauge(t *testing.T) {
	type testCase[VT model.ValueType] struct {
		name    string
		metName string
		isErr   bool
		wMet    model.MetricRepo[VT]
		fnMock  func(sqlmock.Sqlmock)
	}

	tc := []testCase[float64]{
		{
			name:    "ok GetGauge",
			metName: "Gauge-1",
			isErr:   false,
			wMet: model.NewMetricRepo[float64](
				"Gauge-1",
				model.TypeGaugeConst,
				10.01,
			),
			fnMock: func(mock sqlmock.Sqlmock) {
				mockRows := sqlmock.NewRows(
					[]string{"name", "type_id", "val"}).
					AddRow("Gauge-1", 2, 10.01)

				mock.
					ExpectPrepare("SELECT name,type_id,val FROM gauge WHERE name=$1").
					ExpectQuery().
					WithArgs("Gauge-1").
					WillReturnRows(mockRows)
			},
		},

		{
			name:    "err exeStmt",
			metName: "Gauge-1",
			isErr:   true,
			wMet:    model.MetricRepo[float64]{},
			fnMock: func(mock sqlmock.Sqlmock) {
				mock.
					ExpectPrepare("SELECT name,type_id,val FROM gauge WHERE name=$1").
					WillReturnError(errors.New("exec ERR"))
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			database, mock := newMock()

			store := Postgres{
				cRepo: Repo[int64]{tableName: nameTableCounter},
				gRepo: Repo[float64]{tableName: nameTableGauge},
			}
			store.setDatabase(database)

			test.fnMock(mock)

			metDB, err := store.GetGauge(ctx, test.metName)
			fmt.Printf("ERR %v\n", err)
			if test.isErr && err == nil {
				t.Error("want err")
			}

			assert.Equal(t, test.wMet, metDB)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there WERE unfulfilled expectations: %s", err)
			}
		})
	}
}
