package main

import (
	"log"
	"os"
	"strconv"

	"distributed-calculator/internal/agent"
)

func main() {
	os.Setenv("COMPUTING_POWER", "3")
	power, _ := strconv.Atoi(os.Getenv("COMPUTING_POWER"))

	for i := 0; i < power; i++ {
		go agent.RunWorker()
	}

	log.Println("Agent started with", power, "workers")
	select {}
}
