package shutdown

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeStarter struct {
	err   error
	chErr chan error
}

func (fs *fakeStarter) Start(ctx context.Context) error { return fs.err }
func (fs *fakeStarter) Stop(ctx context.Context) error  { return fs.err }

func (fs *fakeStarter) Err() <-chan error {
	return fs.chErr
}

func TestShutdown(t *testing.T) {
	ctx := context.Background()

	t.Run("start ok", func(t *testing.T) {
		ctxCan, cancel := context.WithCancel(ctx)
		timeout := time.Second
		fs := fakeStarter{chErr: make(chan error)}

		shut := New(&fs, timeout)
		cancel()
		if err := shut.Start(ctxCan); err != nil {
			if !errors.Is(context.Canceled, err) {
				t.Errorf("errs[%v]-[%v]\n", err, context.Canceled)
			}
		}
	})

	t.Run("start err", func(t *testing.T) {
		errCheck := errors.New("star err")
		ctxCan, cancel := context.WithCancel(ctx)
		timeout := time.Second
		fs := fakeStarter{chErr: make(chan error), err: errCheck}

		shut := New(&fs, timeout)
		cancel()
		if err := shut.Start(ctxCan); err != nil {
			if !errors.Is(errCheck, err) {
				t.Errorf("errs[%v]-[%v]\n", err, errCheck)
			}
		}
	})
}
