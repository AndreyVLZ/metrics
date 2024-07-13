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
			fnOpt: SetAddr("address"),
			fnCheck: func(cfg Config) bool {
				return cfg.Addr == "address"
			},
		},
		{
			name:  "setStoreInt",
			fnOpt: SetStoreInt(100),
			fnCheck: func(cfg Config) bool {
				return cfg.StoreInt == 100
			},
		},
		{
			name:  "setStorePath",
			fnOpt: SetStorePath(""),
			fnCheck: func(cfg Config) bool {
				return cfg.StorePath == ""
			},
		},
		{
			name:  "setRestore",
			fnOpt: SetRestore(true),
			fnCheck: func(cfg Config) bool {
				return cfg.IsRestore == true
			},
		},
		{
			name:  "setDatabaseDNS",
			fnOpt: SetDatabaseDSN("databaseDNS"),
			fnCheck: func(cfg Config) bool {
				return cfg.ConnDB == "databaseDNS"
			},
		},
		{
			name:  "setKey",
			fnOpt: SetKey("key"),
			fnCheck: func(cfg Config) bool {
				return cfg.Key == "key"
			},
		},
		{
			name:  "setConfigPath",
			fnOpt: SetConfigPath("configPath"),
			fnCheck: func(cfg Config) bool {
				return cfg.ConfigPath == "configPath"
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
			var cfg Config

			test.fnOpt(&cfg)

			if !test.fnCheck(cfg) {
				t.Fatalf("config %v\n", cfg)
			}
		})
	}
}
