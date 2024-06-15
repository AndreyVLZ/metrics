// Адаптер для сервера.
package adapter

import "context"

// iServer интерфейс сервера.
type iServer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Shutdown адаптер для pkg.shutdown.StartStopper.
// Для 'встраиваемого' интерфейса iServer:
// - определяет новый метод Err,
// - переопределяет метод Start.
type Shutdown struct {
	iServer
	chErr chan error
}

// NewShutdown возвращает новый адаптер Shutdown для server'a.
func NewShutdown(server iServer) *Shutdown {
	return &Shutdown{
		iServer: server,
		chErr:   make(chan error),
	}
}

// Имплементация StartStopper.Err.
func (s *Shutdown) Err() <-chan error { return s.chErr }

// Имплементация StartStopper.Start.
func (s *Shutdown) Start(ctx context.Context) error {
	go func() {
		s.chErr <- s.iServer.Start(ctx)
	}()

	return nil
}
