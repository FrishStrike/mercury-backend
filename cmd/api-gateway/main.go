package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("🚀 API Gateway starting...")

	// TODO: инициализация конфига
	// TODO: настройка HTTP сервера
	// TODO: регистрация middleware
	// TODO: gRPC-Gateway роутинг

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("🛑 API Gateway shutting down...")
}
