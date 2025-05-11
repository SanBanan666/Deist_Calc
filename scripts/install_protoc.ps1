$protocVersion = "25.3"
$protocUrl = "https://github.com/protocolbuffers/protobuf/releases/download/v$protocVersion/protoc-$protocVersion-win64.zip"
$protocZip = "protoc.zip"
$protocDir = "C:\protoc"

# Создаем директорию для protoc
New-Item -ItemType Directory -Force -Path $protocDir

# Скачиваем protoc
Invoke-WebRequest -Uri $protocUrl -OutFile $protocZip

# Распаковываем архив
Expand-Archive -Path $protocZip -DestinationPath $protocDir -Force

# Добавляем путь в PATH
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (-not $currentPath.Contains($protocDir)) {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$protocDir\bin", "User")
}

# Удаляем временный файл
Remove-Item $protocZip

Write-Host "Protoc успешно установлен в $protocDir"
Write-Host "Пожалуйста, перезапустите PowerShell для применения изменений PATH" 