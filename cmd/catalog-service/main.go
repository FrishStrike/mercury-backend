package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	catalogv1 "github.com/frishstrike/mercury-backend/api/proto/gen/go/catalog/v1"
	grpcd "github.com/frishstrike/mercury-backend/internal/catalog-service/delivery/grpc"
	"github.com/frishstrike/mercury-backend/internal/catalog-service/usecase"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	// Инициализация логгера
	logger := initLogger()

	// Инициализация зависимостей
	uc := usecase.NewMockProductUseCase()
	handler := grpcd.NewHandler(uc)

	// gRPC сервер с интерцептором для логирования
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcd.LoggerInterceptor(logger)),
	)

	catalogv1.RegisterCatalogServiceServer(grpcServer, handler)
	reflection.Register(grpcServer)

	// Запуск сервера
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Error("listen error", "error", err)
		return err
	}

	// Graceful shutdown
	return runServer(grpcServer, lis, logger)
}

func initLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func runServer(server *grpc.Server, lis net.Listener, logger *slog.Logger) error {
	errCh := make(chan error, 1)
	go func() {
		logger.Info("server started", "address", lis.Addr().String())
		errCh <- server.Serve(lis)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		logger.Info("shutdown signal received")
	case err := <-errCh:
		logger.Error("server error", "error", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(done)
	}()

	select {
	case <-ctx.Done():
		logger.Warn("shutdown timeout")
		server.Stop()
	case <-done:
		logger.Info("server stopped")
	}

	return nil
}
