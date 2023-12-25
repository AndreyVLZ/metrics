package metricagent

import (
	"testing"

	"github.com/AndreyVLZ/metrics/internal/storage"
	"github.com/AndreyVLZ/metrics/internal/storage/memstorage"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	testCase := []struct {
		name   string
		opts   []FuncOpt
		addr   string
		poll   int
		report int
	}{
		{
			name:   "positive #1",
			opts:   []FuncOpt{},
			addr:   AddressDefault,
			poll:   PollIntervalDefault,
			report: ReportIntervalDefault,
		},
		{
			name: "positive #2",
			opts: []FuncOpt{
				SetAddr("set"),
				SetPollInterval(777),
				SetReportInterval(888),
			},
			addr:   "set",
			poll:   777,
			report: 888,
		},
		{
			name: "positive #3",
			opts: []FuncOpt{
				SetAddr("not-set"),
				SetPollInterval(77),
				SetReportInterval(88),
				SetAddr("set"),
				SetPollInterval(777),
				SetReportInterval(888),
			},
			addr:   "set",
			poll:   777,
			report: 888,
		},
	}
	for _, test := range testCase {
		agent := New(test.opts...)
		assert.Equal(t, test.addr, agent.addr)
		assert.Equal(t, test.poll, agent.pollInterval)
		assert.Equal(t, test.report, agent.reportInterval)

	}
}

type testStore struct {
	typeStr string
	name    string
	val     string
}

func (s *testStore) Set(typeStr, name, valStr string) error {
	s.typeStr = typeStr
	s.name = name
	s.val = valStr
	return nil
}

func (s *testStore) Get(typeStr, name string) (string, error) {
	if s.typeStr != typeStr || s.name != name {
		return "", memstorage.ErrValueByNameNotFound
	}
	return s.val, nil
}
func (s *testStore) GaugeRepo() storage.Repository {
	return nil
}
func (s *testStore) CounterRepo() storage.Repository {
	return nil
}

func TestAddMetric(t *testing.T) {
	testCase := []struct {
		name    string
		nameStr string
		typeStr string
		valStr  string
	}{
		{
			name:    "positive #1",
			nameStr: "nameStr",
			typeStr: "typeStr",
			valStr:  "valStr",
		},
		{
			name:    "negative #1",
			nameStr: "nameStr",
			typeStr: "typeStr",
			valStr:  "valStr",
		},
	}
	for _, test := range testCase {
		t.Run(test.name, func(t *testing.T) {
			testStore := &testStore{}
			agent := MetricClient{store: testStore}
			agent.AddMetric(test.typeStr, test.nameStr, test.valStr)

			assert.Equal(t, testStore.typeStr, test.typeStr)
			assert.Equal(t, testStore.name, test.nameStr)
			assert.Equal(t, testStore.val, test.valStr)
		})
	}
}
