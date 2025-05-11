package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"Deist_Calc/internal/client"
	"Deist_Calc/internal/server"
	"Deist_Calc/internal/storage"
)

func main() {
	// Парсим флаги командной строки
	httpPort := flag.Int("http-port", 8080, "HTTP порт для веб-интерфейса")
	grpcPort := flag.Int("grpc-port", 50051, "gRPC порт для сервиса калькулятора")
	dbPath := flag.String("db", "calculator.db", "путь к файлу базы данных SQLite")
	webDir := flag.String("web-dir", "web", "директория с веб-файлами")
	flag.Parse()

	// Инициализируем хранилище
	store, err := storage.NewStorage(*dbPath)
	if err != nil {
		log.Fatalf("Ошибка инициализации хранилища: %v", err)
	}

	// Создаем gRPC клиент
	grpcClient, err := client.NewGRPCClient(fmt.Sprintf("localhost:%d", *grpcPort))
	if err != nil {
		log.Fatalf("Ошибка создания gRPC клиента: %v", err)
	}

	// Создаем и запускаем HTTP сервер
	httpServer := server.NewHTTPServer(store, grpcClient)
	go func() {
		if err := httpServer.Start(*httpPort, *webDir); err != nil {
			log.Fatalf("Ошибка запуска HTTP сервера: %v", err)
		}
	}()

	// Запускаем gRPC сервер
	if err := server.StartGRPCServer(store); err != nil {
		log.Fatalf("Ошибка запуска gRPC сервера: %v", err)
	}

	// Ждем сигнала завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Завершение работы сервера...")
}
