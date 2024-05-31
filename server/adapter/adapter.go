package adapter

import "context"

type iServer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

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

func (s *Shutdown) Err() <-chan error { return s.chErr }

func (s *Shutdown) Start(ctx context.Context) error {
	go func() {
		s.chErr <- s.iServer.Start(ctx)
	}()

	return nil
}
