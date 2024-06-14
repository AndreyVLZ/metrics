package http

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestHTTPServer(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	cfg := Config{Addr: "localhost:8081"}
	h := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})
	srv := NewServer(cfg, h)

	go func() {
		if err := srv.Start(); err != nil && !errors.Is(http.ErrServerClosed, err) {
			t.Errorf("httpServer start err: %v\n", err)
		}
	}()

	time.Sleep(time.Second)

	ctx := context.Background()

	ctxTimeout, cancelTimeout := context.WithTimeout(ctx, 2*time.Second)
	defer cancelTimeout()

	if err := srv.Stop(ctxTimeout); err != nil {
		t.Errorf("httpServer stop err: %v\n", err)
	}
}
