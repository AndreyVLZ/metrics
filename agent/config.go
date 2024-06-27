package agent

const (
	AddressDefault        = "localhost:8080" // Значение по умолчанию для адреса эндпоинта HTTP-сервера.
	PollIntervalDefault   = 2                // Значение по умолчания для частоты опроса метрик из пакета runtime.
	ReportIntervalDefault = 10               // Значение по умолчания для частоты отправки метрик на сервер.
	RateLimitDefault      = 3                // Значение по умолчания для количества одновременно исходящих запросов на сервер.
	// KeyDafault            = "SECRET_KEY"      // Значение по умолчания для ключа.
	CryproKeyPathDefault = "/tmp/public.pem" // Значение по умолчания для пути до файла с публичным ключом
)

/*
// Опции для конфига.
type FuncOpt func(cfg *config)

// Конфиг для агента.
type config struct {
	addr           string
	key            string
	pollInterval   int
	reportInterval int
	rateLimit      int
	cryptoKeyPath  string
	publicKey      *rsa.PublicKey
}

// Новый конфиг с установленными опциями.
func newConfig(opts ...FuncOpt) *config {
	cfg := &config{
		addr:           AddressDefault,
		pollInterval:   PollIntervalDefault,
		reportInterval: ReportIntervalDefault,
		rateLimit:      RateLimitDefault,
		key:            KeyDafault,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// Установка адреса эндпоинта HTTP-сервера.
func SetAddr(addr string) FuncOpt {
	return func(cfg *config) {
		cfg.addr = addr
	}
}

// Установка частоты опроса метрик из пакета runtime.
func SetPollInterval(pollInterval int) FuncOpt {
	return func(cfg *config) {
		cfg.pollInterval = pollInterval
	}
}

// Установка частоты отправки метрик на сервер.
func SetReportInterval(reportInterval int) FuncOpt {
	return func(cfg *config) {
		cfg.reportInterval = reportInterval
	}
}

// Установка ключа.
func SetKey(key string) FuncOpt {
	return func(cfg *config) {
		cfg.key = key
	}
}

// Установка количества одновременно исходящих запросов на сервер.
func SetRateLimit(rateLimit int) FuncOpt {
	return func(cfg *config) {
		cfg.rateLimit = rateLimit
	}
}

// Установка пути до файла с публичным ключом.
func SetCryptoKeyPath(cryptoKeyPath string) FuncOpt {
	return func(cfg *config) {
		cfg.cryptoKeyPath = cryptoKeyPath
	}
}
*/
