package postgres

import (
	"database/sql"
	"errors"

	"github.com/AndreyVLZ/metrics/internal/metric"
	_ "github.com/lib/pq"
)

var (
	ErrTypeNotSupport = errors.New("not type support")
)

type postMetric struct {
	id    string
	mType string
	cVal  *int64
	gVal  *float64
}

func newPostMetricFromMetricDB(metricDB metric.MetricDB) (postMetric, error) {
	postMetric := postMetric{
		id:    metricDB.Name(),
		mType: metricDB.Type(),
	}

	switch metricDB.Type() {
	case metric.CounterType.String():
		postMetric.cVal = new(int64)
		metricDB.ReadTo(postMetric.cVal)
	case metric.GaugeType.String():
		postMetric.gVal = new(float64)
		metricDB.ReadTo(postMetric.gVal)
	}

	return postMetric, nil
}

func NewMetricDBFromPostMetric(postMetric postMetric) (metric.MetricDB, error) {
	var val metric.Valuer

	switch postMetric.mType {
	case metric.CounterType.String():
		if postMetric.cVal == nil {
			val = metric.Counter(0)
		} else {
			val = metric.Counter(*postMetric.cVal)
		}
	case metric.GaugeType.String():
		if postMetric.gVal == nil {
			val = metric.Gauge(0)
		} else {
			val = metric.Gauge(*postMetric.gVal)
		}
	default:
		return metric.MetricDB{}, ErrTypeNotSupport
	}

	return metric.NewMetricDB(postMetric.id, val), nil
}

type PostgresConfig struct {
	ConnDB string
}

type Postgres struct {
	db  *sql.DB
	cfg PostgresConfig
}

func New(config PostgresConfig) *Postgres {
	return &Postgres{
		cfg: config,
	}
}

func (store *Postgres) Close() error {
	err := store.db.Close()
	if err != nil {
		return err
	}

	return nil
}

func (store *Postgres) Open() error {
	db, err := sql.Open("postgres", store.cfg.ConnDB)
	if err != nil {
		return err
	}
	store.db = db

	return store.createTable()

}

func (store *Postgres) Ping() error {
	return store.db.Ping()
}
