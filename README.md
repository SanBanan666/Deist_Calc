# Калькулятор выражений

Веб-приложение для вычисления математических выражений с использованием микросервисной архитектуры.

## Требования

- Go 1.21 или выше
- SQLite3
- Protoc (Protocol Buffers Compiler)
- Go gRPC плагин

## Установка зависимостей

### 1. Установка Protocol Buffers Compiler (protoc)

#### Windows:
```powershell
# Скачайте и установите protoc из https://github.com/protocolbuffers/protobuf/releases
# Или используйте скрипт установки:
.\scripts\install_protoc.ps1
```

#### Linux:
```bash
sudo apt-get update
sudo apt-get install -y protobuf-compiler
```

### 2. Установка Go gRPC плагина
```bash
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 3. Установка зависимостей проекта
```bash
go mod download
```

## Генерация gRPC кода

```bash
# Windows
C:\protoc\bin\protoc.exe --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative internal/proto/calculator.proto

# Linux/Mac
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative internal/proto/calculator.proto
```

## Запуск приложения

### Способ 1: Запуск через отдельные терминалы

1. Запустите сервер:
```bash
go run cmd/server/main.go
```

2. В другом терминале запустите агент:
```bash
go run cmd/agent/main.go
```

### Способ 2: Запуск через скрипт

#### Windows:
```powershell
.\scripts\start.ps1
```

#### Linux/Mac:
```bash
./scripts/start.sh
```

### Способ 3: Запуск через Docker

1. Соберите Docker образ:
```bash
docker build -t calculator .
```

2. Запустите контейнер:
```bash
docker run -p 8080:8080 -p 50051:50051 calculator
```

## Использование

1. Откройте веб-интерфейс в браузере:
```
http://localhost:8080
```

2. Зарегистрируйтесь или войдите в систему

3. Введите математическое выражение (например: "2 + 2 * 2")

4. Нажмите "Вычислить" для получения результата

## Структура проекта

```
.
├── cmd/
│   ├── server/     # HTTP и gRPC сервер
│   └── agent/      # Вычислительный агент
├── internal/
│   ├── proto/      # gRPC протоколы
│   ├── storage/    # Работа с базой данных
│   └── agent/      # Логика вычислений
├── web/           # Веб-интерфейс
└── scripts/       # Скрипты для установки и запуска
```

## Решение проблем

### Порт 50051 занят
```bash
# Windows
netstat -ano | findstr :50051
taskkill /F /PID <PID>

# Linux/Mac
lsof -i :50051
kill -9 <PID>
```

### Ошибка "нет доступных задач"
1. Проверьте, что сервер и агент запущены
2. Проверьте логи на наличие ошибок
3. Перезапустите сервер и агент

### Ошибка компиляции protoc
1. Убедитесь, что protoc установлен и доступен в PATH
2. Проверьте установку Go gRPC плагина
3. Перегенерируйте gRPC код

## Логирование

- HTTP сервер: порт 8080
- gRPC сервер: порт 50051
- База данных: SQLite (calculator.db)

## Лицензия

MIT