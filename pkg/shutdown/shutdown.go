// Start запускает StartStopper.
// Останавливает StartStoper по сигналу os или при возникновении ошибки.
// Возвращает ошибку:
// - возникшую при старте StartStopper.
// - c причиной остановки.
// Возможные причины остановки:
// - получен сингал os. [возвращает context.Canceled].
// - возникла ошибка при работе StartStopper [<-Err()].
// Ошибка с причиной остановки объедняется [errors.Join]
// c ошибкой при остановки StartStopper.
// Интересующую ошибку [errTarget] можно проверить:
// errors.Is(errShutdown, errTarget).
package shutdown

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"time"
)

// StartStoper интерфейс, для которого реализуется завершение по сигналу os.
type StartStoper interface {
	Start(ctx context.Context) error
	Err() <-chan error
	Stop(ctx context.Context) error
}

// Shutdown включает интерфейс StartStoper и таймаут для остановки.
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

// Start запускает Shutdown.
func (s *Shutdown) Start(ctx context.Context) error {
	var err error

	ctxCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	// Контекст для сигнала os
	ctxSinal, stopSignal := signal.NotifyContext(ctx, os.Interrupt)
	defer stopSignal()

	if err = s.delegate.Start(ctxCancel); err != nil {
		return err
	}

	select {
	case err = <-s.delegate.Err():
	case <-ctxSinal.Done():
		err = ctxSinal.Err()
	}

	cancel()

	return errors.Join(err, s.delegateStop(ctx))
}

func (s *Shutdown) delegateStop(ctx context.Context) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	return s.delegate.Stop(ctxTimeout)
}
