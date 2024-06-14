package slog

import (
	"log/slog"
	"os"
)

func New(lvl slog.Level) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: lvl,
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))

	return logger
}
