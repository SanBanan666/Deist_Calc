package main

import (
	"log"

	"Deist_Calc/internal/agent"
)

func main() {
	worker, err := agent.NewWorker("localhost:50051")
	if err != nil {
		log.Fatalf("Ошибка создания воркера: %v", err)
	}

	log.Println("Агент запущен и готов к работе")
	worker.Start()
}
