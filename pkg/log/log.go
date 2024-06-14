package log

import (
	"log/slog"

	myslog "github.com/AndreyVLZ/metrics/pkg/log/slog"
	"github.com/AndreyVLZ/metrics/pkg/log/zap"
	"go.uber.org/zap/zapcore"
)

type keyLog uint8

const (
	SlogKey keyLog = iota
	ZapLogKey
)

const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelErr   = "err"
)

func New(keyLog keyLog, lvl string) *slog.Logger {
	switch keyLog {
	case ZapLogKey:
		return zap.New(buildZapLevel(lvl))
	default:
		return myslog.New(buildSlogLevel(lvl))
	}
}

func buildSlogLevel(level string) slog.Level {
	switch level {
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelErr:
		return slog.LevelError
	default:
		return slog.LevelDebug
	}
}

func buildZapLevel(level string) zapcore.Level {
	switch level {
	case LevelInfo:
		return zapcore.InfoLevel
	case LevelWarn:
		return zapcore.WarnLevel
	case LevelErr:
		return zapcore.ErrorLevel
	default:
		return zapcore.DebugLevel
	}
}
