package storage

type Storage interface {
	Set(name, typeStr, valStr string) error
	Get(name, typeStr string) (string, error)
	//Repo(mType metric.MetricType) (Repository, error)
	GaugeRepo() Repository
	CounterRepo() Repository
	//List() map[string]string
}

type Repository interface {
	Set(name, valStr string) error
	Get(string) (string, error)
	List() map[string]string

	// NOTE: Delete
	Range()
}
