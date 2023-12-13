package config

type Config struct {
	Addr           string
	ReportInterval int
	PollInterval   int
}

func New() *Config {
	conf := &Config{}

	return conf
}
