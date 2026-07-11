# Venera - Система сбора идентификаторов в потоке пакетных данных

[![Windows](https://img.shields.io/badge/OS-Windows-blue.svg)](https://www.microsoft.com/)
[![Go](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

Venera - это высокопроизводительная система для сбора, фильтрации и хранения идентификаторов из потоков пакетных данных. Приложение поддерживает несколько режимов работы, веб-интерфейс и комплексную диагностику.

## Содержание

- [Особенности](#особенности)
- [Архитектура](#архитектура)
- [Требования](#требования)
- [Установка](#установка)
- [Конфигурация](#конфигурация)
- [Использование](#использование)
- [Веб-интерфейс](#веб-интерфейс)
- [Режимы работы](#режимы-работы)
- [Командная строка](#команда-строка)
- [Диагностика](#диагностика)
- [Мониторинг](#мониторинг)
- [Тестирование](#тестирование)
- [Разработка](#разработка)
- [Документация](#документация)
- [Лицензия](#лицензия)

## Особенности

- ✅ **Два режима работы**: системный трей или служба Windows
- ✅ **Конкурентная обработка**: до 20 параллельных процессов
- ✅ **Многоуровневая фильтрация**: белые и черные списки
- ✅ **Веб-интерфейс**: SPA с React и WebSocket
- ✅ **Две СУБД**: DragonflyDB (кэш) + PostgreSQL (хранилище)
- ✅ **Метрики**: сбор системных и процессных метрик
- ✅ **Логирование**: централизованное с ротацией
- ✅ **Диагностика**: комплексная проверка системы
- ✅ **Экспорт**: данные в Excel, CSV, JSON, PDF
- ✅ **Zabbix**: экспорт метрик для мониторинга

## Архитектура

```
┌─────────────────────────────────────────────────────────────┐
│                      Venera Application                      │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │   Tshark     │  │   Tshark     │  │   Tshark     │       │
│  │ (Пакетный    │  │ (Пакетный    │  │ (Пакетный    │       │
│  │  сбор)       │  │  сбор)       │  │  сбор)       │       │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘       │
│         │                  │                  │               │
│         └──────────────────┴──────────────────┘               │
│                            │                                  │
│                    ┌───────▼───────┐                          │
│                    │  DragonflyDB  │                          │
│                    │   (Кэш, List) │                          │
│                    └───────┬───────┘                          │
│                            │                                  │
│                    ┌───────▼───────┐                          │
│                    │   PostgreSQL  │                          │
│                    │  (Хранилище)  │                          │
│                    └───────────────┘                          │
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │  Web Server  │  │   Metrics    │  │  Diagnose    │       │
│  │   (React UI) │  │  Collector   │  │   Module     │       │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘       │
│         │                  │                  │               │
│         └──────────────────┴──────────────────┘               │
│                            │                                  │
│                    ┌───────▼───────┐                          │
│                    │   Tray/       │                          │
│                    │   Service     │                          │
│                    └───────────────┘                          │
└─────────────────────────────────────────────────────────────┘
```

## Требования

### Системные требования

- **ОС**: Windows 10/11 x64
- **RAM**: минимум 4 GB (рекомендуется 8 GB+)
- **Диск**: минимум 20 GB свободного места
- **CPU**: 2+ ядра

### Программные зависимости

- **Go**: 1.21+
- **Tshark**: Wireshark (для анализа пакетов)
- **Podman**: latest (для контейнеров DragonflyDB)
- **PostgreSQL**: 14+ (для хранения данных)

## Установка

### Автоматическая установка

1. Загрузите последнюю версию: [Releases](https://github.com/...)

2. Запустите установочный скрипт с правами администратора:

```powershell
# Установка службы
.\install_venera.ps1

# Или установка в режиме трея
.\venera.exe
```

### Ручная установка

1. Скомпилируйте приложение:

```bash
go build -o venera.exe
```

2. Установите зависимости:

```bash
# Установка Podman
winget install Podman

# Установка Wireshark
winget install Wireshark

# Установка PostgreSQL
winget install PostgreSQL
```

3. Настройте базы данных:

```bash
# Создание контейнера DragonflyDB
.\venera.exe --create_cachedb

# Создание базы PostgreSQL
.\venera.exe --create_pg_db
```

## Конфигурация

### Основной файл конфигурации: `config.toml`

```toml
[Generic]
mode = "tray"  # или "service"
auto_start_processes = false
max_processes = 10
web_server_port = 8080

[Paths]
podman_path = "C:\\Program Files\\Podman\\podman.exe"
tshark_path = "C:\\Program Files\\Wireshark\\tshark.exe"

[DragonflyDB]
host = "localhost"
port = 6379

[PostgreSQL]
host = "localhost"
port = 5432
database = "Venera"
```

### Файл процессов: `processes.toml`

```toml
[process."net_001"]
type = "network"
name = "Сетевой поток"
ip = "192.168.1.100"
port = 5000

[process."folder_001"]
type = "folder"
name = "Папка с файлами"
path = "C:\\Logs"
scan_subfolders = true
monitor_new_files = true
```

### Файл фильтрации: `settings/generic.flt`

```
+key1|Название 1
-value1
-value2
+key2|Название 2
-value3
```

## Использование

### Запуск в режиме трея

```bash
./venera.exe
```

### Запуск в режиме службы

```bash
# Установка службы
./venera.exe --install_srv

# Запуск службы
net start VeneraSrv

# Остановка службы
net stop VeneraSrv
```

### Управление процессами

1. Откройте веб-интерфейс: `http://localhost:8080`
2. Перейдите на вкладку "Процессы"
3. Нажмите "Добавить процесс"
4. Выберите тип источника и настройте параметры
5. Нажмите "Старт" для запуска процесса

## Веб-интерфейс

Venera включает полнофункциональный веб-интерфейс с поддержкой SPA и WebSocket.

### Разделы

#### 1. Процессы (`/processes`)
- Управление процессами обработки
- Добавление новых процессов
- Старт/стоп/удаление процессов
- Сохранение конфигурации

#### 2. Статистика (`/statistics`)
- Метрики системы в реальном времени
- Загрузка CPU, RAM, диска
- Метрики процессов
- WebSocket-подключение для обновления

#### 3. База данных (`/db`)
- Просмотр данных из PostgreSQL
- Фильтрация и поиск
- Экспорт в Excel, CSV, JSON
- Пагинация результатов

#### 4. Настройки (`/settings`)
- Общие настройки
- Пути к ресурсам
- Настройки DragonflyDB
- Настройки PostgreSQL
- Логирование
- Проверка подключения

#### 5. Логи (`/logs`)
- Просмотр логов в реальном времени
- Фильтрация по уровню и тексту
- Поиск с подсветкой
- Экспорт в HTML, PDF
- Горячие клавиши (Ctrl+F, Esc)

#### 6. Диагностика (`/diagnose`)
- Запуск диагностики системы
- Проверка ресурсов
- Сетевые адаптеры
- Состояние служб
- История событий
- Экспорт отчетов

## Режимы работы

### Системный трей

Приложение запускается как приложение в системном трее с иконкой и контекстным меню.

```toml
[Generic]
mode = "tray"
```

### Служба Windows

Приложение запускается как служба Windows с автозапуском.

```toml
[Generic]
mode = "service"
```

Управление службой:

```bash
# Установка
venera.exe --install_srv

# Удаление
venera.exe --uninstall_srv

# Запуск
net start VeneraSrv

# Остановка
net stop VeneraSrv
```

## Команда строка

Venera поддерживает следующие параметры командной строки:

| Параметр | Описание |
|----------|----------|
| `-h, --help` | Показать справку |
| `-v, --version` | Показать версию |
| `-d, --diagnose` | Запустить диагностику |
| `-c, --create_cachedb` | Создать контейнер DragonflyDB |
| `-r, --remove_cachedb` | Удалить контейнер DragonflyDB |
| `-p, --create_pg_db` | Создать базу PostgreSQL |
| `-i, --install_srv` | Установить службу |
| `-u, --uninstall_srv` | Удалить службу |

Примеры:

```bash
# Показать справку
venera.exe --help

# Запустить диагностику
venera.exe --diagnose

# Создать контейнер
venera.exe --create_cachedb

# Установить службу
venera.exe --install_srv
```

## Диагностика

### Автоматическая диагностика

```bash
venera.exe --diagnose
```

### Через веб-интерфейс

1. Перейдите на вкладку "Диагностика"
2. Нажмите "Запустить диагностику"
3. Просмотрите результаты
4. Экспортируйте отчет в PDF или создайте архив

### Скрипт диагностики

```powershell
# С диаграфическим интерфейсом
diagnose-gui.ps1

# Без графического интерфейса
diagnose.ps1
```

## Мониторинг

### Сбор метрик

Venera собирает следующие метрики:

- **Системные**: CPU, RAM, диск, процессы
- **Сетевые**: пакеты, байты, дропы
- **Процессные**: скорость, потребление, статус

### Экспорт для Zabbix

```bash
# Получить метрики в формате Zabbix
curl http://localhost:8080/api/metrics/zabbix
```

### Структура метрик

```json
{
  "system": {
    "cpuUsage": 0.15,
    "ramUsage": 0.45,
    "diskUsage": 0.30
  },
  "processes": {
    "proc_001": {
      "packetsPerSecond": 1000,
      "ramUsage": 1048576,
      "cpuUsage": 0.05
    }
  },
  "network": {
    "packetsReceived": 1000000,
    "bytesReceived": 524288000
  }
}
```

## Тестирование

### Запуск тестов

```bash
# Юнит-тесты
go test ./...

# Интеграционные тесты
go test -tags=integration ./...

# Тесты производительности
go test -bench=. ./...

# Стресс-тесты
go test -tags=stress ./...
```

### Структура тестов

```
tests/
├── unit/          # Юнит-тесты
├── integration/   # Интеграционные тесты
├── performance/   # Тесты производительности
└── stress/        # Стресс-тесты
```

## Разработка

### Структура проекта

```
venera/
├── config/        # Конфигурация
├── data/          # Работа с данными
├── diagnose/      # Диагностика
├── logging/       # Логирование
├── manifest/      # Манифест
├── metrics/       # Метрики
├── models/        # Модели данных
├── notify/        # Уведомления
├── processes/     # Управление процессами
├── react-ui/      # React UI
├── services/      # Windows службы
├── settings/      # Настройки
├── sql/           # SQL скрипты
├── tests/         # Тесты
├── tray/          # Системный трей
├── utils/         # Утилиты
├── web/           # Веб-сервер
├── main.go        # Точка входа
└── config.toml    # Конфигурация
```

### Сборка

```bash
# Сборка
go build -o venera.exe

# Сборка с отладкой
go build -gcflags="all=-N -l" -o venera.exe

# Сборка для Windows x64
GOOS=windows GOARCH=amd64 go build -o venera.exe
```

## Документация

### Генерация документации

```bash
# Генерация Go documentation
godoc -http=:6060

# Просмотр документации
open http://localhost:6060/pkg/venera/
```

### Структура документации

- **README.md**: Основная документация (этот файл)
- **docs/**: Дополнительная документация
- **api/**: API спецификация
- **schema/**: Схемы баз данных

## Лицензия

MIT License

Copyright (c) 2024 Venera Team

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software.

## Поддержка

- **GitHub Issues**: [https://github.com/.../issues](https://github.com/.../issues)
- **Email**: support@venera.local
- **Documentation**: [https://docs.venera.local](https://docs.venera.local)

## Благодарности

Спасибо всем контрибьюторам и пользователям Venera!

---

**Venera** - Your packet data identifier collector