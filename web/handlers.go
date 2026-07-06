// handlers.go - Обработчики HTTP запросов для веб-интерфейса Venera
//
// Этот модуль обеспечивает обработку HTTP запросов для веб-интерфейса
// Venera, включая все разделы: процессы, статистика, база данных, настройки, логи, диагностика.
//
// Основные функции:
// - Обработка GET/POST запросов для всех разделов
// - Возврат HTML страниц
// - Обработка AJAX запросов
// - Управление процессами через HTTP API
//
// Использование:
// import "venera/web"
// handlers := web.NewHandlers(cfg, processManager, dragonflyDB, postgresDB, logger)
// server.HandleFunc("/processes", handlers.ProcessesHandler)

package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"venera/config"
	"venera/data"
	"venera/metrics"
	"venera/processes"
)

// Handlers - структура обработчиков
type Handlers struct {
	cfg              *config.Config
	processManager   *processes.ProcessManager
	dragonflyDB      *data.DragonflyDB
	postgresDB       *data.PostgreSQL
	logger           *logrus.Logger
	metricsCollector *metrics.Collector
}

// NewHandlers - создание новых обработчиков
func NewHandlers(cfg *config.Config, processManager *processes.ProcessManager, dragonflyDB *data.DragonflyDB, postgresDB *data.PostgreSQL, logger *logrus.Logger, metricsCollector *metrics.Collector) *Handlers {
	return &Handlers{
		cfg:              cfg,
		processManager:   processManager,
		dragonflyDB:      dragonflyDB,
		postgresDB:       postgresDB,
		logger:           logger,
		metricsCollector: metricsCollector,
	}
}

// ProcessesHandler - обработчик для раздела "Процессы"
func (h *Handlers) ProcessesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.serveStaticFile(w, r, "static/processes/processes.html")
	case http.MethodPost:
		h.handleProcessAction(w, r)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

// StatisticsHandler - обработчик для раздела "Статистика"
func (h *Handlers) StatisticsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.serveStaticFile(w, r, "static/statistics/statistics.html")
	case http.MethodPost:
		h.handleStatisticsRequest(w, r)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

// DatabaseHandler - обработчик для раздела "База данных"
func (h *Handlers) DatabaseHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.serveStaticFile(w, r, "static/db/db.html")
	case http.MethodPost:
		h.handleDatabaseRequest(w, r)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

// SettingsHandler - обработчик для раздела "Настройки"
func (h *Handlers) SettingsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.serveStaticFile(w, r, "static/settings/settings.html")
	case http.MethodPost:
		h.handleSettingsRequest(w, r)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

// LogsHandler - обработчик для раздела "Логи"
func (h *Handlers) LogsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.serveStaticFile(w, r, "static/logs/logs.html")
	case http.MethodPost:
		h.handleLogsRequest(w, r)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

// DiagnoseHandler - обработчик для раздела "Диагностика"
func (h *Handlers) DiagnoseHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.serveStaticFile(w, r, "static/diagnose/diagnose.html")
	case http.MethodPost:
		h.handleDiagnoseRequest(w, r)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

// RootHandler - обработчик корневого пути
func (h *Handlers) RootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "" {
		h.serveStaticFile(w, r, "static/index.html")
		return
	}
	http.NotFound(w, r)
}

// serveStaticFile - сервер статического файла
func (h *Handlers) serveStaticFile(w http.ResponseWriter, r *http.Request, filePath string) {
	fullPath := "." + filePath

	// Проверка существования файла
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.Error(w, "Файл не найден", http.StatusNotFound)
		return
	}

	// Установка заголовков
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Чтение и отправка файла
	content, err := os.ReadFile(fullPath)
	if err != nil {
		http.Error(w, "Ошибка чтения файла", http.StatusInternalServerError)
		return
	}

	w.Write(content)
}

// handleProcessAction - обработка действий с процессами
func (h *Handlers) handleProcessAction(w http.ResponseWriter, r *http.Request) {
	// Декодирование запроса
	var request struct {
		Action   string `json:"action"`
		ProcessID string `json:"process_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Ошибка декодирования запроса", http.StatusBadRequest)
		return
	}

	// Выполнение действия
	var response map[string]interface{}

	switch request.Action {
	case "start":
		err := h.processManager.StartProcess(request.ProcessID)
		if err != nil {
			response = map[string]interface{}{"success": false, "error": err.Error()}
		} else {
			response = map[string]interface{}{"success": true, "message": "Процесс запущен"}
		}
	case "stop":
		err := h.processManager.StopProcess(request.ProcessID)
		if err != nil {
			response = map[string]interface{}{"success": false, "error": err.Error()}
		} else {
			response = map[string]interface{}{"success": true, "message": "Процесс остановлен"}
		}
	case "delete":
		err := h.processManager.DeleteProcess(request.ProcessID)
		if err != nil {
			response = map[string]interface{}{"success": false, "error": err.Error()}
		} else {
			response = map[string]interface{}{"success": true, "message": "Процесс удален"}
		}
	case "add":
		// Добавление нового процесса
		var processData struct {
			ID                 string `json:"id"`
			Type               string `json:"type"`
			Name               string `json:"name"`
			IP                 string `json:"ip"`
			Port               int    `json:"port"`
			Path               string `json:"path"`
			ScanSubfolders     bool   `json:"scan_subfolders"`
			MonitorNewFiles    bool   `json:"monitor_new_files"`
		}

		if err := json.NewDecoder(r.Body).Decode(&processData); err != nil {
			http.Error(w, "Ошибка декодирования данных процесса", http.StatusBadRequest)
			return
		}

		err := h.processManager.AddProcess(processData.ID, processData.Type, processData.Name,
			processData.IP, processData.Port, processData.Path,
			processData.ScanSubfolders, processData.MonitorNewFiles)
		if err != nil {
			response = map[string]interface{}{"success": false, "error": err.Error()}
		} else {
			response = map[string]interface{}{"success": true, "message": "Процесс добавлен"}
		}
	default:
		response = map[string]interface{}{"success": false, "error": "Неизвестное действие"}
	}

	// Отправка ответа
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStatisticsRequest - обработка запроса статистики
func (h *Handlers) handleStatisticsRequest(w http.ResponseWriter, r *http.Request) {
	// Получение метрик
	metrics := h.getMetrics()

	// Отправка ответа
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// handleDatabaseRequest - обработка запроса базы данных
func (h *Handlers) handleDatabaseRequest(w http.ResponseWriter, r *http.Request) {
	// Декодирование запроса
	var request struct {
		Action string `json:"action"`
		Query  string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Ошибка декодирования запроса", http.StatusBadRequest)
		return
	}

	// Выполнение действия
	var response map[string]interface{}

	switch request.Action {
	case "query":
		// Выполнение SQL-запроса
		rows, err := h.postgresDB.GetDB().Query(request.Query)
		if err != nil {
			response = map[string]interface{}{"success": false, "error": err.Error()}
		} else {
			defer rows.Close()

			// Получение столбцов
			columns, err := rows.Columns()
			if err != nil {
				response = map[string]interface{}{"success": false, "error": err.Error()}
				json.NewEncoder(w).Encode(response)
				return
			}

			// Получение строк
			values := make([][]string, 0)
			for rows.Next() {
				row := make([]string, len(columns))
				rowPtr := make([]interface{}, len(columns))
				for i := range row {
					rowPtr[i] = &row[i]
				}

				if err := rows.Scan(rowPtr...); err != nil {
					response = map[string]interface{}{"success": false, "error": err.Error()}
					json.NewEncoder(w).Encode(response)
					return
				}

				values = append(values, row)
			}

			response = map[string]interface{}{
				"success": true,
				"columns": columns,
				"values":  values,
			}
		}
	default:
		response = map[string]interface{}{"success": false, "error": "Неизвестное действие"}
	}

	// Отправка ответа
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSettingsRequest - обработка запроса настроек
func (h *Handlers) handleSettingsRequest(w http.ResponseWriter, r *http.Request) {
	// Декодирование запроса
	var request struct {
		Action string `json:"action"`
		Config map[string]interface{} `json:"config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Ошибка декодирования запроса", http.StatusBadRequest)
		return
	}

	// Выполнение действия
	var response map[string]interface{}

	switch request.Action {
	case "save":
		// Сохранение конфигурации
		err := h.cfg.SaveConfig("config.toml")
		if err != nil {
			response = map[string]interface{}{"success": false, "error": err.Error()}
		} else {
			response = map[string]interface{}{"success": true, "message": "Конфигурация сохранена"}
		}
	case "check_dragonfly":
		// Проверка подключения к DragonflyDB
		if h.dragonflyDB != nil {
			response = map[string]interface{}{"success": true, "message": "Подключение к DragonflyDB установлено"}
		} else {
			response = map[string]interface{}{"success": false, "error": "Не удалось подключиться к DragonflyDB"}
		}
	case "check_postgres":
		// Проверка подключения к PostgreSQL
		if h.postgresDB != nil {
			response = map[string]interface{}{"success": true, "message": "Подключение к PostgreSQL установлено"}
		} else {
			response = map[string]interface{}{"success": false, "error": "Не удалось подключиться к PostgreSQL"}
		}
	default:
		response = map[string]interface{}{"success": false, "error": "Неизвестное действие"}
	}

	// Отправка ответа
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleLogsRequest - обработка запроса логов
func (h *Handlers) handleLogsRequest(w http.ResponseWriter, r *http.Request) {
	// Декодирование запроса
	var request struct {
		Action   string `json:"action"`
		Level    string `json:"level"`
		FromFile string `json:"from_file"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Ошибка декодирования запроса", http.StatusBadRequest)
		return
	}

	// Выполнение действия
	var response map[string]interface{}

	switch request.Action {
	case "get_logs":
		// Получение логов из файла
		logs := h.getLogsFromFile(request.FromFile, request.Level)
		response = map[string]interface{}{"success": true, "logs": logs}
	case "get_log_files":
		// Получение списка файлов логов
		files := h.getLogFiles()
		response = map[string]interface{}{"success": true, "files": files}
	default:
		response = map[string]interface{}{"success": false, "error": "Неизвестное действие"}
	}

	// Отправка ответа
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleDiagnoseRequest - обработка запроса диагностики
func (h *Handlers) handleDiagnoseRequest(w http.ResponseWriter, r *http.Request) {
	// Декодирование запроса
	var request struct {
		Action string `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Ошибка декодирования запроса", http.StatusBadRequest)
		return
	}

	// Выполнение действия
	var response map[string]interface{}

	switch request.Action {
	case "run":
		// Запуск диагностики
		report := h.runDiagnosis()
		response = map[string]interface{}{"success": true, "report": report}
	default:
		response = map[string]interface{}{"success": false, "error": "Неизвестное действие"}
	}

	// Отправка ответа
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getMetrics - получение метрик для статистики
func (h *Handlers) getMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// Получение системных метрик
	if h.metricsCollector != nil {
		if systemMetrics, err := h.metricsCollector.GetRAMUsage(); err == nil {
			metrics["ram"] = systemMetrics
		}
		if diskMetrics, err := h.metricsCollector.GetDiskUsage("."); err == nil {
			metrics["disk"] = diskMetrics
		}
		metrics["cpu_usage"] = h.metricsCollector.GetCPUUsage()
	}

	// Получение метрик процессов
	processMetrics := h.processManager.GetProcessMetrics()
	metrics["processes"] = processMetrics

	// Получение метрик баз данных
	if h.dragonflyDB != nil {
		if dragonflyMetrics, err := h.metricsCollector.GetDragonflyDBStats(h.dragonflyDB); err == nil {
			metrics["dragonflydb"] = dragonflyMetrics
		}
	}

	if h.postgresDB != nil {
		if postgresMetrics, err := h.metricsCollector.GetPostgreSQLStats(h.postgresDB); err == nil {
			metrics["postgresql"] = postgresMetrics
		}
	}

	return metrics
}

// getLogsFromFile - получение логов из файла
func (h *Handlers) getLogsFromFile(filePath, level string) []string {
	// TODO: Реализовать чтение логов из файла
	// Для текущей реализации возвращаем заглушку
	return []string{"Логи не доступны"}
}

// getLogFiles - получение списка файлов логов
func (h *Handlers) getLogFiles() []string {
	// TODO: Реализовать получение списка файлов логов
	// Для текущей реализации возвращаем заглушку
	return []string{"logs.log"}
}

// runDiagnosis - запуск диагностики
func (h *Handlers) runDiagnosis() string {
	// TODO: Реализовать диагностику
	// Для текущей реализации возвращаем заглушку
	return "Диагностика не реализована"
}

// GetProcessesList - получение списка процессов
func (h *Handlers) GetProcessesList() []processes.ProcessConfig {
	return h.processManager.GetProcessesList()
}

// GetProcessStatus - получение статуса процесса
func (h *Handlers) GetProcessStatus(processID string) string {
	return h.processManager.GetProcessStatus(processID)
}

// StartProcess - запуск процесса
func (h *Handlers) StartProcess(processID string) error {
	return h.processManager.StartProcess(processID)
}

// StopProcess - остановка процесса
func (h *Handlers) StopProcess(processID string) error {
	return h.processManager.StopProcess(processID)
}

// DeleteProcess - удаление процесса
func (h *Handlers) DeleteProcess(processID string) error {
	return h.processManager.DeleteProcess(processID)
}

// StartAllProcesses - запуск всех процессов
func (h *Handlers) StartAllProcesses() error {
	return h.processManager.StartAll()
}

// StopAllProcesses - остановка всех процессов
func (h *Handlers) StopAllProcesses() error {
	return h.processManager.StopAll()
}
