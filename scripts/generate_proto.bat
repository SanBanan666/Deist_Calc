@echo off
set PATH=%PATH%;C:\protoc
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative internal/proto/calculator.proto 