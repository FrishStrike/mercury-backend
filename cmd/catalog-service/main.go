package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	catalogv1 "github.com/frishstrike/mercury-backend/api/proto/gen/go/catalog/v1"
	grpcd "github.com/frishstrike/mercury-backend/internal/catalog-service/delivery/grpc"
	"github.com/frishstrike/mercury-backend/internal/catalog-service/repository/postgres"
	"github.com/frishstrike/mercury-backend/internal/catalog-service/usecase"
	"github.com/frishstrike/mercury-backend/pkg/database"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	"google.golang.org/grpc/reflection"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	logger := initLogger()

	dbPool, err := initDatabase(logger)
	if err != nil {
		logger.Error("failed to initialize database", "error", err)
		return err
	}
	defer dbPool.Close()

	logger.Info("database connected")

	// Инициализация зависимостей
	repo := postgres.NewProductRepository(dbPool)
	uc := usecase.NewProductUseCase(repo)
	handler := grpcd.NewHandler(uc)

	// gRPC сервер (для межсервисного общения)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcd.LoggerInterceptor(logger)),
	)
	catalogv1.RegisterCatalogServiceServer(grpcServer, handler)
	reflection.Register(grpcServer)

	// Запуск gRPC сервера
	grpcLis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Error("gRPC listen error", "error", err)
		return err
	}

	go func() {
		logger.Info("gRPC server started", "address", grpcLis.Addr().String())
		if err := grpcServer.Serve(grpcLis); err != nil {
			logger.Error("gRPC server error", "error", err)
		}
	}()

	// HTTP Gateway (для REST API)
	if err := runHTTPGateway(logger, handler); err != nil {
		return err
	}

	// Graceful shutdown
	return waitForShutdown(logger, grpcServer)
}

func initLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func initDatabase(logger *slog.Logger) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg := database.Config{
		Host:     getEnv("POSTGRES_HOST", "localhost"),
		Port:     getEnv("POSTGRES_PORT", "5432"),
		User:     getEnv("POSTGRES_USER", "mercury"),
		Password: getEnv("POSTGRES_PASSWORD", "mercury_secret_2024"),
		Database: getEnv("POSTGRES_DB", "mercury"),
		SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
	}

	logger.Info("connecting to database",
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.Database,
	)

	return database.NewPostgresPool(ctx, cfg)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func runHTTPGateway(logger *slog.Logger, grpcHandler *grpcd.Handler) error {
	ctx := context.Background()

	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(runtime.DefaultHTTPErrorHandler),
		runtime.WithForwardResponseOption(func(ctx context.Context, w http.ResponseWriter, msg proto.Message) error {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			return nil
		}),
	)

	if err := catalogv1.RegisterCatalogServiceHandlerServer(ctx, mux, grpcHandler); err != nil {
		logger.Error("failed to register gateway", "error", err)
		return err
	}

	httpMux := http.NewServeMux()
	httpMux.Handle("/", mux)
	httpMux.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir("swagger"))))
	httpMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Middleware для логирования реального IP
	withLogging := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Получаем реальный IP из заголовков Nginx
			realIP := r.Header.Get("X-Real-IP")
			if realIP == "" {
				realIP = r.Header.Get("X-Forwarded-For")
			}
			if realIP == "" {
				realIP = r.RemoteAddr
			}

			logger.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"ip", realIP,
			)

			next.ServeHTTP(w, r)
		})
	}

	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      withLogging(httpMux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("HTTP gateway started", "address", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
		}
	}()

	return nil
}
func waitForShutdown(logger *slog.Logger, grpcServer *grpc.Server) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-ctx.Done():
		logger.Warn("shutdown timeout")
		grpcServer.Stop()
	case <-done:
		logger.Info("server stopped")
	}

	return nil
}
