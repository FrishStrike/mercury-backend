package grpc

import (
	"context"
	"log/slog"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoggerInterceptor — логирование всех gRPC запросов
func LoggerInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		statusCode := codes.OK
		if err != nil {
			statusCode = status.Code(err)
		}

		logger.Info("grpc request",
			"method", info.FullMethod,
			"duration_ms", duration.Milliseconds(),
			"status", statusCode.String(),
		)

		return resp, err
	}
}

// RecoveryInterceptor — перехват паник
func RecoveryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered",
					"error", r,
					"stack", string(debug.Stack()),
				)
			}
		}()
		return handler(ctx, req)
	}
}
