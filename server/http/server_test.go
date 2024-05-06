package http

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	type testCase struct {
		name  string
		cfg   Config
		isErr bool
	}

	tc := []testCase{
		{
			name:  "ok",
			cfg:   Config{Addr: "localhost:8080"},
			isErr: false,
		},
		{
			name:  "not valid address",
			cfg:   Config{Addr: "l:8080"},
			isErr: true,
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			h := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

			srv := NewServer(test.cfg, h)

			ctx := context.Background()
			ctxTime, stopTime := context.WithTimeout(ctx, 2*time.Second)

			chErr := make(chan error)
			go func(ce chan<- error) {
				defer close(ce)
				ce <- srv.Start()
			}(chErr)

			var err error
			select {
			case <-ctxTime.Done():
			case er := <-chErr:
				err = er
			}
			stopTime()

			if test.isErr && err == nil {
				t.Error("want err")
			}

			if err := srv.Stop(ctx); err != nil {
				t.Errorf("stop err: %v\n", err)
			}
		})
	}
}
