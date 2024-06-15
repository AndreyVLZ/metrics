package server

import "testing"

func TestNewConfig(t *testing.T) {
	type testCase struct {
		name    string
		fnOpt   FuncOpt
		fnCheck func(cfg Config) bool
	}

	tc := []testCase{
		{
			name:  "setAddr",
			fnOpt: SetAddr("address"),
			fnCheck: func(cfg Config) bool {
				return cfg.addr == "address"
			},
		},
		{
			name:  "setStoreInt",
			fnOpt: SetStoreInt(100),
			fnCheck: func(cfg Config) bool {
				return cfg.storeInt == 100
			},
		},
		{
			name:  "setStorePath",
			fnOpt: SetStorePath("storePath"),
			fnCheck: func(cfg Config) bool {
				return cfg.storePath == "storePath"
			},
		},
		{
			name:  "setRestore",
			fnOpt: SetRestore(true),
			fnCheck: func(cfg Config) bool {
				return cfg.isRestore == true
			},
		},
		{
			name:  "setDatabaseDNS",
			fnOpt: SetDatabaseDNS("databaseDNS"),
			fnCheck: func(cfg Config) bool {
				return cfg.dbDNS == "databaseDNS"
			},
		},
		{
			name:  "setKey",
			fnOpt: SetKey("key"),
			fnCheck: func(cfg Config) bool {
				return cfg.key == "key"
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			cfg := NewConfig(test.fnOpt)
			if !test.fnCheck(*cfg) {
				t.Fatalf("config %v\n", cfg)
			}
		})
	}
}
