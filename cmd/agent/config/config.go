package config

import "flag"

type Config struct {
	Addr           string
	ReportInterval int
	PollInterval   int
}

func New() *Config {
	conf := &Config{}
	flag.StringVar(&conf.Addr, "a", "localhost:8080", "doc-1")
	flag.IntVar(&conf.ReportInterval, "r", 10, "doc-2")
	flag.IntVar(&conf.PollInterval, "p", 2, "doc1-3")
	flag.Parse()

	return conf
}
