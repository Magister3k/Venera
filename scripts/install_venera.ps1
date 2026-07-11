# install_venera.ps1 - Скрипт установки сервиса Venera
#
# Этот скрипт устанавливает и запускает сервис Venera в Windows.
# Поддерживает установку как обычной программы или как Windows-сервиса.
#
# Использование:
#   .\\install_venera.ps1 [-Service] [-Verbose]
#
# Параметры:
#   -Service: Установить как Windows-сервис
#   -Verbose: Подробный вывод
#
# Примеры:
#   .\\install_venera.ps1
#   .\\install_venera.ps1 -Service

param (
    [switch]$Service = $false,
    [switch]$Verbose = $false
)

Set-StrictMode -Version Latest

$script:ProjectName = "venera"
$script:ServiceName = "VeneraSrv"
$script:InstallPath = "C:\\Program Files\\Venera"
$script:BinaryPath = Join-Path $script:InstallPath "$script:ProjectName.exe"
$script:LogPath = "C:\\ProgramData\\Venera\\Logs"

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

# Функция для проверки прав администратора
function Test-Administrator {
    $currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
    return $currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# Функция для создания директорий
function Create-Directories {
    Write-Message "Создание директорий..." -Level Info
    
    if (-not (Test-Path $script:InstallPath)) {
        New-Item -Path $script:InstallPath -ItemType Directory | Out-Null
        Write-Message "Создана директория: $script:InstallPath" -Level Info
    }
    
    if (-not (Test-Path $script:LogPath)) {
        New-Item -Path $script:LogPath -ItemType Directory | Out-Null
        Write-Message "Создана директория: $script:LogPath" -Level Info
    }
}

# Функция для копирования файлов
function Copy-Files {
    param (
        [string]$SourcePath
    )
    
    Write-Message "Копирование файлов..." -Level Info
    
    # Копирование бинарного файла
    $binarySource = Join-Path $SourcePath "$script:ProjectName.exe"
    if (Test-Path $binarySource) {
        Copy-Item -Path $binarySource -Destination $script:BinaryPath -Force
        Write-Message "Скопирован бинарный файл" -Level Info
    } else {
        Write-Message "Бинарный файл не найден: $binarySource" -Level Error
        return $false
    }
    
    # Копирование конфигурационных файлов
    $configSource = Join-Path $SourcePath "config.toml"
    if (Test-Path $configSource) {
        Copy-Item -Path $configSource -Destination (Join-Path $script:InstallPath "config.toml") -Force
        Write-Message "Скопирован config.toml" -Level Info
    }
    
    $processesSource = Join-Path $SourcePath "processes.toml"
    if (Test-Path $processesSource) {
        Copy-Item -Path $processesSource -Destination (Join-Path $script:InstallPath "processes.toml") -Force
        Write-Message "Скопирован processes.toml" -Level Info
    }
    
    # Копирование settings
    $settingsSource = Join-Path $SourcePath "settings"
    if (Test-Path $settingsSource) {
        Copy-Item -Path $settingsSource -Destination (Join-Path $script:InstallPath "settings") -Recurse -Force
        Write-Message "Скопированы настройки" -Level Info
    }
    
    return $true
}

# Функция для установки как сервиса
function Install-Service {
    Write-Message "Установка Windows-сервиса..." -Level Info
    
    # Проверка наличия сервиса
    $service = Get-Service -Name $script:ServiceName -ErrorAction SilentlyContinue
    
    if ($service -ne $null) {
        Write-Message "Сервис уже установлен: $script:ServiceName" -Level Warning
        $response = Read-Host "Удалить существующий сервис? (y/n)"
        if ($response -eq 'y') {
            Stop-Service -Name $script:ServiceName -Force
            sc.exe delete $script:ServiceName
            Start-Sleep -Seconds 2
        } else {
            return $true
        }
    }
    
    # Установка сервиса
    $installCommand = "sc.exe create $script:ServiceName binPath=`"$script:BinaryPath --service`" start=auto DisplayName=`"Venera Service`""
    Write-Message "Команда: $installCommand" -Level Info
    
    $result = Invoke-Expression $installCommand
    
    if ($LASTEXITCODE -eq 0) {
        Write-Message "Сервис успешно установлен" -Level Success
        return $true
    } else {
        Write-Message "Ошибка установки сервиса: $result" -Level Error
        return $false
    }
}

# Функция для запуска сервиса
function Start-Service {
    Write-Message "Запуск сервиса..." -Level Info
    
    try {
        Start-Service -Name $script:ServiceName
        Start-Sleep -Seconds 1
        
        $service = Get-Service -Name $script:ServiceName
        if ($service.Status -eq 'Running') {
            Write-Message "Сервис запущен успешно" -Level Success
            return $true
        } else {
            Write-Message "Не удалось запустить сервис" -Level Error
            return $false
        }
    } catch {
        Write-Message "Ошибка запуска сервиса: $_" -Level Error
        return $false
    }
}

# Функция для запуска в режиме приложения
function Start-Application {
    Write-Message "Запуск приложения..." -Level Info
    
    $process = Start-Process -FilePath $script:BinaryPath -PassThru -WindowStyle Normal
    
    if ($process -ne $null) {
        Write-Message "Приложение запущено (PID: $($process.Id))" -Level Success
        return $true
    } else {
        Write-Message "Не удалось запустить приложение" -Level Error
        return $false
    }
}

# Основная функция
function Main {
    Write-Message "========================================" -Level Info
    Write-Message "Venera Installation Script" -Level Info
    Write-Message "========================================" -Level Info
    
    # Проверка прав
    if (-not (Test-Administrator)) {
        Write-Message "Запустите скрипт от имени администратора" -Level Error
        exit 1
    }
    
    # Создание директорий
    Create-Directories
    
    # Копирование файлов
    $sourcePath = if (Test-Path ".\\bin") { ".\\bin" } elseif (Test-Path "..\\bin") { "..\\bin" } else { $PSScriptRoot }
    
    if (-not (Copy-Files -SourcePath $sourcePath)) {
        exit 1
    }
    
    # Установка сервиса или запуск приложения
    if ($Service) {
        if (-not (Install-Service)) {
            exit 1
        }
        
        Start-Service
    } else {
        Start-Application
    }
    
    Write-Message "========================================" -Level Info
    Write-Message "Установка завершена успешно!" -Level Success
    Write-Message "========================================" -Level Info
}

Main