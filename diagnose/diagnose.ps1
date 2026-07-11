# diagnose.ps1 - Консольная диагностика для Venera
# 
# Этот скрипт выполняет консольную диагностику системы сбора идентификаторов,
# включая проверку конфигурации, ресурсов, зависимостей и создание отчетов.
#
# Основные функции:
# - Проверка конфигурации
# - Проверка регистрации
# - Проверка ресурсов
# - Проверка сетевых интерфейсов
# - Проверка зависимостей (podman, tshark)
# - Проверка служб
# - История событий Event Log
# - Экспорт отчетов в файл
# - Создание архивов
#
# Использование:
#   .\diagnose.ps1 [-OutputPath "путь\к\отчету"] [-ArchivePath "путь\к\архиву"]
#   .\diagnose.ps1 -Verbose
#
# Примеры:
#   .\diagnose.ps1
#   .\diagnose.ps1 -OutputPath ".\diagnose-report.txt"
#   .\diagnose.ps1 -OutputPath ".\diagnose-report.txt" -ArchivePath ".\diagnose-archive.zip"
#   .\diagnose.ps1 -Verbose

[CmdletBinding()]
param(
    [string]$OutputPath = ".\diagnose-report.txt",
    [string]$ArchivePath = ".\diagnose-archive.zip"
)

Write-Host "=== Diagnose.ps1 - Консольная диагностика Venera ===" -ForegroundColor Cyan
Write-Host "Версия скрипта: 1.0.0" -ForegroundColor Cyan
Write-Host "Время начала: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')" -ForegroundColor Cyan
Write-Host ""

# Инициализация результатов
$Results = @()

# Функция добавления результата
function Add-Result {
    param(
        [string]$Name,
        [string]$Status,
        [string]$Message
    )
    
    $Result = [PSCustomObject]@{
        Name = $Name
        Status = $Status
        Message = $Message
        Timestamp = Get-Date -Format 'yyyy-MM-dd HH:mm:ss'
    }
    
    $script:Results += $Result
    
    $Color = switch ($Status) {
        "success" { "Green" }
        "warning" { "Yellow" }
        "error" { "Red" }
        default { "White" }
    }
    
    Write-Host "[$Status] $Name: $Message" -ForegroundColor $Color
}

# Проверка конфигурации
Write-Host "" -ForegroundColor Yellow
Write-Host "=== Проверка конфигурации ===" -ForegroundColor Yellow

if (Test-Path "config.toml") {
    Add-Result -Name "Файл конфигурации" -Status "success" -Message "config.toml найден"
    
    try {
        $Config = Get-Content "config.toml" -Raw
        Add-Result -Name "Чтение конфигурации" -Status "success" -Message "Файл прочитан успешно"
    } catch {
        Add-Result -Name "Чтение конфигурации" -Status "error" -Message "Ошибка чтения: $_"
    }
} else {
    Add-Result -Name "Файл конфигурации" -Status "error" -Message "config.toml не найден"
}

# Проверка регистрации
Write-Host "" -ForegroundColor Yellow
Write-Host "=== Проверка регистрации ===" -ForegroundColor Yellow

try {
    $RegKey = "HKLM:\SYSTEM\CurrentControlSet\Services\EventLog\Application\Venera"
    if (Test-Path $RegKey) {
        $RegValues = Get-ItemProperty $RegKey -ErrorAction Stop
        Add-Result -Name "Регистрация Event Log" -Status "success" -Message "Регистрация найдена"
    } else {
        Add-Result -Name "Регистрация Event Log" -Status "warning" -Message "Регистрация не найдена (необязательно)"
    }
} catch {
    Add-Result -Name "Регистрация Event Log" -Status "warning" -Message "Ошибка проверки: $_"
}

# Проверка ресурсов
Write-Host "" -ForegroundColor Yellow
Write-Host "=== Проверка ресурсов ===" -ForegroundColor Yellow

# Проверка RAM
try {
    $MemInfo = Get-CimInstance -ClassName Win32_OperatingSystem
    $TotalRAM = [math]::Round($MemInfo.TotalVisibleMemorySize / 1MB, 2)
    $FreeRAM = [math]::Round($MemInfo.FreePhysicalMemory / 1MB, 2)
    $UsedRAM = [math]::Round(($MemInfo.TotalVisibleMemorySize - $MemInfo.FreePhysicalMemory) / 1MB, 2)
    $UsagePercent = [math]::Round(($UsedRAM / $TotalRAM) * 100, 2)
    
    Add-Result -Name "RAM" -Status "success" -Message "$UsedRAM GB / $TotalRAM GB ($UsagePercent%)"
    
    if ($UsagePercent -gt 90) {
        Add-Result -Name "RAM warning" -Status "warning" -Message "Использование RAM > 90%"
    }
} catch {
    Add-Result -Name "RAM" -Status "error" -Message "Ошибка получения информации: $_"
}

# Проверка диска
try {
    $Disk = Get-CimInstance -ClassName Win32_LogicalDisk -Filter "DeviceID='C:'"
    $TotalDisk = [math]::Round($Disk.Size / 1GB, 2)
    $FreeDisk = [math]::Round($Disk.FreeSpace / 1GB, 2)
    $UsedDisk = [math]::Round(($Disk.Size - $Disk.FreeSpace) / 1GB, 2)
    $DiskUsagePercent = [math]::Round(($UsedDisk / $TotalDisk) * 100, 2)
    
    Add-Result -Name "Disk C:" -Status "success" -Message "$UsedDisk GB / $TotalDisk GB ($DiskUsagePercent%)"
    
    if ($DiskUsagePercent -gt 90) {
        Add-Result -Name "Disk C: warning" -Status "warning" -Message "Использование диска > 90%"
    }
} catch {
    Add-Result -Name "Disk C:" -Status "error" -Message "Ошибка получения информации: $_"
}

# Проверка сетевых интерфейсов
Write-Host "" -ForegroundColor Yellow
Write-Host "=== Проверка сетевых интерфейсов ===" -ForegroundColor Yellow

try {
    $Interfaces = Get-NetAdapter | Where-Object { $_.Status -eq "Up" }
    $InterfaceCount = $Interfaces.Count
    
    Add-Result -Name "Активные сетевые интерфейсы" -Status "success" -Message "Найдено $InterfaceCount интерфейса(ов)"
    
    foreach ($Interface in $Interfaces) {
        Add-Result -Name "Интерфейс $($Interface.Name)" -Status "success" -Message "$($Interface.InterfaceDescription) - $($Interface.LinkSpeed)"
    }
} catch {
    Add-Result -Name "Сетевые интерфейсы" -Status "error" -Message "Ошибка получения информации: $_"
}

# Проверка tshark
Write-Host "" -ForegroundColor Yellow
Write-Host "=== Проверка tshark ===" -ForegroundColor Yellow

$TsharkPath = "C:\Program Files\Wireshark\tshark.exe"
if (Test-Path $TsharkPath) {
    try {
        $Version = & $TsharkPath -v 2>&1 | Select-String "Tshark"
        Add-Result -Name "tshark" -Status "success" -Message "Найден: $Version"
    } catch {
        Add-Result -Name "tshark" -Status "error" -Message "Ошибка проверки версии: $_"
    }
} else {
    Add-Result -Name "tshark" -Status "warning" -Message "Не найден: $TsharkPath"
}

# Проверка podman
Write-Host "" -ForegroundColor Yellow
Write-Host "=== Проверка podman ===" -ForegroundColor Yellow

try {
    $PodmanVersion = podman --version 2>&1
    if ($LASTEXITCODE -eq 0) {
        Add-Result -Name "podman" -Status "success" -Message "$PodmanVersion"
    } else {
        Add-Result -Name "podman" -Status "warning" -Message "Команда завершилась с ошибкой"
    }
} catch {
    Add-Result -Name "podman" -Status "warning" -Message "Не найден в PATH"
}

# Проверка образа DragonflyDB
Write-Host "" -ForegroundColor Yellow
Write-Host "=== Проверка образа DragonflyDB ===" -ForegroundColor Yellow

try {
    $Images = podman images | Select-String "dragonflydb"
    if ($Images) {
        Add-Result -Name "Образ DragonflyDB" -Status "success" -Message "Найден: $Images"
    } else {
        Add-Result -Name "Образ DragonflyDB" -Status "warning" -Message "Не найден образ dragonflydb"
    }
} catch {
    Add-Result -Name "Образ DragonflyDB" -Status "warning" -Message "Ошибка проверки образа: $_"
}

# Проверка контейнера DragonflyDB
Write-Host "" -ForegroundColor Yellow
Write-Host "=== Проверка контейнера DragonflyDB ===" -ForegroundColor Yellow

try {
    $Containers = podman ps -a | Select-String "cachedb"
    if ($Containers) {
        Add-Result -Name "Контейнер cachedb" -Status "success" -Message "Найден: $Containers"
    } else {
        Add-Result -Name "Контейнер cachedb" -Status "warning" -Message "Не найден контейнер cachedb"
    }
} catch {
    Add-Result -Name "Контейнер cachedb" -Status "warning" -Message "Ошибка проверки контейнера: $_"
}

# Проверка службы
Write-Host "" -ForegroundColor Yellow
Write-Host "=== Проверка службы ===" -ForegroundColor Yellow

try {
    $Service = Get-Service -Name "VeneraSrv" -ErrorAction Stop
    Add-Result -Name "Служба VeneraSrv" -Status "success" -Message "Статус: $($Service.Status)"
    
    if ($Service.Status -ne "Running") {
        Add-Result -Name "Служба VeneraSrv warning" -Status "warning" -Message "Служба не запущена"
    }
} catch {
    Add-Result -Name "Служба VeneraSrv" -Status "warning" -Message "Служба не найдена (необязательно)"
}

# Проверка Event Log
Write-Host "" -ForegroundColor Yellow
Write-Host "=== Проверка Event Log ===" -ForegroundColor Yellow

try {
    $Events = Get-EventLog -LogName Application -EntryType Error -Newest 10 -ErrorAction Stop
    $EventCount = $Events.Count
    
    Add-Result -Name "Event Log" -Status "success" -Message "Найдено $EventCount ошибок за последнее время"
    
    foreach ($Event in $Events) {
        Add-Result -Name "Event Log [$($Event.EntryType)]" -Status "warning" -Message "$($Event.Source): $($Event.Message)"
    }
} catch {
    Add-Result -Name "Event Log" -Status "error" -Message "Ошибка получения событий: $_"
}

# Экспорт отчета
Write-Host "" -ForegroundColor Yellow
Write-Host "=== Экспорт отчета ===" -ForegroundColor Yellow

try {
    $Report = @()
    $Report += "=== Отчет диагностики Venera ==="
    $Report += "Время создания: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
    $Report += "Версия ОС: $($env:OS)"
    $Report += "Версия скрипта: 1.0.0"
    $Report += ""
    $Report += "=== Результаты диагностики ==="
    
    foreach ($Result in $Results) {
        $Report += "[$($Result.Status)] $($Result.Name): $($Result.Message) (Время: $($Result.Timestamp))"
    }
    
    $Report | Out-File -FilePath $OutputPath -Encoding UTF8
    Add-Result -Name "Экспорт отчета" -Status "success" -Message "Отчет сохранен в $OutputPath"
} catch {
    Add-Result -Name "Экспорт отчета" -Status "error" -Message "Ошибка сохранения отчета: $_"
}

# Создание архива
Write-Host "" -ForegroundColor Yellow
Write-Host "=== Создание архива ===" -ForegroundColor Yellow

try {
    $ArchiveItems = @()
    
    if (Test-Path "Logs") {
        $ArchiveItems += "Logs\*"
    }
    
    if (Test-Path "config.toml") {
        $ArchiveItems += "config.toml"
    }
    
    if ($ArchiveItems.Count -gt 0) {
        Compress-Archive -Path $ArchiveItems -DestinationPath $ArchivePath -Force
        Add-Result -Name "Архив" -Status "success" -Message "Архив создан: $ArchivePath"
    } else {
        Add-Result -Name "Архив" -Status "warning" -Message "Нет файлов для архивации"
    }
} catch {
    Add-Result -Name "Архив" -Status "error" -Message "Ошибка создания архива: $_"
}

# Финальный отчет
Write-Host "" -ForegroundColor Cyan
Write-Host "=== Финальный отчет ===" -ForegroundColor Cyan

$SuccessCount = ($Results | Where-Object { $_.Status -eq "success" }).Count
$WarningCount = ($Results | Where-Object { $_.Status -eq "warning" }).Count
$ErrorCount = ($Results | Where-Object { $_.Status -eq "error" }).Count

Write-Host "Всего проверок: $($Results.Count)" -ForegroundColor White
Write-Host "  Успешно: $SuccessCount" -ForegroundColor Green
Write-Host "  Предупреждений: $WarningCount" -ForegroundColor Yellow
Write-Host "  Ошибок: $ErrorCount" -ForegroundColor Red
Write-Host ""
Write-Host "Время окончания: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')" -ForegroundColor Cyan
Write-Host "Отчет сохранен в: $OutputPath" -ForegroundColor Cyan
Write-Host "Архив сохранен в: $ArchivePath" -ForegroundColor Cyan
Write-Host ""

# Выход с кодом в зависимости от результатов
if ($ErrorCount -gt 0) {
    exit 1
} elseif ($WarningCount -gt 0) {
    exit 2
} else {
    exit 0
}