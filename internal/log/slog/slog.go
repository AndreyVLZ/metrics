package slog

import (
	"log/slog"
	_ "net/http/pprof"
	"os"
)

func New() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	return logger
}
