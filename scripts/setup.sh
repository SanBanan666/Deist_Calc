#!/bin/bash

# Установка protoc
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux
    sudo apt-get update
    sudo apt-get install -y protobuf-compiler
elif [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    brew install protobuf
elif [[ "$OSTYPE" == "msys" ]]; then
    # Windows
    echo "Для Windows установите protoc вручную: https://grpc.io/docs/protoc-installation/"
    exit 1
fi

# Установка Go плагинов для protoc
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Добавление GOPATH/bin в PATH
export PATH="$PATH:$(go env GOPATH)/bin"

# Генерация gRPC кода
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    internal/proto/calculator.proto

echo "Установка завершена!" 