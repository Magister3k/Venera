// model_data.go - Модели данных для API и интерфейса
//
// Этот модуль содержит структуры данных для обмена между
// сервером и клиентом, а также для работы с базой данных.
//
// Основные функции:
// - Модели для JSON сериализации
// - Модели для веб-интерфейса
// - Модели для обработки данных
//
// Использование:
// import "venera/models"
// record := models.DataRecord{Source: "camera1", Key: "id", Value: "12345"}

package models

import (
	"time"
)

// DataRecord - запись данных
type DataRecord struct {
	ID         int64     `json:"id"`
	Source     string    `json:"source"`
	Key        string    `json:"key"`
	Value      string    `json:"value"`
	DateFirst  time.Time `json:"date_first"`
	DateLast   time.Time `json:"date_last"`
}

// DataFilter - фильтр данных
type DataFilter struct {
	Source     string `json:"source"`
	Key        string `json:"key"`
	Value      string `json:"value"`
	DateFrom   string `json:"date_from"`
	DateTo     string `json:"date_to"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
}

// DataStatistics - статистика данных
type DataStatistics struct {
	TotalCount     int64 `json:"total_count"`
	SourceCount    int64 `json:"source_count"`
	KeyCount       int64 `json:"key_count"`
	ValueCount     int64 `json:"value_count"`
	PeriodFirst    string `json:"period_first"`
	PeriodLast     string `json:"period_last"`
}

// DataExportRequest - запрос на экспорт данных
type DataExportRequest struct {
	Format     string `json:"format"` // excel, csv, json
	Filter     DataFilter `json:"filter"`
}

// DataExportResponse - ответ экспорта данных
type DataExportResponse struct {
	Success    bool   `json:"success"`
	Filename   string `json:"filename"`
	DownloadURL string `json:"download_url"`
	Error      string `json:"error,omitempty"`
}

// DataBatch - пакет данных для вставки
type DataBatch struct {
	Records []DataRecord `json:"records"`
	Total   int          `json:"total"`
}

// Response - стандартный ответ API
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PaginatedResponse - пагинированный ответ
type PaginatedResponse struct {
	Total     int         `json:"total"`
	Page      int         `json:"page"`
	PageSize  int         `json:"page_size"`
	Data      interface{} `json:"data"`
	HasMore   bool        `json:"has_more"`
}

// Statistics - общая статистика системы
type Statistics struct {
	Processes     int64 `json:"processes"`
	ActiveProcesses   int64 `json:"active_processes"`
	DataRecords     int64 `json:"data_records"`
	QueueSize       int64 `json:"queue_size"`
	ProcessingRate  float64 `json:"processing_rate"`
	MemoryUsage     int64 `json:"memory_usage"`
	CPUUsage        float64 `json:"cpu_usage"`
	DiskUsage       int64 `json:"disk_usage"`
	DiskFree        int64 `json:"disk_free"`
}

// ProcessStatistics - статистика процесса
type ProcessStatistics struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Type            string  `json:"type"`
	Status          string  `json:"status"`
	RecordsProcessed int64  `json:"records_processed"`
	RecordsFiltered int64  `json:"records_filtered"`
	RecordsAdded    int64  `json:"records_added"`
	StartTime       string  `json:"start_time"`
	LastActivity    string  `json:"last_activity"`
	MemoryUsage     int64   `json:"memory_usage"`
	CPUUsage        float64 `json:"cpu_usage"`
	InputRate       float64 `json:"input_rate"`
}

// LogMessage - сообщение лога
type LogMessage struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Source    string `json:"source"`
}

// LogFilter - фильтр логов
type LogFilter struct {
	Level     string `json:"level"`
	FromDate  string `json:"from_date"`
	ToDate    string `json:"to_date"`
	Search    string `json:"search"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}

// ConfigOption - опция конфигурации
type ConfigOption struct {
	Name        string      `json:"name"`
	Value       interface{} `json:"value"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Options     []string    `json:"options,omitempty"`
}

// ConfigSection - раздел конфигурации
type ConfigSection struct {
	Name        string         `json:"name"`
	Options     []ConfigOption `json:"options"`
	Description string         `json:"description"`
}

// DiagnosticReport - отчет диагностики
type DiagnosticReport struct {
	AppVersion    string `json:"app_version"`
	ManifestValid bool   `json:"manifest_valid"`
	ManifestVersion string `json:"manifest_version"`
	MemoryTotal   int64  `json:"memory_total"`
	MemoryFree    int64  `json:"memory_free"`
	DiskTotal     int64  `json:"disk_total"`
	DiskFree      int64  `json:"disk_free"`
	DragonflyDB   string `json:"dragonflydb"`
	PostgreSQL    string `json:"postgresql"`
	Tshark        string `json:"tshark"`
	Podman        string `json:"podman"`
	NetworkAdapters []NetworkAdapter `json:"network_adapters"`
	Errors        []string `json:"errors"`
}

// NetworkAdapter - сетевой адаптер
type NetworkAdapter struct {
	Name        string `json:"name"`
	IPAddress   string `json:"ip_address"`
	MACAddress  string `json:"mac_address"`
	Status      string `json:"status"`
}

// Alert - алерт
type Alert struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Condition   string `json:"condition"`
	Enabled     bool   `json:"enabled"`
	LastTrigger string `json:"last_trigger"`
}

// Manifest - манифест приложения
type Manifest struct {
	Version     string `json:"version"`
	BuildDate   string `json:"build_date"`
	Commit      string `json:"commit"`
	Environment string `json:"environment"`
}

// Notification - уведомление
type Notification struct {
	Title       string `json:"title"`
	Message     string `json:"message"`
	Type        string `json:"type"` // info, warning, error
	Duration    int    `json:"duration"`
	AutoClose   bool   `json:"auto_close"`
}

// ProcessActionRequest - запрос действия над процессом
type ProcessActionRequest struct {
	ID     string `json:"id"`
	Action string `json:"action"` // start, stop, delete, restart
}

// ProcessActionResponse - ответ действия над процессом
type ProcessActionResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
	Action  string `json:"action"`
	Message string `json:"message,omitempty"`
}

// WebSocketMessage - сообщение WebSocket
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	Timestamp string    `json:"timestamp"`
}

// WebSocketSubscribe - подписка на WebSocket
type WebSocketSubscribe struct {
	Type string `json:"type"` // statistics, logs
}

// WebSocketUnsubscribe - отписка от WebSocket
type WebSocketUnsubscribe struct {
	Type string `json:"type"` // statistics, logs
}
