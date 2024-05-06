package agent

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
			fnOpt: SetAddr("addres"),
			fnCheck: func(cfg Config) bool {
				return cfg.addr == "addres"
			},
		},
		{
			name:  "setPollInterval",
			fnOpt: SetPollInterval(100),
			fnCheck: func(cfg Config) bool {
				return cfg.pollInterval == 100
			},
		},
		{
			name:  "setReportInterval",
			fnOpt: SetReportInterval(100),
			fnCheck: func(cfg Config) bool {
				return cfg.reportInterval == 100
			},
		},
		{
			name:  "setRateLimit",
			fnOpt: SetRateLimit(100),
			fnCheck: func(cfg Config) bool {
				return cfg.rateLimit == 100
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
