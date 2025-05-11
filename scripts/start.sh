#!/bin/bash

# Останавливаем все процессы Go
pkill -f "go run" 2>/dev/null

# Запускаем сервер
go run cmd/server/main.go &
SERVER_PID=$!

# Ждем 2 секунды
sleep 2

# Запускаем агент
go run cmd/agent/main.go &
AGENT_PID=$!

echo "Приложение запущено!"
echo "HTTP сервер: http://localhost:8080"
echo "gRPC сервер: localhost:50051"

# Обработка Ctrl+C
trap "kill $SERVER_PID $AGENT_PID; exit" INT

# Ждем завершения
wait 