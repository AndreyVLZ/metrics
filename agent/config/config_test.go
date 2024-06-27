package config

import "testing"

func TestNewConfig(t *testing.T) {
	type testCase struct {
		fnOpt   FuncOpt
		fnCheck func(cfg Config) bool
		name    string
	}

	tc := []testCase{
		{
			name:  "setAddr",
			fnOpt: SetAddr("addres"),
			fnCheck: func(cfg Config) bool {
				return cfg.Addr == "addres"
			},
		},
		{
			name:  "setPollInterval",
			fnOpt: SetPollInterval(100),
			fnCheck: func(cfg Config) bool {
				return cfg.PollInterval == 100
			},
		},
		{
			name:  "setReportInterval",
			fnOpt: SetReportInterval(100),
			fnCheck: func(cfg Config) bool {
				return cfg.ReportInterval == 100
			},
		},
		{
			name:  "setRateLimit",
			fnOpt: SetRateLimit(100),
			fnCheck: func(cfg Config) bool {
				return cfg.RateLimit == 100
			},
		},
		{
			name:  "setKey",
			fnOpt: SetKey("key"),
			fnCheck: func(cfg Config) bool {
				return string(cfg.Key) == "key"
			},
		},
		{
			name:  "setCryptoKeyPath",
			fnOpt: SetCryptoKeyPath(""),
			fnCheck: func(cfg Config) bool {
				return cfg.CryptoKeyPath == ""
			},
		},
		{
			name:  "setLogLevel",
			fnOpt: SetLogLevel("logLevel"),
			fnCheck: func(cfg Config) bool {
				return cfg.LogLevel == "logLevel"
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			cfg, err := New(test.fnOpt)
			if err != nil {
				t.Errorf("new config: %v\n", err)
			}

			if !test.fnCheck(*cfg) {
				t.Fatalf("config %v\n", cfg)
			}
		})
	}
}
