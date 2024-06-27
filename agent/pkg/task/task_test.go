package task

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		poll := NewPoll(2, slog.Default())

		poll.Add(
			New("task-1",
				1*time.Second,
				func() error { return nil },
			),

			New("task-2",
				1*time.Second,
				func() error { return nil },
			),
		)

		ctx := context.Background()

		ctxTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		chErr := poll.Run(ctxTimeout)

		for err := range chErr {
			if err != nil {
				t.Error(err)
			}
		}
	})

	t.Run("whit err", func(t *testing.T) {
		poll := NewPoll(2, slog.Default())

		myErr := errors.New("my err")

		poll.Add(
			New("task-1",
				1*time.Second,
				func() error { return myErr },
			),
			New("task-2",
				1*time.Second,
				func() error { return nil },
			),
		)

		ctx := context.Background()

		ctxTimeout, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		chErr := poll.Run(ctxTimeout)

		for err := range chErr {
			cancel()
			assert.Equal(t, myErr, err)
		}
	})
}
