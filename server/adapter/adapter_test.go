package adapter

import (
	"context"
	"testing"
)

type fakeServer struct {
	err error
}

func (fs *fakeServer) Start(ctx context.Context) error { return fs.err }
func (fs *fakeServer) Stop(ctx context.Context) error  { return fs.err }

func TestAdapterShutdown(t *testing.T) {
	srv := fakeServer{}
	shut := NewShutdown(&srv)

	ctx, cancel := context.WithCancel(context.Background())

	if err := shut.Start(ctx); err != nil {
		t.Errorf("want not error")
	}

	cancel()

	err := <-shut.Err()
	if err != nil {
		t.Errorf("errs[%v]-[%v]\n", context.Canceled, err)
	}
}
