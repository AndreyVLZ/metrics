// Адаптер для сервера.
package adapter

import "context"

// Интнрфейс для сервера.
type iServer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Адаптер для shutdown.
type Shutdown struct {
	iServer
	chErr chan error
}

func NewShutdown(server iServer) *Shutdown {
	return &Shutdown{
		iServer: server,
		chErr:   make(chan error),
	}
}

// Имплементация shutdown.Err().
func (s *Shutdown) Err() <-chan error { return s.chErr }

// Имплементация shutdown.Start().
func (s *Shutdown) Start(ctx context.Context) error {
	go func() {
		s.chErr <- s.iServer.Start(ctx)
	}()

	return nil
}
