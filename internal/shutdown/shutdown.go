// Отстановка StartStoper по сигналу os.Interrupt.
package shutdown

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"
)

type StartStoper interface {
	Start(ctx context.Context) error
	Err() <-chan error
	Stop(ctx context.Context) error
}

type Shutdown struct {
	delegate StartStoper
	timeout  time.Duration
}

func New(delegate StartStoper, timeout time.Duration) *Shutdown {
	return &Shutdown{
		delegate: delegate,
		timeout:  timeout,
	}
}

func (s *Shutdown) Start(ctx context.Context) error {
	ctxSinal, stopSignal := signal.NotifyContext(ctx, os.Interrupt)
	defer stopSignal()

	if err := s.delegate.Start(ctxSinal); err != nil {
		return err
	}

	select {
	case err := <-s.delegate.Err():
		return err
	case <-ctxSinal.Done():
		fmt.Println("-Singal-")

		return s.delegateStop(ctx)
	}
}

func (s *Shutdown) delegateStop(ctx context.Context) error {
	chErr := make(chan error)

	ctxTimeout, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	go func(errc chan error) {
		defer close(chErr)

		if err := s.delegate.Stop(ctxTimeout); err != nil {
			errc <- err
		}
	}(chErr)

	select {
	case <-ctxTimeout.Done():
		return ctx.Err()
	case err := <-chErr:
		return err
	}
}
