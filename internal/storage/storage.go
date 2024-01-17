package storage

import "github.com/AndreyVLZ/metrics/internal/metric"

type Storage interface {
	Get(metric.MetricDB) (metric.MetricDB, error)
	Set(metric.MetricDB) error
	List() []metric.MetricDB
}
