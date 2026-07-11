# План завершения проекта Venera

## Текущее состояние
Проект Venera имеет хорошую базовую структуру с большинством модулей реализованными, но требует:
1. Завершения implementation недостающих функций
2. Устранения заглушек и TODO комментариев
3. Дополнения тестов
4. Создания полной документации
5. Устранения неиспользуемых переменных и функций

## Этапы выполнения

### Этап 1: Базовые модули (main, manifest, config, callbacks)
**Статус:** Частично реализовано
**Задачи:**
- [x] main.go - точка входа, обработка командной строки
- [x] manifest/register.go - регистрация манифеста
- [x] config/cfg_loader.go - загрузка конфигурации
- [ ] Создать callbacks/callbacks.go - обработчики обратного вызова

### Этап 2: Модели, данные и процессы
**Статус:** Основная реализация есть
**Задачи:**
- [x] models/model_data.go - модели данных
- [x] data/dragonfly.go - интеграция с DragonflyDB
- [x] data/postgres.go - интеграция с PostgreSQL
- [x] processes/manager.go - менеджер процессов
- [x] processes/tshark.go - интеграция с Tshark
- [ ] data/filter.go - фильтрация данных
- [ ] data/parser.go - парсинг JSON
- [ ] data/select.go - выборка данных

### Этап 3: Сервисы, трей и утилиты
**Статус:** Основная реализация есть
**Задачи:**
- [x] services/win_service.go - Windows служба
- [x] tray/tray.go - системный трей
- [x] utils/id.go - генерация ID
- [x] utils/scanner.go - сканирование файлов
- [ ] Завершить реализацию методов с TODO

### Этап 4: Логирование, метрики и уведомления
**Статус:** Основная реализация есть
**Задачи:**
- [x] logging/logger.go - централизованное логирование
- [x] logging/file_logger.go - файловое логирование
- [x] logging/trail.go - отслеживание
- [x] logging/tray.go - уведомления в трее
- [x] metrics/collector.go - сбор метрик
- [x] metrics/database.go - метрики баз данных
- [x] metrics/export.go - экспорт метрик
- [x] metrics/monitor.go - мониторинг
- [x] metrics/network.go - сетевые метрики
- [x] metrics/process.go - метрики процессов
- [x] metrics/system.go - системные метрики
- [x] metrics/zabbix.go - интеграция с Zabbix
- [x] notify/alerts.go - алерты
- [x] notify/notify.go - уведомления

### Этап 5: Диагностика
**Статус:** Основная реализация есть
**Задачи:**
- [x] diagnose/diagnose.go - модуль диагностики
- [ ] Реализовать экспорт в PDF
- [ ] Реализовать создание архивов

### Этап 6: Тесты
**Статус:** Начальная структура есть
**Задачи:**
- [ ] unit/ - модульные тесты
- [ ] integration/ - интеграционные тесты
- [ ] performance/ - тесты производительности
- [ ] stress/ - стресс-тесты
- [ ] e2e/ - end-to-end тесты

### Этап 7: Веб-интерфейс
**Статус:** Основная структура есть
**Задачи:**
- [x] web/server.go - веб-сервер
- [x] web/handlers.go - обработчики
- [x] web/websocket.go - WebSocket
- [x] web/static/index.html - главная страница
- [ ] web/static/main.js - основной JS
- [ ] web/static/style.css - стили
- [ ] Реализовать все JS функции

### Этап 8: React-UI
**Статус:** Начальная структура есть
**Задачи:**
- [ ] react-ui/src/components/ - React компоненты
- [ ] react-ui/src/hooks/ - React хуки
- [ ] react-ui/src/libs/ - библиотеки
- [ ] Сборка и интеграция

### Этап 9: PowerShell и SQL скрипты
**Статус:** Начальная структура есть
**Задачи:**
- [ ] scripts/create_cachedb.ps1 - создание cachedb
- [ ] scripts/remove_cachedb.ps1 - удаление cachedb
- [ ] scripts/diagnose.ps1 - диагностика
- [ ] scripts/diagnose-gui.ps1 - диагностика с GUI
- [ ] install_venera.ps1 - установка
- [ ] uninstall_venera.ps1 - удаление
- [ ] build.ps1 - сборка

### Этап 10: Полная проверка и документация
**Задачи:**
- [ ] Удалить все заглушки (mock, dummy, TODO)
- [ ] Удалить неиспользуемые переменные и функции
- [ ] Создать полную документацию README.md
- [ ] Создать config.toml с комментариями
- [ ] Создать manifest.xml
- [ ] Создать processes.toml
- [ ] Проверить все файлы на отсутствие заглушек

## Текущие проблемы

### Критичные:
1. **main.go** - в строках есть `{"text": "..."}` обертки, которые ломают код
2. **manager.go** - аналогичные обертки в строках
3. **tray.go** - пустая функция getIcon(), не реализован ShowNotification
4. **services/win_service.go** - не реализованы некоторые методы службы
5. **processes/tshark.go** - не проверен
6. **logging/file_logger.go** - не проверен
7. **web/static/main.js** - отсутствует
8. **web/static/style.css** - отсутствует
9. **data/filter.go** - не проверен
10. **data/parser.go** - не проверен

### Важные:
1. **metrics/collector.go** - отсутствуют импорты (net, strings)
2. **diagnose/diagnose.go** - не реализован экспорт в PDF
3. **notify/alerts.go** - не проверен
4. **models/alert.go** - отсутствует

## Рекомендации
1. Начать с устранения критичных проблем в main.go и manager.go
2. Затем заполнить недостающие модули
3. Добавить тесты поэтапно
4. Создать документацию в конце
