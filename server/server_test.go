package server

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"
)

func TestStartStop(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctxTime, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cfg := NewConfig()
	log := slog.Default()

	srv := New(cfg, log)

	errc := make(chan error)
	go func(erc chan error) {
		if err := srv.Start(ctxTime); err != nil {
			errc <- err
		}
	}(errc)

	select {
	case <-ctxTime.Done():
		fmt.Println("timeOut")
	case err := <-errc:
		if err != nil {
			t.Errorf("err %v\n", err)
		}
	}

	if err := srv.Stop(ctx); err != nil {
		t.Errorf("err %v\n", err)
	}
}
