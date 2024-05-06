package agent

const (
	AddressDefault        = "localhost:8080"
	PollIntervalDefault   = 2
	ReportIntervalDefault = 10
	RateLimitDefault      = 3
	KeyDafault            = "SECRET_KEY"
)

type FuncOpt func(cfg *Config)

type Config struct {
	addr           string
	pollInterval   int
	reportInterval int
	rateLimit      int
	key            string
}

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

func SetAddr(addr string) FuncOpt {
	return func(cfg *Config) {
		cfg.addr = addr
	}
}

func SetPollInterval(pollInterval int) FuncOpt {
	return func(cfg *Config) {
		cfg.pollInterval = pollInterval
	}
}

func SetReportInterval(reportInterval int) FuncOpt {
	return func(cfg *Config) {
		cfg.reportInterval = reportInterval
	}
}

func SetKey(key string) FuncOpt {
	return func(cfg *Config) {
		cfg.key = key
	}
}

func SetRateLimit(rateLimit int) FuncOpt {
	return func(cfg *Config) {
		cfg.rateLimit = rateLimit
	}
}
