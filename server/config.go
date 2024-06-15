package server

const (
	AddressDefault       string = "localhost:8081"       // Значение по умолчанию для адреса эндпоинта HTTP-сервера.
	StoreIntervalDefault int    = 300                    // Значение по умолчанию для интервала времени в секундах, по истечении которого текущие показания сервера сохраняются на диск.
	StoragePathDefault   string = "/tmp/metrics-db.json" // Значение по умолчанию для имени файла, куда сохраняются текущие значения.
	IsRestoreDefault     bool   = true                   // Значение по умолчанию для значения определяющее, загружать или нет ранее сохранённые значения из указанного файла при старте сервера.
)

// Конфиг для Сервера.
type config struct {
	addr      string
	storePath string
	dbDNS     string
	key       string
	storeInt  int
	isRestore bool
}

// Опции кофига.
type FuncOpt func(*config)

// Новый конфиг с установленными опциями.
func newConfig(opts ...FuncOpt) *config {
	cfg := config{
		addr:      AddressDefault,
		storeInt:  StoreIntervalDefault,
		storePath: StoragePathDefault,
		isRestore: IsRestoreDefault,
	}

	for i := range opts {
		opts[i](&cfg)
	}

	return &cfg
}

// Установка адреса эндпоинта.
func SetAddr(addr string) FuncOpt {
	return func(cfg *config) {
		cfg.addr = addr
	}
}

// Установка интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск.
func SetStoreInt(interval int) FuncOpt {
	return func(cfg *config) {
		cfg.storeInt = interval
	}
}

// Установка имени файла, куда сохраняются текущие значения.
func SetStorePath(path string) FuncOpt {
	return func(cfg *config) {
		cfg.storePath = path
	}
}

// Установка значения, определяющее загружать или нет ранее сохранённые значения из указанного файла при старте сервера.
func SetRestore(b bool) FuncOpt {
	return func(cfg *config) {
		cfg.isRestore = b
	}
}

// Установка строки с адресом подключения к БД.
func SetDatabaseDNS(dns string) FuncOpt {
	return func(cfg *config) {
		cfg.dbDNS = dns
	}
}

// Установка ключа.
func SetKey(key string) FuncOpt {
	return func(cfg *config) {
		cfg.key = key
	}
}
