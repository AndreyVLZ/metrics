package memstore

import (
	"context"
	"testing"

	"github.com/AndreyVLZ/metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

type data[VT model.ValueType] struct {
	name string
	val  VT
	err  error
}

type test[VT model.ValueType] struct {
	name    string
	init    []model.MetricRepo[VT]
	exect   []data[VT]
	wantArr []model.MetricRepo[VT]
}
type tCase struct {
	tInt   test[int64]
	tFloat test[float64]
}

func TestStartStop(t *testing.T) {
	mem := New()
	if err := mem.Start(context.Background()); err != nil {
		t.Errorf("start err: %v\n", err)
	}

	if err := mem.Stop(context.Background()); err != nil {
		t.Errorf("start err: %v\n", err)
	}
}

func TestName(t *testing.T) {
	t.Run("checkName", func(t *testing.T) {
		store := New()
		assert.Equal(t, NameConst, store.Name())
	})
}

func TestPing(t *testing.T) {
	t.Run("checkPing", func(t *testing.T) {
		store := New()
		assert.Equal(t, nil, store.Ping())
	})
}

func TestGet(t *testing.T) {
	tCase := tCase{
		tInt: test[int64]{
			name: "counter",
			init: []model.MetricRepo[int64]{
				model.NewMetricRepo[int64]("Counter-1", model.TypeCountConst, 10),
				model.NewMetricRepo[int64]("Counter-2", model.TypeCountConst, 20),
				model.NewMetricRepo[int64]("Counter-2", model.TypeCountConst, 30),
			},
			exect: []data[int64]{
				{
					name: "Counter-1",
					val:  10,
					err:  nil,
				},
				{
					name: "Counter-2",
					val:  50,
					err:  nil,
				},
				{
					name: "Counter-0",
					val:  0,
					err:  errNotFind,
				},
			},
			wantArr: []model.MetricRepo[int64]{
				model.NewMetricRepo[int64]("Counter-1", model.TypeCountConst, 10),
				model.NewMetricRepo[int64]("Counter-2", model.TypeCountConst, 50),
			},
		},

		tFloat: test[float64]{
			name: "gauge",
			init: []model.MetricRepo[float64]{
				model.NewMetricRepo("Gauge-1", model.TypeGaugeConst, 10.01),
				model.NewMetricRepo("Gauge-2", model.TypeGaugeConst, 20.02),
				model.NewMetricRepo("Gauge-2", model.TypeGaugeConst, 30.03),
			},
			exect: []data[float64]{
				{
					name: "Gauge-1",
					val:  10.01,
					err:  nil,
				},
				{
					name: "Gauge-2",
					val:  30.03,
					err:  nil,
				},
				{
					name: "Gauge-0",
					val:  0,
					err:  errNotFind,
				},
			},
			wantArr: []model.MetricRepo[float64]{
				model.NewMetricRepo("Gauge-1", model.TypeGaugeConst, 10.01),
				model.NewMetricRepo("Gauge-2", model.TypeGaugeConst, 30.03),
			},
		},
	}

	ctx := context.Background()
	store := New()
	fUpd(ctx, t, tCase.tInt.name, store.UpdateCounter, tCase.tInt.init)
	fGet(ctx, t, tCase.tInt.name, store.GetCounter, tCase.tInt)
	fUpd(ctx, t, tCase.tFloat.name, store.UpdateGauge, tCase.tFloat.init)
	fGet(ctx, t, tCase.tFloat.name, store.GetGauge, tCase.tFloat)

	store = New()
	batch := model.Batch{CList: tCase.tInt.init, GList: tCase.tFloat.init}
	fUpdbatch(ctx, t, store.AddBatch, batch)

	actBatch, err := store.List(ctx)
	if err != nil {
		t.Errorf("list: %v\n", err)
	}

	assert.ElementsMatch(t,
		tCase.tInt.wantArr,
		actBatch.CList)

	assert.ElementsMatch(t,
		tCase.tFloat.wantArr,
		actBatch.GList)
}

func fUpdbatch(
	ctx context.Context,
	t *testing.T,
	fnUpdBatch func(ctx context.Context, batch model.Batch) error,
	batch model.Batch,
) {
	t.Run("update batch", func(t *testing.T) {
		if err := fnUpdBatch(ctx, batch); err != nil {
			t.Errorf("updBatch: %v\n", err)
		}
	})
}

func fUpd[VT model.ValueType](
	ctx context.Context,
	t *testing.T,
	nameTest string,
	fnAdd func(context.Context, model.MetricRepo[VT]) (model.MetricRepo[VT], error),
	arr []model.MetricRepo[VT],
) {
	t.Run("update "+nameTest, func(t *testing.T) {
		for i := range arr {
			metAdd := arr[i]
			_, err := fnAdd(ctx, metAdd)
			if err != nil {
				t.Fatalf("initStore %v\n", err)
			}
		}
	})
}

func fGet[VT model.ValueType](
	ctx context.Context,
	t *testing.T,
	nameTest string,
	fnGet func(context.Context, string) (model.MetricRepo[VT], error),
	test test[VT],
) {
	t.Run("get "+nameTest, func(t *testing.T) {
		for i := range test.exect {
			metEx := test.exect[i]

			metCheck, err := fnGet(ctx, metEx.name)

			if !assert.Equal(t, metEx.err, err) {
				t.Errorf("err- exec %v act%v exName %v cheName %v\n", metEx.err, err, metEx.name, metCheck.Name())

				return
			}

			if metEx.err != nil {
				continue
			}

			assert.Equal(t, metEx.name, metCheck.Name())
			assert.Equal(t, metEx.val, metCheck.Value())
		}
	})
}
