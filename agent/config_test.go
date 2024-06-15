package agent

import "testing"

func TestNewConfig(t *testing.T) {
	type testCase struct {
		fnOpt   FuncOpt
		fnCheck func(cfg config) bool
		name    string
	}

	tc := []testCase{
		{
			name:  "setAddr",
			fnOpt: SetAddr("addres"),
			fnCheck: func(cfg config) bool {
				return cfg.addr == "addres"
			},
		},
		{
			name:  "setPollInterval",
			fnOpt: SetPollInterval(100),
			fnCheck: func(cfg config) bool {
				return cfg.pollInterval == 100
			},
		},
		{
			name:  "setReportInterval",
			fnOpt: SetReportInterval(100),
			fnCheck: func(cfg config) bool {
				return cfg.reportInterval == 100
			},
		},
		{
			name:  "setRateLimit",
			fnOpt: SetRateLimit(100),
			fnCheck: func(cfg config) bool {
				return cfg.rateLimit == 100
			},
		},
		{
			name:  "setKey",
			fnOpt: SetKey("key"),
			fnCheck: func(cfg config) bool {
				return cfg.key == "key"
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			cfg := newConfig(test.fnOpt)
			if !test.fnCheck(*cfg) {
				t.Fatalf("config %v\n", cfg)
			}
		})
	}
}
