package grps

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
)

// Logging Логирование исходящего gprc.
func Logging(log *slog.Logger) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context, method string, req interface{},
		reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if err := invoker(ctx, method, req, reply, cc, opts...); err != nil {
			log.ErrorContext(ctx, "grpc", "err", err, "method", method)

			return err
		}

		log.InfoContext(ctx, "grpc", "method", method)

		return nil
	}
}
