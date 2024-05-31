package server

const (
	AddressDefault       string = "localhost:8080"
	StoreIntervalDefault int    = 300
	StoragePathDefault   string = "/tmp/metrics-db.json"
	IsRestoreDefault     bool   = true
)

type Config struct {
	addr      string // api
	storeInt  int    // fileStore
	storePath string // fileStore
	isRestore bool   // fileStore
	dbDNS     string // postgers
	key       string
}

type FuncOpt func(*Config)

func NewConfig(opts ...FuncOpt) *Config {
	cfg := Config{
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

func SetAddr(addr string) FuncOpt {
	return func(cfg *Config) {
		cfg.addr = addr
	}
}

func SetStoreInt(interval int) FuncOpt {
	return func(cfg *Config) {
		cfg.storeInt = interval
	}
}

func SetStorePath(path string) FuncOpt {
	return func(cfg *Config) {
		cfg.storePath = path
	}
}

func SetRestore(b bool) FuncOpt {
	return func(cfg *Config) {
		cfg.isRestore = b
	}
}

func SetDatabaseDNS(dns string) FuncOpt {
	return func(cfg *Config) {
		cfg.dbDNS = dns
	}
}

func SetKey(key string) FuncOpt {
	return func(cfg *Config) {
		cfg.key = key
	}
}
