package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	authhttp "github.com/frishstrike/mercury-backend/internal/auth-service/delivery/http"
	"github.com/frishstrike/mercury-backend/internal/auth-service/repository/postgres"
	"github.com/frishstrike/mercury-backend/internal/auth-service/usecase"
	"github.com/frishstrike/mercury-backend/pkg/auth"
	"github.com/frishstrike/mercury-backend/pkg/database"
	redispkg "github.com/frishstrike/mercury-backend/pkg/redis"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbPool, err := database.NewPostgresPool(ctx, database.Config{
		Host:     getEnv("POSTGRES_HOST", "localhost"),
		Port:     getEnv("POSTGRES_PORT", "5432"),
		User:     getEnv("POSTGRES_USER", "mercury"),
		Password: getEnv("POSTGRES_PASSWORD", "mercury_secret_2024"),
		Database: getEnv("POSTGRES_DB", "mercury"),
		SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
	})
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		return err
	}
	defer dbPool.Close()
	logger.Info("database connected")

	// Redis
	redisClient, err := redispkg.NewClient(ctx, redispkg.Config{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     getEnv("REDIS_PORT", "6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
	})
	if err != nil {
		logger.Error("failed to connect to redis", "error", err)
		return err
	}
	defer redisClient.Close()
	logger.Info("redis connected")

	// Dependencies
	jwtManager := auth.NewManager(getEnv("JWT_SECRET", "mercury-secret-key-change-in-production"))
	tokenStore := auth.NewTokenStore(redisClient)
	userRepo := postgres.NewUserRepository(dbPool)
	authUC := usecase.NewAuthUseCase(userRepo, jwtManager, tokenStore)
	handler := authhttp.NewHandler(authUC, jwtManager)

	// HTTP server
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"auth"}`))
	})

	httpServer := &http.Server{
		Addr:         ":" + getEnv("AUTH_HTTP_PORT", "8081"),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("auth service started", "address", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server error", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", "error", err)
		return err
	}

	logger.Info("auth service stopped")
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
