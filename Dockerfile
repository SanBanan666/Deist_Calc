FROM golang:1.21-alpine

# Установка зависимостей
RUN apk add --no-cache protobuf-dev

# Установка Go gRPC плагина
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Копирование файлов проекта
WORKDIR /app
COPY . .

# Генерация gRPC кода
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    internal/proto/calculator.proto

# Сборка приложения
RUN go build -o server cmd/server/main.go
RUN go build -o agent cmd/agent/main.go

# Запуск приложения
CMD ["./scripts/docker-entrypoint.sh"] 