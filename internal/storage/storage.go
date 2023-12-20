package storage

type Storage interface {
	Set(typeStr, name, valStr string) error
	Get(typeStr, name string) (string, error)
	GaugeRepo() Repository
	CounterRepo() Repository
}

type Repository interface {
	Set(name, valStr string) error
	Get(string) (string, error)
	List() map[string]string
}
