#!/bin/sh

# Запускаем сервер
./server &
SERVER_PID=$!

# Ждем 2 секунды
sleep 2

# Запускаем агент
./agent &
AGENT_PID=$!

echo "Приложение запущено в Docker контейнере!"
echo "HTTP сервер: http://localhost:8080"
echo "gRPC сервер: localhost:50051"

# Обработка сигналов
trap "kill $SERVER_PID $AGENT_PID; exit" INT TERM

# Ждем завершения
wait 