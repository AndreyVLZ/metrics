package metricagent

import (
	"testing"

	"github.com/AndreyVLZ/metrics/internal/metric"
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
	//val     string
	val metric.Valuer
}

func (s *testStore) Set(metric metric.MetricDB) error {
	s.typeStr = metric.Type()
	s.name = metric.Name()
	s.val = metric.Valuer
	return nil
}

func (s *testStore) Get(metricDB metric.MetricDB) (metric.MetricDB, error) {
	if s.typeStr != metricDB.Type() || s.name != metricDB.Name() {
		return metric.MetricDB{}, memstorage.ErrValueByNameNotFound
	}
	return metric.NewMetricDB(metricDB.Name(), metricDB.Valuer), nil
}

func (s *testStore) List() []metric.MetricDB {
	return []metric.MetricDB{}
}
