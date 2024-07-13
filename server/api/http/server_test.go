package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/AndreyVLZ/metrics/internal/store"
	"github.com/AndreyVLZ/metrics/server/config"
)

func TestHTTPServer(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	cfg := config.Config{
		Addr: "localhost:8081",
		StorageConfig: config.StorageConfig{
			StorePath: "",
		},
	}

	// h := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})
	srv := NewServer(&cfg, store.New(cfg.StorageConfig), slog.Default())

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
