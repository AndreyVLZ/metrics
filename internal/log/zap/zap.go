package zap

import (
	"log/slog"
	_ "net/http/pprof"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
)

func DefaultConfig() zap.Config {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	//encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeTime = zapcore.TimeEncoder(func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("02.01.06 15:04:05"))
	})

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     encoderCfg,
		OutputPaths: []string{
			"stderr",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
	}

	return config
}

func New(conf zap.Config) *slog.Logger {
	zapL := newZap(conf)
	logger := slog.New(zapslog.NewHandler(zapL.Core(), nil))

	return logger
}

func newZap(conf zap.Config) *zap.Logger {
	zapL := zap.Must(conf.Build())
	defer zapL.Sync()

	return zapL
}
