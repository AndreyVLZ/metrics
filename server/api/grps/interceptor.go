package grps

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
)

// Logging Логирование входящего grpc.
func Logging(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		log.InfoContext(ctx, "grpc", "method", info.FullMethod)

		m, err := handler(ctx, req)
		if err != nil {
			log.ErrorContext(ctx, "grpc", "err", err, "method", info.FullMethod)
		}

		return m, err
	}
}
