// model_process.go - Модели процессов для Venera
//
// Этот модуль содержит структуры данных для описания и управления
// процессами обработки данных в системе Venera.
//
// Основные структуры:
// - Process - описание процесса
// - ProcessConfig - конфигурация процесса
// - ProcessStatus - статус процесса
// - ProcessMetrics - метрики процесса
//
// Использование:
// import "venera/models"
// process := models.Process{ID: "proc_001", Name: "Network Stream"}

package models

import (
	"time"
)

// Process - описание процесса
type Process struct {
	ID                 string        `json:"id"`
	Type               string        `json:"type"` // network, folder, file
	Name               string        `json:"name"`
	SourceIP           string        `json:"source_ip"`
	SourcePort         int           `json:"source_port"`
	SourcePath         string        `json:"source_path"`
	ScanSubfolders     bool          `json:"scan_subfolders"`
	MonitorNewFiles    bool          `json:"monitor_new_files"`
	Enabled            bool          `json:"enabled"`
	StartTime          time.Time     `json:"start_time"`
	LastActivity       time.Time     `json:"last_activity"`
	Status             string        `json:"status"` // running, stopped, error
	Config             ProcessConfig `json:"config"`
	Statistics         ProcessStats  `json:"statistics"`
	QueueName          string        `json:"queue_name"`
	SortedSetName      string        `json:"sorted_set_name"`
	ChannelName        string        `json:"channel_name"`
	CounterName        string        `json:"counter_name"`
	TimerName          string        `json:"timer_name"`
}

// ProcessConfig - конфигурация процесса
type ProcessConfig struct {
	ID                 string `json:"id"`
	Type               string `json:"type"`
	Name               string `json:"name"`
	SourceIP           string `json:"source_ip"`
	SourcePort         int    `json:"source_port"`
	SourcePath         string `json:"source_path"`
	ScanSubfolders     bool   `json:"scan_subfolders"`
	MonitorNewFiles    bool   `json:"monitor_new_files"`
	MaxQueueSize       int    `json:"max_queue_size"`
	ProcessingInterval int    `json:"processing_interval"`
}

// ProcessStats - статистика процесса
type ProcessStats struct {
	RecordsProcessed   int64 `json:"records_processed"`
	RecordsFiltered    int64 `json:"records_filtered"`
	RecordsAdded       int64 `json:"records_added"`
	QueueSize          int64 `json:"queue_size"`
	ProcessingRate     float64 `json:"processing_rate"`
	InputRate          float64 `json:"input_rate"`
	MemoryUsage        int64 `json:"memory_usage"`
	CPUUsage           float64 `json:"cpu_usage"`
	Duration           int64 `json:"duration"`
}

// ProcessStatus - статус процесса
type ProcessStatus struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Status       string `json:"status"`
	RecordsProcessed int64 `json:"records_processed"`
	QueueSize    int64  `json:"queue_size"`
	StartTime    string `json:"start_time"`
	LastActivity string `json:"last_activity"`
}

// ProcessMetrics - метрики процесса
type ProcessMetrics struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	Type               string  `json:"type"`
	Status             string  `json:"status"`
	RecordsProcessed   int64   `json:"records_processed"`
	RecordsFiltered    int64   `json:"records_filtered"`
	RecordsAdded       int64   `json:"records_added"`
	InputRate          float64 `json:"input_rate"`
	ProcessingRate     float64 `json:"processing_rate"`
	MemoryUsage        int64   `json:"memory_usage"`
	CPUUsage           float64 `json:"cpu_usage"`
	DiskUsage          int64   `json:"disk_usage"`
	QueueSize          int64   `json:"queue_size"`
	ActiveTime         int64   `json:"active_time"`
}

// ProcessAction - действие над процессом
type ProcessAction struct {
	ID     string `json:"id"`
	Action string `json:"action"` // start, stop, restart, delete
}

// ProcessList - список процессов"
type ProcessList struct {
	Processes []Process `json:"processes"`
	Total     int       `json:"total"`
	Active    int       `json:"active"`
}

// ProcessBatch - пакет процессов
type ProcessBatch struct {
	Processes []Process `json:"processes"`
	Total     int       `json:"total"`
}

// ProcessFilter - фильтр процессов
type ProcessFilter struct {
	Status string `json:"status"`
	Type   string `json:"type"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

// ProcessUpdateRequest - запрос обновления процесса
type ProcessUpdateRequest struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	SourceIP           string `json:"source_ip"`
	SourcePort         int    `json:"source_port"`
	SourcePath         string `json:"source_path"`
	ScanSubfolders     bool   `json:"scan_subfolders"`
	MonitorNewFiles    bool   `json:"monitor_new_files"`
	Enabled            bool   `json:"enabled"`
}

// ProcessUpdateResponse - ответ обновления процесса
type ProcessUpdateResponse struct {
	Success bool    `json:"success"`
	Process Process `json:"process,omitempty"`
	Error   string  `json:"error,omitempty"`
}

// ProcessEvent - событие процесса
type ProcessEvent struct {
	Type       string `json:"type"` // started, stopped, error, updated
	ProcessID  string `json:"process_id"`
	Timestamp  string `json:"timestamp"`
	Message    string `json:"message,omitempty"`
}

// ProcessQueueStats - статистика очереди процесса
type ProcessQueueStats struct {
	QueueName       string `json:"queue_name"`
	QueueSize       int64  `json:"queue_size"`
	QueueMaxSize    int64  `json:"queue_max_size"`
	QueueFullRate   int64  `json:"queue_full_rate"`
}

// ProcessTimerStats - статистика таймера процесса
type ProcessTimerStats struct {
	TimerName        string `json:"timer_name"`
	Interval         int    `json:"interval"`
	LastTrigger      string `json:"last_trigger"`
	NextTrigger      string `json:"next_trigger"`
	TriggerCount     int64  `json:"trigger_count"`
}

// ProcessBatchStats - статистика пакетной обработки
type ProcessBatchStats struct {
	BatchSize       int64  `json:"batch_size"`
	BatchMaxSize    int64  `json:"batch_max_size"`
	BatchCount      int64  `json:"batch_count"`
	BatchAvgSize    float64 `json:"batch_avg_size"`
}

// ProcessMetricsSummary - суммарные метрики процессов
type ProcessMetricsSummary struct {
	TotalProcesses    int    `json:"total_processes"`
	ActiveProcesses   int    `json:"active_processes"`
	TotalRecords      int64  `json:"total_records"`
	AvgProcessingRate float64 `json:"avg_processing_rate"`
	TotalMemoryUsage  int64  `json:"total_memory_usage"`
	TotalCPUUsage     float64 `json:"total_cpu_usage"`
}

// ProcessHealth - состояние здоровья процесса
type ProcessHealth struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"` // healthy, warning, critical
	LastCheck   string `json:"last_check"`
	Errors      int    `json:"errors"`
	Warnings    int    `json:"warnings"`
}

// ProcessConfigDiff - различия в конфигурации процесса
type ProcessConfigDiff struct {
	OldConfig ProcessConfig `json:"old_config"`
	NewConfig ProcessConfig `json:"new_config"`
	Changed   []string      `json:"changed"`
}
