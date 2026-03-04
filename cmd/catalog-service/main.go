package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("🚀 Catalog Service starting...")

	// Заглушка для будущего кода
	// Здесь будет:
	// - Инициализация конфига
	// - Подключение к БД
	// - gRPC сервер
	// - Redis кэш

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("🛑 Catalog Service shutting down...")
}
