package storage

import "github.com/AndreyVLZ/metrics/internal/metric"

type Storage interface {
	//Set(typeStr, name, valStr string) error
	//Get(typeStr, name string) (string, error)
	//GaugeRepo() Repository
	//CounterRepo() Repository

	Get(metric.MetricDB) (metric.MetricDB, error)
	Set(metric.MetricDB) error
	List() []metric.MetricDB
}
