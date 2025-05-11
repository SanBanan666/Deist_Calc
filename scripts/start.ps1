# Останавливаем все процессы Go
taskkill /F /IM go.exe 2>$null

# Запускаем сервер
Start-Process powershell -ArgumentList "go run cmd/server/main.go"

# Ждем 2 секунды
Start-Sleep -Seconds 2

# Запускаем агент
Start-Process powershell -ArgumentList "go run cmd/agent/main.go"

Write-Host "Приложение запущено!"
Write-Host "HTTP сервер: http://localhost:8080"
Write-Host "gRPC сервер: localhost:50051" 