# build.ps1 - Скрипт сборки проекта Venera
#
# Этот скрипт собирает приложение Venera для Windows x64.
# Поддерживает отладочную и релизную сборку.
#
# Использование:
#   .\build.ps1 [-Release] [-Clean] [-Verbose]
#
# Параметры:
#   -Release: Сборка в релизном режиме
#   -Clean: Очистка предыдущих сборок
#   -Verbose: Подробный вывод
#
# Примеры:
#   .\build.ps1
#   .\build.ps1 -Release
#   .\build.ps1 -Clean -Release

param (
    [switch]$Release = $false,
    [switch]$Clean = $false,
    [switch]$Verbose = $false
)

# Включение строгого режима
Set-StrictMode -Version Latest

# Настройки
$script:ProjectName = "venera"
$script:OutputDir = "bin"
$script:BuildDir = "build"
$script:GoVersion = "1.21"
$script:TargetOS = "windows"
$script:TargetArch = "amd64"

# Функция для вывода сообщений
function Write-Message {
    param (
        [Parameter(Mandatory=$true)]
        [string]$Message,
        
        [ValidateSet('Info', 'Success', 'Warning', 'Error')]
        [string]$Level = 'Info'
    )
    
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $color = switch ($Level) {
        'Success' { 'Green' }
        'Warning' { 'Yellow' }
        'Error' { 'Red' }
        default { 'White' }
    }
    
    Write-Host "[$timestamp] [$Level] $Message" -ForegroundColor $color
}

# Функция для проверки наличия утилит
function Test-Command {
    param (
        [Parameter(Mandatory=$true)]
        [string]$Command
    )
    
    $exists = Get-Command -Name $Command -ErrorAction SilentlyContinue
    return $exists -ne $null
}

# Функция для проверки версии Go
function Test-GoVersion {
    if (-not (Test-Command "go")) {
        Write-Message "Go не установлен" -Level Error
        return $false
    }
    
    $version = go version 2>&1
    if ($version -match "go version go(\d+\.\d+)") {
        $goVersion = $matches[1]
        Write-Message "Обнаружена версия Go: $goVersion" -Level Info
        
        if ([version]$goVersion -lt [version]$script:GoVersion) {
            Write-Message "Требуется Go $script:GoVersion или выше" -Level Warning
            return $false
        }
        
        return $true
    }
    
    return $false
}

# Функция для очистки
function Clean-Build {
    Write-Message "Очистка предыдущих сборок..." -Level Info
    
    # Удаление директорий
    if (Test-Path $script:OutputDir) {
        Remove-Item -Path $script:OutputDir -Recurse -Force
        Write-Message "Удалена директория $script:OutputDir" -Level Info
    }
    
    if (Test-Path $script:BuildDir) {
        Remove-Item -Path $script:BuildDir -Recurse -Force
        Write-Message "Удалена директория $script:BuildDir" -Level Info
    }
    
    # Очистка Go кэша
    go clean -cache
    Write-Message "Очищен кэш Go" -Level Info
}

# Функция для сборки
function Build-Project {
    param (
        [switch]$Release
    )
    
    Write-Message "Начало сборки..." -Level Info
    
    # Создание директорий
    if (-not (Test-Path $script:OutputDir)) {
        New-Item -Path $script:OutputDir -ItemType Directory | Out-Null
    }
    
    if (-not (Test-Path $script:BuildDir)) {
        New-Item -Path $script:BuildDir -ItemType Directory | Out-Null
    }
    
    # Настройки сборки
    $buildFlags = @()
    
    if ($Release) {
        $buildFlags += "-release"
        Write-Message "Режим: Release" -Level Info
    } else {
        Write-Message "Режим: Debug" -Level Info
    }
    
    # Имена выходных файлов
    $exePath = Join-Path $script:OutputDir "$script:ProjectName.exe"
    
    # Сборка
    Write-Message "Компиляция..." -Level Info
    
    $env:GOOS = $script:TargetOS
    $env:GOARCH = $script:TargetArch
    
    $buildArgs = @("build")
    
    if ($Release) {
        $buildArgs += "-ldflags=-s -w"
    }
    
    $buildArgs += "-o"
    $buildArgs += $exePath
    $buildArgs += "."
    
    Write-Message "Команда: go $([string]::Join(' ', $buildArgs))" -Level Info
    
    $startTime = Get-Date
    $result = go @buildArgs 2>&1
    
    if ($LASTEXITCODE -ne 0) {
        Write-Message "Ошибка сборки: $result" -Level Error
        return $false
    }
    
    $endTime = Get-Date
    $duration = ($endTime - $startTime).TotalSeconds
    
    # Проверка выходного файла
    if (Test-Path $exePath) {
        $fileSize = (Get-Item $exePath).Length
        Write-Message "Сборка завершена успешно!" -Level Success
        Write-Message "Выходной файл: $exePath" -Level Info
        Write-Message "Размер: $([math]::Round($fileSize / 1KB, 2)) KB" -Level Info
        Write-Message "Время сборки: $([math]::Round($duration, 2)) сек" -Level Info
        
        # Копирование конфигурационных файлов
        Write-Message "Копирование конфигурационных файлов..." -Level Info
        Copy-Item -Path "config.toml" -Destination $script:OutputDir -Force
        Copy-Item -Path "processes.toml" -Destination $script:OutputDir -Force
        
        # Копирование settings
        if (Test-Path "settings") {
            Copy-Item -Path "settings\*" -Destination (Join-Path $script:OutputDir "settings") -Recurse -Force
        }
        
        return $true
    } else {
        Write-Message "Файл не создан" -Level Error
        return $false
    }
}

# Функция для запуска тестов
function Run-Tests {
    Write-Message "Запуск тестов..." -Level Info
    
    $result = go test ./... 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Message "Тесты пройдены успешно!" -Level Success
        return $true
    } else {
        Write-Message "Тесты завершились с ошибками" -Level Error
        return $false
    }
}

# Функция для создания архива
function Create-Archive {
    param (
        [string]$Version
    )
    
    Write-Message "Создание архива..." -Level Info
    
    $archiveName = "$script:ProjectName-$Version-windows-amd64.zip"
    $archivePath = Join-Path $script:OutputDir $archiveName
    
    # Создание архива
    Compress-Archive -Path (Join-Path $script:OutputDir "*") -DestinationPath $archivePath -Force
    
    if (Test-Path $archivePath) {
        $fileSize = (Get-Item $archivePath).Length
        Write-Message "Архив создан: $archivePath" -Level Success
        Write-Message "Размер архива: $([math]::Round($fileSize / 1KB, 2)) KB" -Level Info
        return $true
    } else {
        Write-Message "Ошибка создания архива" -Level Error
        return $false
    }
}

# Основная функция
function Main {
    Write-Message "========================================" -Level Info
    Write-Message "Venera Build Script" -Level Info
    Write-Message "========================================" -Level Info
    
    # Проверка Go
    if (-not (Test-GoVersion)) {
        Write-Message "Установите Go $script:GoVersion или выше" -Level Error
        exit 1
    }
    
    # Очистка
    if ($Clean) {
        Clean-Build
    }
    
    # Сборка
    $buildSuccess = Build-Project -Release:$Release
    
    if (-not $buildSuccess) {
        Write-Message "Сборка завершилась с ошибками" -Level Error
        exit 1
    }
    
    # Запуск тестов
    Write-Message "Запуск тестов..." -Level Info
    Run-Tests
    
    # Создание архива
    $version = "1.0.0"  # Получите версию из кода или файла
    Create-Archive -Version $version
    
    Write-Message "========================================" -Level Info
    Write-Message "Сборка завершена успешно!" -Level Success
    Write-Message "========================================" -Level Info
    
    Write-Message "Готовый файл: $exePath" -Level Success
    Write-Message "Архив: $archivePath" -Level Success
}

# Запуск основной функции
Main
