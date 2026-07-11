# План завершения разработки Venera

## Статус проекта

### Выявленные проблемы

1. **Синтаксические ошибки в существующих файлах:**
   - `diagnose/diagnose.go` - импорты в конце файла (после кода)
   - `services/win_service.go` - синтаксические ошибки в структуре служб
   - `processes/manager.go` - синтаксические ошибки и лишние символы

2. **Отсутствующие файлы веб-интерфейса:**
   - `web/static/processes/processes.html`
   - `web/static/logs/logs.html`
   - `web/static/settings/settings.html`
   - `web/static/diagnose/diagnose.html`
   - `web/static/db/db.html`
   - `web/static/statistics/statistics.html`
   - CSS и JS файлы для всех разделов

3. **Неполные реализации модулей:**
   - `data/filter.go` - фильтрация данных
   - `data/parser.go` - парсинг JSON
   - `data/select.go` - выборка данных
   - `logging/file_logger.go` - файловый логгер
   - `logging/tray.go` - уведомления в трее
   - `manifest/register.go` - регистрация манифеста
   - `metrics/*.go` - сбор метрик
   - `notify/alerts.go`, `notify/notify.go` - уведомления
   - `utils/*.go` - утилиты

4. **Отсутствующие файлы:**
   - `sql/create_pg_db.sql` - SQL-скрипт создания БД
   - `scripts/*.ps1` - PowerShell скрипты
   - `react-ui/` - React-интерфейс
   - `tests/*` - тесты

---

## Пошаговый план исправления

### Этап 1: Исправление критических ошибок (Приоритет: КРИТИЧЕСКИЙ)

#### Задача 1.1: Исправить diagnose.go
- Переместить импорты в начало файла
- Исправить синтаксис (unsafe, context, syscall)
- Завершить реализацию всех методов диагностики
- Удалить TODO комментарии

#### Задача 1.2: Исправить win_service.go
- Исправить синтаксические ошибки в структуре ServiceHandler
- Завершить реализацию всех методов службы
- Удалить TODO комментарии

#### Задача 1.3: Исправить manager.go
- Исправить синтаксические ошибки
- Завершить реализацию ProcessManager
- Добавить недостающие методы

---

### Этап 2: Создание веб-интерфейса (Приоритет: ВЫСОКИЙ)

#### Задача 2.1: Создать CSS файлы
- `web/static/style.css` - основные стили
- `web/static/processes/processes.css`
- `web/static/logs/logs.css`
- `web/static/settings/settings.css`
- `web/static/diagnose/diagnose.css`
- `web/static/db/db.css`
- `web/static/statistics/statistics.css`

#### Задача 2.2: Создать HTML файлы
- `web/static/processes/processes.html`
- `web/static/logs/logs.html`
- `web/static/settings/settings.html`
- `web/static/diagnose/diagnose.html`
- `web/static/db/db.html`
- `web/static/statistics/statistics.html`

#### Задача 2.3: Создать JavaScript файлы
- `web/static/main.js` - основной JS
- `web/static/processes/processes.js`
- `web/static/logs/logs.js`
- `web/static/settings/settings.js`
- `web/static/diagnose/diagnose.js`
- `web/static/db/db.js`
- `web/static/statistics/statistics.js`

---

### Этап 3: Завершение модулей (Приоритет: ВЫСОКИЙ)

#### Задача 3.1: Завершить data модули
- `data/filter.go` - фильтрация по белому/черному спискам
- `data/parser.go` - парсинг JSON с поддержкой ":"
- `data/select.go` - выборка данных из DragonflyDB

#### Задача 3.2: Завершить logging модули
- `logging/file_logger.go` - логирование с ротацией и сжатием
- `logging/tray.go` - уведомления в системном трее

#### Задача 3.3: Завершить notify модули
- `notify/alerts.go` - обработка алертов
- `notify/notify.go` - отправка уведомлений

#### Задача 3.4: Завершить metrics модули
- `metrics/collector.go` - сбор системных метрик
- `metrics/database.go` - метрики БД
- `metrics/export.go` - экспорт метрик
- `metrics/monitor.go` - мониторинг процессов
- `metrics/network.go` - сетевые метрики
- `metrics/process.go` - метрики процессов
- `metrics/system.go` - системные метрики
- `metrics/zabbix.go` - экспорт для Zabbix

#### Задача 3.5: Завершить utils модули
- `utils/id.go` - генерация ID
- `utils/scanner.go` - сканирование файлов

#### Задача 3.6: Завершить manifest модуль
- `manifest/register.go` - регистрация манифеста

---

### Этап 4: Создание конфигурационных файлов (Приоритет: СРЕДНИЙ)

#### Задача 4.1: Создать config.toml с комментариями
- Полная конфигурация со всеми разделами
- Подробные комментарии на русском языке

#### Задача 4.2: Создать processes.toml
- Шаблон конфигурации процессов

#### Задача 4.3: Создать файлы фильтрации
- `settings/generic.flt` - белый/черный список
- `settings/generic.ctr` - контрольные значения
- `settings/generic.alr` - алерты

---

### Этап 5: Создание SQL-скриптов (Приоритет: СРЕДНИЙ)

#### Задача 5.1: Создать SQL-скрипт
- `sql/create_pg_db.sql` - создание базы PostgreSQL
- Таблица data с полями: source, key, value, date_first, date_last

---

### Этап 6: Создание скриптов PowerShell (Приоритет: СРЕДНИЙ)

#### Задача 6.1: Создать скрипты
- `scripts/create_cachedb.ps1` - создание контейнера
- `scripts/remove_cachedb.ps1` - удаление контейнера
- `scripts/diagnose.ps1` - диагностика
- `scripts/diagnose-gui.ps1` - диагностика с GUI (WinForms)
- `scripts/install_venera.ps1` - установка службы
- `scripts/uninstall_venera.ps1` - удаление службы

---

### Этап 7: Создание React-UI (Приоритет: ВЫСОКИЙ)

#### Задача 7.1: Создать структуру React
- `react-ui/src` - исходный код
- `react-ui/public` - статические файлы
- `react-ui/dst` - собранные файлы

#### Задача 7.2: Реализовать компоненты
- Компоненты для всех разделов веб-интерфейса
- Хуки для работы с API
- Библиотеки для стилей и компонентов

---

### Этап 8: Создание тестов (Приоритет: СРЕДНИЙ)

#### Задача 8.1: Создать структуру тестов
- `tests/unit` - модульные тесты
- `tests/integration` - интеграционные тесты
- `tests/performance` - тесты производительности
- `tests/stress` - стресс-тесты
- `tests/e2e` - end-to-end тесты для основных сценариев

---

### Этап 9: Создание документации (Приоритет: СРЕДНИЙ)

#### Задача 9.1: Создать README.md
- Схема проекта
- Функции программы
- Логическая схема бизнес-процессов
- Описание интерфейса
- Описание параметров командной строки
- Инструкция по развёртыванию
- Инструкции по настройке СУБД
- Инструкция по настройке Zabbix
- Описание тестов

---

### Этап 10: Финальная проверка (Приоритет: ВЫСОКИЙ)

#### Задача 10.1: Проверка на заглушки
- Удалить все `//TODO`, `//в проде...`, `//заглушка`
- Удалить неиспользуемые переменные и функции

#### Задача 10.2: Проверка зависимостей
- Убедиться, что все зависимости установлены
- Обновить go.mod

#### Задача 10.3: Сборка и тестирование
- Сборка проекта
- Проверка запуска
- Проверка веб-интерфейса

---

## Ожидаемые результаты

После выполнения плана:
1. Все синтаксические ошибки будут исправлены
2. Веб-интерфейс будет полностью работоспособен
3. Все модули будут реализованы
4. Конфигурационные файлы будут готовы
5. SQL-скрипты и PowerShell скрипты будут созданы
6. React-UI будет реализован
7. Тесты будут созданы
8. Полная документация будет готова
9. Проект будет готов к сборке и развертыванию

---

## Примечания

- Все комментарии должны быть на русском языке
- Использовать библиотеки: getlantern/systray, pgx v5, fastjson, logrus
- Поддерживать Windows x64 архитектуру
- Соблюдать структуру проекта из технического задания
- React-UI должен быть полным (сборка, компоненты, хуки)
- PowerShell скрипты должны быть обоих типов: консольные и с GUI (WinForms)
- End-to-end тесты для основных сценариев являются приоритетными