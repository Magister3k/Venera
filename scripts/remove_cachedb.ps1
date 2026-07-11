# remove_cachedb.ps1 - Скрипт удаления кэш-базы данных DragonflyDB
#
# Этот скрипт удаляет все данные из кэш-базы DragonflyDB.
# Предназначен для очистки кэша при необходимости.
#
# Использование:
#   .\remove_cachedb.ps1 [-Verbose]
#
# Параметры:
#   -Verbose: Подробный вывод
#
# Примеры:
#   .\remove_cachedb.ps1

param (
    [switch]$Verbose = $false
)

Set-StrictMode -Version Latest

$script:DragonflyHost = "localhost"
$script:DragonflyPort = 6379
$script:DragonflyPassword = $null  # Установите пароль при необходимости

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

# Функция для проверки подключения к DragonflyDB
function Test-DragonflyConnection {
    param (
        [string]$Host = $script:DragonflyHost,
        [int]$Port = $script:DragonflyPort
    )
    
    try {
        $socket = New-Object System.Net.Sockets.TcpClient($Host, $Port)
        if ($socket -ne $null) {
            $socket.Close()
            return $true
        }
    } catch {
        return $false
    }
    
    return $false
}

# Функция для очистки базы данных
function Clear-DragonflyDatabase {
    param (
        [string]$Host = $script:DragonflyHost,
        [int]$Port = $script:DragonflyPort,
        [string]$Password = $script:DragonflyPassword
    )
    
    Write-Message "Подключение к DragonflyDB ($Host:$Port)...`n" -Level Info
    
    # Проверка подключения
    if (-not (Test-DragonflyConnection -Host $Host -Port $Port)) {
        Write-Message "Не удалось подключиться к DragonflyDB" -Level Error
        return $false
    }
    
    Write-Message "Подключение успешно" -Level Success
    
    # Подключение через redis-cli (если установлен)
    if (Get-Command "redis-cli" -ErrorAction SilentlyContinue) {
        Write-Message "Очистка базы данных..." -Level Info
        
        $args = @("-h", $Host, "-p", $Port)
        if ($Password) {
            $args += @("-a", $Password)
        }
        $args += "FLUSHALL"
        
        $result = redis-cli @args 2>&1
        
        if ($LASTEXITCODE -eq 0) {
            Write-Message "Файл data.json успешно создан" -Level Success
            return $true
        } else {
            Write-Message "Ошибка создания data.json: $result" -Level Error
            return $false
        }
    } else {
        Write-Message "redis-cli не найден. Установите Redis для Windows или используйте другую утилиту для подключения к DragonflyDB." -Level Warning
        Write-Message "Выполните вручную: redis-cli -h $Host -p $Port FLUSHALL" -Level Info
        return $false
    }
}

# Основная функция
function Main {
    Write-Message "========================================" -Level Info
    Write-Message "Venera Cache Database Removal Script" -Level Info
    Write-Message "========================================" -Level Info
    
    # Очистка базы данных
    if (-not (Clear-DragonflyDatabase)) {
        Write-Message "Удаление кэш-базы данных завершено с ошибками" -Level Error
        exit 1
    }
    
    Write-Message "========================================" -Level Info
    Write-Message "Кэш-база данных успешно удалена!" -Level Success
    Write-Message "========================================" -Level Info
}

Main