package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type test[VT ValueType] struct {
	name  string
	mType TypeMetric
	val   VT
	exep  MetricRepo[VT]
}

type tCase struct {
	tInt   test[int64]
	tFloat test[float64]
}

func TestNewMetricRepo(t *testing.T) {
	tc := tCase{
		tInt: test[int64]{
			name:  "Count1",
			mType: TypeCountConst,
			val:   1,
			exep:  MetricRepo[int64]{name: "Count1", mType: TypeCountConst, val: 1},
		},
		tFloat: test[float64]{
			name:  "Gauge1",
			mType: TypeGaugeConst,
			val:   1.1,
			exep:  MetricRepo[float64]{name: "Gauge1", mType: TypeGaugeConst, val: 1.1},
		},
	}
	fNew(t, "counter", tc.tInt)
	fNew(t, "gauge", tc.tFloat)
}

func fNew[VT ValueType](t *testing.T, tName string, test test[VT]) {
	t.Run("newMetricRepo "+tName, func(t *testing.T) {
		m := NewMetricRepo[VT](test.name, test.mType, test.val)
		assert.Equal(t, test.exep, m)
	})
}

type testType[VT ValueType] struct {
	met     MetricRepo[VT]
	expType TypeMetric
}

type caseType struct {
	tInt   testType[int64]
	tFloat testType[float64]
}

func TestTypeMetricRepo(t *testing.T) {
	tc := caseType{
		tInt: testType[int64]{
			met:     MetricRepo[int64]{name: "Count1", mType: TypeCountConst, val: 1},
			expType: TypeCountConst,
		},
		tFloat: testType[float64]{
			met:     MetricRepo[float64]{name: "Gauge1", mType: TypeGaugeConst, val: 1.1},
			expType: TypeGaugeConst,
		},
	}

	fType(t, "counter", tc.tInt)
	fType(t, "gauge", tc.tFloat)
}

func fType[VT ValueType](t *testing.T, tName string, test testType[VT]) {
	t.Run("valMetricRepo "+tName, func(t *testing.T) {
		assert.Equal(t, test.met.Type(), test.expType.String())
	})
}

type testVal[VT ValueType] struct {
	met    MetricRepo[VT]
	expVal VT
}

type caseVal struct {
	tInt   testVal[int64]
	tFloat testVal[float64]
}

func TestValMetricRepo(t *testing.T) {
	tc := caseVal{
		tInt: testVal[int64]{
			met:    MetricRepo[int64]{name: "Count1", mType: TypeCountConst, val: 1},
			expVal: 1,
		},
		tFloat: testVal[float64]{
			met:    MetricRepo[float64]{name: "Gauge1", mType: TypeGaugeConst, val: 1.1},
			expVal: 1.1,
		},
	}

	fVal(t, "counter", tc.tInt)
	fVal(t, "gauge", tc.tFloat)
}

func fVal[VT ValueType](t *testing.T, tName string, test testVal[VT]) {
	t.Run("valMetricRepo "+tName, func(t *testing.T) {
		assert.Equal(t, test.met.Value(), test.expVal)
	})
}

type testName[VT ValueType] struct {
	met     MetricRepo[VT]
	expName string
}

type caseName struct {
	tInt   testName[int64]
	tFloat testName[float64]
}

func TestNameMetricRepo(t *testing.T) {
	tc := caseName{
		tInt: testName[int64]{
			met:     MetricRepo[int64]{name: "Count1", mType: TypeCountConst, val: 1},
			expName: "Count1",
		},
		tFloat: testName[float64]{
			met:     MetricRepo[float64]{name: "Gauge1", mType: TypeGaugeConst, val: 1.1},
			expName: "Gauge1",
		},
	}

	fName(t, "counter", tc.tInt)
	fName(t, "gauge", tc.tFloat)
}

func fName[VT ValueType](t *testing.T, tName string, test testName[VT]) {
	t.Run("valMetricRepo "+tName, func(t *testing.T) {
		assert.Equal(t, test.met.Name(), test.expName)
	})
}

type testUpd[VT ValueType] struct {
	met    MetricRepo[VT]
	updVal VT
	exMet  MetricRepo[VT]
}

type caseUpd struct {
	tInt   testUpd[int64]
	tFloat testUpd[float64]
}

func TestUpdateMetricRepo(t *testing.T) {
	tc := caseUpd{
		tInt: testUpd[int64]{
			met:    MetricRepo[int64]{name: "Count1", mType: TypeCountConst, val: 1},
			updVal: 10,
			exMet:  MetricRepo[int64]{name: "Count1", mType: TypeCountConst, val: 11},
		},
		tFloat: testUpd[float64]{
			met:    MetricRepo[float64]{name: "Gauge1", mType: TypeGaugeConst, val: 1.1},
			updVal: 10.01,
			exMet:  MetricRepo[float64]{name: "Gauge1", mType: TypeGaugeConst, val: 10.01},
		},
	}

	fUpd(t, "counter", tc.tInt)
	fUpd(t, "gauge", tc.tFloat)
}

func fUpd[VT ValueType](t *testing.T, tName string, test testUpd[VT]) {
	t.Run("updMetricRepo "+tName, func(t *testing.T) {
		test.met.Update(test.updVal)
		assert.Equal(t, test.exMet, test.met)
	})
}

func TestParseMetricJSON(t *testing.T) {
	type testCase struct {
		name   string
		metStr MetricStr
		isErr  bool
		fnWant func() MetricJSON
	}

	tc := []testCase{
		{
			name: "ok counter",
			metStr: MetricStr{
				Info: Info{
					Name:  "Counter-1",
					MType: TypeCountConst.String(),
				},
				Val: "10",
			},
			isErr: false,
			fnWant: func() MetricJSON {
				val := int64(10)
				return MetricJSON{
					ID:    "Counter-1",
					MType: TypeCountConst.String(),
					Delta: &val,
				}
			},
		},

		{
			name: "ok gauge",
			metStr: MetricStr{
				Info: Info{
					Name:  "Gauge-1",
					MType: TypeGaugeConst.String(),
				},
				Val: "10.01",
			},
			isErr: false,
			fnWant: func() MetricJSON {
				val := float64(10.01)
				return MetricJSON{
					ID:    "Gauge-1",
					MType: TypeGaugeConst.String(),
					Value: &val,
				}
			},
		},

		{
			name: "no valid counter val",
			metStr: MetricStr{
				Info: Info{
					Name:  "Counter-1",
					MType: TypeCountConst.String(),
				},
				Val: "a10",
			},
			isErr: true,
			fnWant: func() MetricJSON {
				return MetricJSON{}
			},
		},

		{
			name: "no valid gauge val",
			metStr: MetricStr{
				Info: Info{
					Name:  "Gauge-1",
					MType: TypeGaugeConst.String(),
				},
				Val: "a10",
			},
			isErr: true,
			fnWant: func() MetricJSON {
				return MetricJSON{}
			},
		},

		{
			name: "not sypport val",
			metStr: MetricStr{
				Info: Info{
					Name:  "Gauge-1",
					MType: "-",
				},
				Val: "a10",
			},
			isErr: true,
			fnWant: func() MetricJSON {
				return MetricJSON{}
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			acrMet, err := ParseMetricJSON(test.metStr)
			if test.isErr && err == nil {
				t.Error("want err")
			}
			assert.Equal(t, test.fnWant(), acrMet)
		})
	}
}

func TestMetricJSONString(t *testing.T) {
	type testCase struct {
		name   string
		fnInit func() MetricJSON
		valStr string
	}

	tc := []testCase{
		{
			name: "ok counter",
			fnInit: func() MetricJSON {
				val := int64(10)
				return MetricJSON{
					ID:    "Counter-1",
					MType: TypeCountConst.String(),
					Delta: &val,
				}
			},
			valStr: "10",
		},

		{
			name: "ok gauge",
			fnInit: func() MetricJSON {
				val := float64(10.01)
				return MetricJSON{
					ID:    "Gauge-1",
					MType: TypeGaugeConst.String(),
					Value: &val,
				}
			},
			valStr: "10.01",
		},

		{
			name: "type not support",
			fnInit: func() MetricJSON {
				val := float64(10.01)
				return MetricJSON{
					ID:    "Gauge-1",
					MType: "--",
					Value: &val,
				}
			},
			valStr: TypeNoCorrect.String(),
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			met := test.fnInit()
			assert.Equal(t, test.valStr, met.String())
		})
	}
}

func TestBatchToArrMetricJSON(t *testing.T) {
	int1 := int64(10)
	int2 := int64(20)
	float1 := float64(10.01)
	float2 := float64(20.02)

	initBatch := Batch{
		CList: []MetricRepo[int64]{
			NewMetricRepo[int64](
				"Counter-1",
				TypeCountConst,
				int1,
			),
			NewMetricRepo[int64](
				"Counter-2",
				TypeCountConst,
				int2,
			),
		},
		GList: []MetricRepo[float64]{
			NewMetricRepo[float64](
				"Gauge-1",
				TypeGaugeConst,
				float1,
			),
			NewMetricRepo[float64](
				"Gauge-2",
				TypeGaugeConst,
				float2,
			),
		},
	}

	wantArr := []MetricJSON{
		{
			ID:    "Counter-1",
			MType: TypeCountConst.String(),
			Value: nil,
			Delta: &int1,
		},
		{
			ID:    "Counter-2",
			MType: TypeCountConst.String(),
			Value: nil,
			Delta: &int2,
		},
		{
			ID:    "Gauge-1",
			MType: TypeGaugeConst.String(),
			Value: &float1,
			Delta: nil,
		},
		{
			ID:    "Gauge-2",
			MType: TypeGaugeConst.String(),
			Value: &float2,
			Delta: nil,
		},
	}

	assert.Equal(t, wantArr, initBatch.ToArrMetricJSON())
}

func TestInfoValid(t *testing.T) {
	type testCase struct {
		name  string
		info  Info
		isErr bool
	}

	tc := []testCase{
		{
			name: "ok",
			info: Info{
				Name:  "Name",
				MType: "MType",
			},
			isErr: false,
		},

		{
			name: "err mtype empty",
			info: Info{
				Name:  "Name",
				MType: "",
			},
			isErr: true,
		},

		{
			name: "err name empty",
			info: Info{
				Name:  "",
				MType: "MType",
			},
			isErr: true,
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			err := test.info.Valid()

			if test.isErr && err == nil {
				t.Error("want err")
			}
		})
	}
}
