// model_data.go - Модели данных для Venera
//
// Этот модуль содержит определения структур данных для хранения и обработки
// информации в системе сбора идентификаторов.
//
// Основные структуры:
// - DataRecord - запись данных
// - BatchData - пакет данных для вставки
// - Statistics - статистика данных
// - FilterRecord - запись фильтрации
// - ControlRecord - запись контроля
// - Alert - алерт

package models

// DataRecord - запись данных
type DataRecord struct {
	ID         int64  `json:"id" db:"id"`
	Source     string `json:"source" db:"source"`
	Key        string `json:"key" db:"key"`
	Value      string `json:"value" db:"value"`
	DateFirst  int64  `json:"date_first" db:"date_first"`
	DateLast   int64  `json:"date_last" db:"date_last"`
}

// BatchData - пакет данных для вставки
type BatchData struct {
	Source    string `json:"source"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
}

// Statistics - статистика данных
type Statistics struct {
	TotalCount   int64 `json:"total_count"`
	SourceCount  int64 `json:"source_count"`
	KeyCount     int64 `json:"key_count"`
	ValueCount   int64 `json:"value_count"`
	ProcessedCount int64 `json:"processed_count"`
	FilteredCount  int64 `json:"filtered_count"`
}

// FilterRecord - запись фильтрации
type FilterRecord struct {
	Key      string   `json:"key"`
	Values   []string `json:"values"`
	IsWhite  bool     `json:"is_white"`
}

// ControlRecord - запись контроля
type ControlRecord struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Alert - алерт
type Alert struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Condition    string                 `json:"condition"`
	Action       string                 `json:"action"`
	Severity     string                 `json:"severity"`
	Enabled      bool                   `json:"enabled"`
	CreatedAt    int64                  `json:"created_at"`
	LastTriggered int64                  `json:"last_triggered"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// AlertHistory - история алертов
type AlertHistory struct {
	AlertID     string `json:"alert_id"`
	Timestamp   int64  `json:"timestamp"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

// ProcessMetric - метрика процесса
type ProcessMetric struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	ProcessID       int    `json:"process_id"`
	InputRate       float64 `json:"input_rate"`
	MemoryUsage     int64  `json:"memory_usage"`
	CPUUsage        float64 `json:"cpu_usage"`
	ProcessedCount  int64  `json:"processed_count"`
	FilteredCount   int64  `json:"filtered_count"`
	QueueSize       int64  `json:"queue_size"`
	StartTime       int64  `json:"start_time"`
	LastActivity    int64  `json:"last_activity"`
}

// SystemMetric - системная метрика
type SystemMetric struct {
	Timestamp       int64   `json:"timestamp"`
	CPUUsage        float64 `json:"cpu_usage"`
	MemoryTotal     int64   `json:"memory_total"`
	MemoryUsed      int64   `json:"memory_used"`
	MemoryFree      int64   `json:"memory_free"`
	DiskTotal       int64   `json:"disk_total"`
	DiskUsed        int64   `json:"disk_used"`
	DiskFree        int64   `json:"disk_free"`
	NetworkReceived int64   `json:"network_received"`
	NetworkSent     int64   `json:"network_sent"`
}

// LogEntry - запись лога
type LogEntry struct {
	Timestamp int64  `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Module    string `json:"module"`
	FilePath  string `json:"file_path"`
	Line      int    `json:"line"`
}

// FilterConfig - конфигурация фильтрации
type FilterConfig struct {
	Whitelist map[string][]string `json:"whitelist"`
	Blacklist map[string][]string `json:"blacklist"`
}

// ProcessConfig - конфигурация процесса
type ProcessConfig struct {
	ID                 string `json:"id"`
	Type               string `json:"type"`
	Name               string `json:"name"`
	IP                 string `json:"ip"`
	Port               int    `json:"port"`
	Path               string `json:"path"`
	ScanSubfolders     bool   `json:"scan_subfolders"`
	MonitorNewFiles    bool   `json:"monitor_new_files"`
	Enabled            bool   `json:"enabled"`
	StartTime          int64  `json:"start_time"`
	LastActivity       int64  `json:"last_activity"`
	ProcessStatus      string `json:"process_status"`
}

// ProcessStatus - статус процесса
type ProcessStatus struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	InputRate    float64 `json:"input_rate"`
	ProcessedCount int64  `json:"processed_count"`
	StartTime    int64  `json:"start_time"`
}

// DatabaseConfig - конфигурация базы данных
type DatabaseConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Database     string `json:"database"`
	User         string `json:"user"`
	Password     string `json:"password"`
	MaxConnections int    `json:"max_connections"`
}

// DiagnosticsResult - результат диагностики
type DiagnosticsResult struct {
	Checks         []DiagnosticCheck `json:"checks"`
	Report         string            `json:"report"`
	ExportPath     string            `json:"export_path"`
	ArchivePath    string            `json:"archive_path"`
}

// DiagnosticCheck - проверка диагностики
type DiagnosticCheck struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Message  string `json:"message"`
	Duration int64  `json:"duration"`
}

// MetricExport - экспорт метрик
type MetricExport struct {
	Format     string                 `json:"format"`
	Data       map[string]interface{} `json:"data"`
	Timestamp  int64                  `json:"timestamp"`
	Source     string                 `json:"source"`
}
