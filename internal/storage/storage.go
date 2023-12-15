package storage

type Storage interface {
	Set(name, typeStr, valStr string) error
	//Repo(mType metric.MetricType) (Repository, error)
	GaugeRepo() Repository
	CounterRepo() Repository
}

type Repository interface {
	Set(name, valStr string) error
	Get(string) (string, error)
	List() map[string]string

	// NOTE: Delete
	Range()
}
