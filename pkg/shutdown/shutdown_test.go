package shutdown

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"
)

type fakeStarter struct {
	errStart error
	errStop  error
	errRun   error
	chErr    chan error
}

func (fs *fakeStarter) Start(_ context.Context) error { return fs.errStart }
func (fs *fakeStarter) Stop(_ context.Context) error  { return fs.errStop }

func (fs *fakeStarter) Err() <-chan error {
	if fs.errRun != nil {
		fs.chErr <- fs.errRun
	}

	return fs.chErr
}

func TestShutdown(t *testing.T) {
	ctx := context.Background()

	t.Run("start ok", func(t *testing.T) {
		ctxCan, cancel := context.WithCancel(ctx)

		fs := fakeStarter{chErr: make(chan error)}
		shut := New(&fs, time.Second)

		cancel()

		signals := []os.Signal{
			syscall.SIGTERM,
			syscall.SIGINT,
			syscall.SIGQUIT,
		}

		if err := shut.Start(ctxCan, signals...); err != nil {
			if !errors.Is(err, context.Canceled) {
				t.Errorf("errs [%v]!=[%v]\n", err, context.Canceled)
			}
		}
	})

	t.Run("start err", func(t *testing.T) {
		ctxCan, cancel := context.WithCancel(ctx)

		errCheck := errors.New("star err")
		fs := fakeStarter{chErr: make(chan error), errStart: errCheck}
		shutdown := New(&fs, time.Second)

		cancel()

		signals := []os.Signal{
			syscall.SIGTERM,
			syscall.SIGINT,
			syscall.SIGQUIT,
		}

		if err := shutdown.Start(ctxCan, signals...); err != nil {
			if !errors.Is(err, errCheck) {
				t.Errorf("errs [%v]!=[%v]\n", err, errCheck)
			}
		}
	})

	t.Run("stop err", func(t *testing.T) {
		ctxCan, cancel := context.WithCancel(ctx)

		errCheck := errors.New("stop err")
		fs := fakeStarter{chErr: make(chan error), errStop: errCheck}

		shutdown := New(&fs, time.Second)

		cancel()
		signals := []os.Signal{
			syscall.SIGTERM,
			syscall.SIGINT,
			syscall.SIGQUIT,
		}

		if err := shutdown.Start(ctxCan, signals...); err != nil {
			if !errors.Is(err, errCheck) {
				t.Errorf("errs [%v]!=[%v]\n", err, errCheck)
			}
		}
	})

	t.Run("run err", func(t *testing.T) {
		ctxCan, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		errCheck := errors.New("run err")
		fs := fakeStarter{chErr: make(chan error, 1), errRun: errCheck}
		shutdown := New(&fs, time.Second)

		signals := []os.Signal{
			syscall.SIGTERM,
			syscall.SIGINT,
			syscall.SIGQUIT,
		}

		if err := shutdown.Start(ctxCan, signals...); err != nil {
			if !errors.Is(err, errCheck) {
				t.Errorf("errs [%v]!=[%v]\n", err, errCheck)
			}
		}
	})
}
