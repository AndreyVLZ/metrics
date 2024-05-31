package agent

const (
	AddressDefault        = "localhost:8080" // Значение по умолчанию для адреса эндпоинта HTTP-сервера.
	PollIntervalDefault   = 2                // Значение по умолчания для частоты опроса метрик из пакета runtime.
	ReportIntervalDefault = 10               // Значение по умолчания для частоты отправки метрик на сервер.
	RateLimitDefault      = 3                // Значение по умолчания для количества одновременно исходящих запросов на сервер.
	KeyDafault            = "SECRET_KEY"     // Значение по умолчания для ключа.
)

// Опции для конфига.
type FuncOpt func(cfg *Config)

// Конфиг для агента.
type Config struct {
	addr           string
	pollInterval   int
	reportInterval int
	rateLimit      int
	key            string
}

// Новый конфиг с установленными опциями.
func NewConfig(opts ...FuncOpt) *Config {
	cfg := &Config{
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
	return func(cfg *Config) {
		cfg.addr = addr
	}
}

// Установка частоты опроса метрик из пакета runtime.
func SetPollInterval(pollInterval int) FuncOpt {
	return func(cfg *Config) {
		cfg.pollInterval = pollInterval
	}
}

// Установка частоты отправки метрик на сервер.
func SetReportInterval(reportInterval int) FuncOpt {
	return func(cfg *Config) {
		cfg.reportInterval = reportInterval
	}
}

// Установка ключа.
func SetKey(key string) FuncOpt {
	return func(cfg *Config) {
		cfg.key = key
	}
}

// Установка количества одновременно исходящих запросов на сервер.
func SetRateLimit(rateLimit int) FuncOpt {
	return func(cfg *Config) {
		cfg.rateLimit = rateLimit
	}
}
