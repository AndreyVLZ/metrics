package storage

type Storage interface {
	Set(name, typeStr, valStr string) error
}
